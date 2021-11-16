package handler

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/http/attributes"
	"github.com/spiral/roadrunner-plugins/v2/http/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner/v2/events"
	"github.com/spiral/roadrunner/v2/payload"
	"github.com/spiral/roadrunner/v2/pool"
)

const (
	// MB is 1024 bytes
	MB         uint64 = 1024 * 1024
	ContentLen string = "Content-Length"
)

// Handler serves http connections to underlying PHP application using PSR-7 protocol. Context will include request headers,
// parsed files and query, payload will include parsed form dataTree (if any).
type Handler struct {
	maxRequestSize uint64
	uploads        *config.Uploads
	trusted        config.Cidrs
	log            logger.Logger
	pool           pool.Pool

	events   events.EventBus
	eventsID string

	accessLogs       bool
	internalHTTPCode uint64

	// internal
	reqPool  sync.Pool
	respPool sync.Pool
	pldPool  sync.Pool
}

// NewHandler return handle interface implementation
func NewHandler(maxReqSize uint64, internalHTTPCode uint64, uploads *config.Uploads, trusted config.Cidrs, pool pool.Pool, log logger.Logger, accessLogs bool) (*Handler, error) {
	if pool == nil {
		return nil, errors.E(errors.Str("pool should be initialized"))
	}

	eb, id := events.Bus()

	return &Handler{
		maxRequestSize:   maxReqSize * MB,
		uploads:          uploads,
		pool:             pool,
		log:              log,
		trusted:          trusted,
		events:           eb,
		eventsID:         id,
		internalHTTPCode: internalHTTPCode,
		accessLogs:       accessLogs,
		reqPool: sync.Pool{
			New: func() interface{} {
				return &Request{
					Attributes: make(map[string]interface{}),
					Cookies:    make(map[string]string),
					body:       nil,
				}
			},
		},
		respPool: sync.Pool{
			New: func() interface{} {
				return &Response{
					Headers: make(map[string][]string),
					Status:  -1,
				}
			},
		},
		pldPool: sync.Pool{
			New: func() interface{} {
				return &payload.Payload{
					Body:    make([]byte, 0, 100),
					Context: make([]byte, 0, 100),
				}
			},
		},
	}, nil
}

// ServeHTTP transform original request to the PSR-7 passed then to the underlying application. Attempts to serve static files first if enabled.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = errors.Op("serve_http")
	start := time.Now()

	// validating request size
	if h.maxRequestSize != 0 {
		const op = errors.Op("http_handler_max_size")
		if length := r.Header.Get(ContentLen); length != "" {
			// try to parse the value from the `content-length` header
			size, err := strconv.ParseInt(length, 10, 64)
			if err != nil {
				// if got an error while parsing -> assign 500 code to the writer and return
				http.Error(w, "", 500)
				h.log.Error("error while parsing value from the content-length header", "start", start, "elapsed", time.Since(start))
				return
			}

			if size > int64(h.maxRequestSize) {
				h.log.Error("request body max size is exceeded", "allowed_size", h.maxRequestSize, "actual_size", size, "start", start, "elapsed", time.Since(start))
				http.Error(w, errors.E(op, errors.Str("request body max size is exceeded")).Error(), http.StatusBadRequest)
				return
			}
		}
	}

	req := h.getReq(r)

	err := request(r, req, h.uploads)
	if err != nil {
		// if pipe is broken, there is no sense to write the header
		// in this case we just report about error
		if err == errEPIPE {
			h.putReq(req)
			h.log.Error("write response error", "error", err, "start", start, "elapsed", time.Since(start))
			return
		}

		h.putReq(req)
		http.Error(w, errors.E(op, err).Error(), 500)
		h.log.Error("request forming error", "error", err, "start", start, "elapsed", time.Since(start))
		return
	}

	// proxy IP resolution
	h.resolveIP(req)
	req.Open(h.log)
	// get payload from the pool
	pld := h.getPld()

	err = req.Payload(pld)
	if err != nil {
		req.Close(h.log)
		h.putReq(req)
		h.putPld(pld)
		h.handleError(w, start, err)
		h.log.Error("payload forming error", "error", err, "start", start, "elapsed", time.Since(start))
		return
	}

	wResp, err := h.pool.Exec(pld)
	if err != nil {
		req.Close(h.log)
		h.putReq(req)
		h.putPld(pld)
		h.handleError(w, start, err)
		h.log.Error("execute error", "error", err, "start", start, "elapsed", time.Since(start))
		return
	}

	status, err := h.Write(wResp, w)
	if err != nil {
		req.Close(h.log)
		h.putReq(req)
		h.putPld(pld)
		h.handleError(w, start, err)
		h.log.Error("write response error", "error", err, "start", start, "elapsed", time.Since(start))
		return
	}

	if !h.accessLogs {
		h.log.Info("http request processed", "status", status, "method", req.Method, "URI", req.URI, "remote_address", req.RemoteAddr, "start", start, "elapsed", time.Since(start))
	} else {
		body, _ := json.Marshal(r.Header)
		reqLen := len(body) + int(r.ContentLength)

		h.log.Info("http access log", "status", status, "method", req.Method, "URI", req.URI, "remote_address", req.RemoteAddr, "query", req.RawQuery, "request_length", reqLen, "bytes_sent", r.ContentLength, "host", r.Host, "user_agent", r.UserAgent(), "referer", r.Referer(), "time_local", time.Now().Format("02/Jan/06:15:04:05 -0700"), "request_time", time.Now(), "start", start, "elapsed", time.Since(start))
	}

	h.putPld(pld)
	req.Close(h.log)
	h.putReq(req)
}

