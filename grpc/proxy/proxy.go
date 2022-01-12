package proxy

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/spiral/roadrunner-plugins/v2/grpc/codec"
	"github.com/spiral/roadrunner/v2/payload"
	"github.com/spiral/roadrunner/v2/pool"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	peerAddr     string = ":peer.address"
	peerAuthType string = ":peer.auth-type"
	delimiter    string = "|:|"
)

// base interface for Proxy class
type proxyService interface {
	// RegisterMethod registers new RPC method.
	RegisterMethod(method string)

	// ServiceDesc returns service description for the proxy.
	ServiceDesc() *grpc.ServiceDesc
}

// carry details about service, method and RPC context to PHP process
type rpcContext struct {
	Service string              `json:"service"`
	Method  string              `json:"method"`
	Context map[string][]string `json:"context"`
}

// Proxy manages GRPC/RoadRunner bridge.
type Proxy struct {
	mu       *sync.RWMutex
	grpcPool pool.Pool
	name     string
	metadata string
	methods  []string

	pldPool sync.Pool
}

// NewProxy creates new service proxy object.
func NewProxy(name string, metadata string, grpcPool pool.Pool, mu *sync.RWMutex) *Proxy {
	return &Proxy{
		mu:       mu,
		grpcPool: grpcPool,
		name:     name,
		metadata: metadata,
		methods:  make([]string, 0),
		pldPool: sync.Pool{
			New: func() interface{} {
				return &payload.Payload{
					Codec:   frame.CODEC_JSON,
					Context: make([]byte, 0, 100),
					Body:    make([]byte, 0, 100),
				}
			},
		},
	}
}

// RegisterMethod registers new RPC method.
func (p *Proxy) RegisterMethod(method string) {
	p.methods = append(p.methods, method)
}

// ServiceDesc returns service description for the proxy.
func (p *Proxy) ServiceDesc() *grpc.ServiceDesc {
	desc := &grpc.ServiceDesc{
		ServiceName: p.name,
		Metadata:    p.metadata,
		HandlerType: (*proxyService)(nil),
		Methods:     []grpc.MethodDesc{},
		Streams:     []grpc.StreamDesc{},
	}

	// Registering methods
	for _, m := range p.methods {
		desc.Methods = append(desc.Methods, grpc.MethodDesc{
			MethodName: m,
			Handler:    p.methodHandler(m),
		})
	}

	return desc
}

// Generate method handler proxy.
func (p *Proxy) methodHandler(method string) func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
		in := codec.RawMessage{}
		if err := dec(&in); err != nil {
			return nil, wrapError(err)
		}

		if interceptor == nil {
			return p.invoke(ctx, method, in)
		}

		info := &grpc.UnaryServerInfo{
			Server:     srv,
			FullMethod: fmt.Sprintf("/%s/%s", p.name, method),
		}

		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return p.invoke(ctx, method, req.(codec.RawMessage))
		}

		return interceptor(ctx, in, info, handler)
	}
}

func (p *Proxy) invoke(ctx context.Context, method string, in codec.RawMessage) (interface{}, error) {
	pld := p.getPld()
	defer p.putPld(pld)

	err := p.makePayload(ctx, method, in, pld)
	if err != nil {
		return nil, err
	}

	p.mu.RLock()
	resp, err := p.grpcPool.Exec(pld)
	p.mu.RUnlock()

	if err != nil {
		return nil, wrapError(err)
	}

	md, err := p.responseMetadata(resp)
	if err != nil {
		return nil, err
	}
	ctx = metadata.NewIncomingContext(ctx, md)
	err = grpc.SetHeader(ctx, md)
	if err != nil {
		return nil, err
	}

	return codec.RawMessage(resp.Body), nil
}

// responseMetadata extracts metadata from roadrunner response Payload.Context and converts it to metadata.MD
func (p *Proxy) responseMetadata(resp *payload.Payload) (metadata.MD, error) {
	var md metadata.MD
	if resp == nil || len(resp.Context) == 0 {
		return md, nil
	}

	var rpcMetadata map[string]string
	err := json.Unmarshal(resp.Context, &rpcMetadata)
	if err != nil {
		return md, err
	}

	if len(rpcMetadata) > 0 {
		md = metadata.New(rpcMetadata)
	}

	return md, nil
}

// makePayload generates RoadRunner compatible payload based on GRPC message.
func (p *Proxy) makePayload(ctx context.Context, method string, body codec.RawMessage, pld *payload.Payload) error {
	ctxMD := make(map[string][]string)

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, v := range md {
			ctxMD[k] = v
		}
	}

	if pr, ok := peer.FromContext(ctx); ok {
		ctxMD[peerAddr] = []string{pr.Addr.String()}
		if pr.AuthInfo != nil {
			ctxMD[peerAuthType] = []string{pr.AuthInfo.AuthType()}
		}
	}

	ctxData, err := json.Marshal(rpcContext{Service: p.name, Method: method, Context: ctxMD})

	if err != nil {
		return err
	}

	pld.Body = body
	pld.Context = ctxData

	return nil
}

func (p *Proxy) putPld(pld *payload.Payload) {
	pld.Body = nil
	pld.Context = nil
	p.pldPool.Put(pld)
}

func (p *Proxy) getPld() *payload.Payload {
	pld := p.pldPool.Get().(*payload.Payload)
	pld.Codec = frame.CODEC_JSON
	return pld
}

// mounts proper error code for the error
func wrapError(err error) error {
	// internal agreement
	if strings.Contains(err.Error(), delimiter) {
		chunks := strings.Split(err.Error(), delimiter)
		code := codes.Internal

		// protect the slice access
		if len(chunks) < 2 {
			return err
		}

		phpCode, errConv := strconv.ParseUint(chunks[0], 10, 32)
		if errConv != nil {
			return err
		}

		if phpCode > 0 && phpCode < math.MaxUint32 {
			code = codes.Code(phpCode)
		}

		st := status.New(code, chunks[1]).Proto()

		for _, detailsMessage := range chunks[2:] {
			anyDetailsMessage := anypb.Any{}
			errP := proto.Unmarshal([]byte(detailsMessage), &anyDetailsMessage)
			if errP == nil {
				st.Details = append(st.Details, &anyDetailsMessage)
			}
		}

		return status.ErrorProto(st)
	}

	return status.Error(codes.Internal, err.Error())
}
