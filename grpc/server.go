package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/grpc/parser"
	"github.com/spiral/roadrunner-plugins/v2/grpc/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

func (p *Plugin) createGRPCserver() (*grpc.Server, error) {
	const op = errors.Op("grpc_plugin_create_server")
	opts, err := p.serverOptions()
	if err != nil {
		return nil, errors.E(op, err)
	}

	server := grpc.NewServer(opts...)

	for i := 0; i < len(p.config.Proto); i++ {
		if p.config.Proto[i] == "" {
			continue
		}

		// php proxy services
		services, err := parser.File(p.config.Proto[i], path.Dir(p.config.Proto[i]))
		if err != nil {
			return nil, err
		}

		for _, service := range services {
			px := proxy.NewProxy(fmt.Sprintf("%s.%s", service.Package, service.Name), p.config.Proto[i], p.gPool, p.mu)
			for _, m := range service.Methods {
				px.RegisterMethod(m.Name)
			}

			server.RegisterService(px.ServiceDesc(), px)
			p.proxyList = append(p.proxyList, px)
		}
	}

	// external and native  services
	for _, r := range p.services {
		r(server)
	}

	return server, nil
}

func (p *Plugin) interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	if err != nil {
		p.log.Error("method call finished with error", "error", err, "method", info.FullMethod, "start", start, "elapsed", time.Since(start))

		return nil, err
	}

	p.log.Debug("method called successfully", "method", info.FullMethod, "start", start, "elapsed", time.Since(start))
	return resp, nil
}

func (p *Plugin) serverOptions() ([]grpc.ServerOption, error) {
	const op = errors.Op("grpc_plugin_server_options")

	var tcreds credentials.TransportCredentials
	var opts []grpc.ServerOption
	var cert tls.Certificate
	var certPool *x509.CertPool
	var rca []byte
	var err error

	if p.config.EnableTLS() {
		// if client CA is not empty we combine it with Cert and Key
		if p.config.TLS.RootCA != "" {
			cert, err = tls.LoadX509KeyPair(p.config.TLS.Cert, p.config.TLS.Key)
			if err != nil {
				return nil, err
			}

			certPool, err = x509.SystemCertPool()
			if err != nil {
				return nil, err
			}
			if certPool == nil {
				certPool = x509.NewCertPool()
			}

			rca, err = os.ReadFile(p.config.TLS.RootCA)
			if err != nil {
				return nil, err
			}

			if ok := certPool.AppendCertsFromPEM(rca); !ok {
				return nil, errors.E(op, errors.Str("could not append Certs from PEM"))
			}

			tcreds = credentials.NewTLS(&tls.Config{
				MinVersion:   tls.VersionTLS12,
				ClientAuth:   p.config.TLS.auth,
				Certificates: []tls.Certificate{cert},
				ClientCAs:    certPool,
			})
		} else {
			tcreds, err = credentials.NewServerTLSFromFile(p.config.TLS.Cert, p.config.TLS.Key)
			if err != nil {
				return nil, err
			}
		}

		serverOptions := []grpc.ServerOption{
			grpc.MaxSendMsgSize(int(p.config.MaxSendMsgSize)),
			grpc.MaxRecvMsgSize(int(p.config.MaxRecvMsgSize)),
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle:     p.config.MaxConnectionIdle,
				MaxConnectionAge:      p.config.MaxConnectionAge,
				MaxConnectionAgeGrace: p.config.MaxConnectionAge,
				Time:                  p.config.PingTime,
				Timeout:               p.config.Timeout,
			}),
			grpc.MaxConcurrentStreams(uint32(p.config.MaxConcurrentStreams)),
		}

		opts = append(opts, grpc.Creds(tcreds))
		opts = append(opts, serverOptions...)
	}

	opts = append(opts, p.opts...)

	// custom codec is required to bypass protobuf, common interceptor used for debug and stats
	return append(
		opts,
		grpc.UnaryInterceptor(p.interceptor),
	), nil
}
