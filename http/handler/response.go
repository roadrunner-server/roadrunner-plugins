package handler

import (
	"net/http"
	"strings"

	json "github.com/goccy/go-json"
	"github.com/spiral/roadrunner/v2/payload"
)

const (
	Trailer   string = "Trailer"
	Http2Push string = "Http2-Push" //nolint:stylecheck
)

// Response handles PSR7 response logic.
type Response struct {
	// Status contains response status.
	Status int `json:"status"`

	// Header contains list of response headers.
	Headers map[string][]string `json:"headers"`
}

// Write writes response headers, status and body into ResponseWriter.
func (h *Handler) Write(pld *payload.Payload, w http.ResponseWriter) (int, error) {
	rsp := h.getRsp()
	defer h.putRsp(rsp)

	// unmarshal context into response
	err := json.Unmarshal(pld.Context, rsp)
	if err != nil {
		return 0, err
	}

	// handle push headers
	if len(rsp.Headers[Http2Push]) != 0 {
		push := rsp.Headers[Http2Push]

		if pusher, ok := w.(http.Pusher); ok {
			for i := 0; i < len(push); i++ {
				err = pusher.Push(rsp.Headers[Http2Push][i], nil)
				if err != nil {
					return 0, err
				}
			}
		}
	}

	if len(rsp.Headers[Trailer]) != 0 {
		handleTrailers(rsp.Headers)
	}

	// write all headers from the response to the writer
	for k := range rsp.Headers {
		for kk := range rsp.Headers[k] {
			w.Header().Add(k, rsp.Headers[k][kk])
		}
	}

	w.WriteHeader(rsp.Status)
	_, err = w.Write(pld.Body)
	if err != nil {
		return 0, err
	}

	status := rsp.Status
	return status, nil
}

func handleTrailers(h map[string][]string) {
	for _, tr := range h[Trailer] {
		for _, n := range strings.Split(tr, ",") {
			n = strings.Trim(n, "\t ")
			if v, ok := h[n]; ok {
				h["Trailer:"+n] = v

				delete(h, n)
			}
		}
	}

	delete(h, Trailer)
}
