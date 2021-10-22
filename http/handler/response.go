package handler

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/spiral/roadrunner/v2/payload"
	"github.com/spiral/roadrunner/v2/utils"
)

const (
	rrNewRelicKey          string = "rr_newrelic"
	newRelicTransactionKey string = "transaction_name"
)

// Response handles PSR7 response logic.
type Response struct {
	// Status contains response status.
	Status int `json:"status"`

	// Header contains list of response headers.
	Headers map[string][]string `json:"headers"`
}

// Write writes response headers, status and body into ResponseWriter.
func (h *Handler) Write(rq *http.Request, pld *payload.Payload, w http.ResponseWriter) (int, error) {
	rsp := h.getRsp()
	defer h.putRsp(rsp)

	// unmarshal context into response
	err := json.Unmarshal(pld.Context, rsp)
	if err != nil {
		return 0, err
	}

	// handle push headers
	if len(rsp.Headers[http2pushHeaderKey]) != 0 {
		push := rsp.Headers[http2pushHeaderKey]

		if pusher, ok := w.(http.Pusher); ok {
			for i := 0; i < len(push); i++ {
				err = pusher.Push(rsp.Headers[http2pushHeaderKey][i], nil)
				if err != nil {
					return 0, err
				}
			}
		}
	}

	tx := newrelic.FromContext(rq.Context())
	// we have a new relic mdw enabled
	if tx != nil {
		hdr := rsp.Headers[rrNewRelicKey]
		if len(hdr) == 0 {
			// to be sure
			delete(rsp.Headers, rrNewRelicKey)
			goto cont
		}

		for i := 0; i < len(hdr); i++ {
			// 58 according to the ASCII table is -> :
			pos := bytes.IndexByte(utils.AsBytes(hdr[i]), 58)

			// not found
			if pos == -1 {
				continue
			}

			// handle :foo and foo: cases
			if len(hdr[i][:pos]) == 0 || len(hdr[i][pos:]) == 0 {
				continue
			}

			if bytes.Equal(utils.AsBytes(hdr[i][pos:]), utils.AsBytes(newRelicTransactionKey)) {
				tx.SetName(hdr[i][pos:])
				continue
			}

			tx.AddAttribute(hdr[i][pos:], hdr[i][:pos])
		}
	}

	// delete sensitive information
	delete(rsp.Headers, rrNewRelicKey)

cont:

	if len(rsp.Headers[TrailerHeaderKey]) != 0 {
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
	for _, tr := range h[TrailerHeaderKey] {
		for _, n := range strings.Split(tr, ",") {
			n = strings.Trim(n, "\t ")
			if v, ok := h[n]; ok {
				h["Trailer:"+n] = v

				delete(h, n)
			}
		}
	}

	delete(h, TrailerHeaderKey)
}