func (h *Handler) Dispose() {
	h.events.Unsubscribe(h.eventsID)
}

// handleError will handle internal RR errors and return 500
func (h *Handler) handleError(w http.ResponseWriter, start time.Time, err error) {
	// internal error types, user should not see them
	if errors.Is(errors.SoftJob, err) ||
		errors.Is(errors.WatcherStopped, err) ||
		errors.Is(errors.WorkerAllocate, err) ||
		errors.Is(errors.NoFreeWorkers, err) ||
		errors.Is(errors.ExecTTL, err) ||
		errors.Is(errors.IdleTTL, err) ||
		errors.Is(errors.TTL, err) ||
		errors.Is(errors.Encode, err) ||
		errors.Is(errors.Decode, err) ||
		errors.Is(errors.Network, err) {
		// write an internal server error
		w.WriteHeader(int(h.internalHTTPCode))
	}
}

// get real ip passing multiple proxy
func (h *Handler) resolveIP(r *Request) {
	if h.trusted.IsTrusted(r.RemoteAddr) == false { //nolint:gosimple
		return
	}

	if r.Header.Get("X-Forwarded-For") != "" {
		ips := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		ipCount := len(ips)

		for i := ipCount - 1; i >= 0; i-- {
			addr := strings.TrimSpace(ips[i])
			if net.ParseIP(addr) != nil {
				r.RemoteAddr = addr
				return
			}
		}

		return
	}

	// The logic here is the following:
	// In general case, we only expect X-Real-Ip header. If it exist, we get the IP address from header and set request Remote address
	// But, if there is no X-Real-Ip header, we also trying to check CloudFlare headers
	// True-Client-IP is a general CF header in which copied information from X-Real-Ip in CF.
	// CF-Connecting-IP is an Enterprise feature and we check it last in order.
	// This operations are near O(1) because Headers struct are the map type -> type MIMEHeader map[string][]string
	if r.Header.Get("X-Real-Ip") != "" {
		r.RemoteAddr = FetchIP(r.Header.Get("X-Real-Ip"))
		return
	}

	if r.Header.Get("True-Client-IP") != "" {
		r.RemoteAddr = FetchIP(r.Header.Get("True-Client-IP"))
		return
	}

	if r.Header.Get("CF-Connecting-IP") != "" {
		r.RemoteAddr = FetchIP(r.Header.Get("CF-Connecting-IP"))
	}
}

func (h *Handler) putReq(req *Request) {
	h.reqPool.Put(req)
}

func (h *Handler) getReq(r *http.Request) *Request {
	req := h.reqPool.Get().(*Request)
	req.RemoteAddr = FetchIP(r.RemoteAddr)
	req.Protocol = r.Proto
	req.Method = r.Method
	req.URI = URI(r)
	req.Header = r.Header
	req.Cookies = make(map[string]string)
	req.RawQuery = r.URL.RawQuery
	req.Attributes = attributes.All(r)

	req.Parsed = false
	req.body = nil
	return req
}

func (h *Handler) putRsp(rsp *Response) {
	rsp.Headers = nil
	rsp.Status = -1
	h.respPool.Put(rsp)
}

func (h *Handler) getRsp() *Response {
	return h.respPool.Get().(*Response)
}

func (h *Handler) putPld(pld *payload.Payload) {
	pld.Body = nil
	pld.Context = nil
	h.pldPool.Put(pld)
}

func (h *Handler) getPld() *payload.Payload {
	return h.pldPool.Get().(*payload.Payload)
}
