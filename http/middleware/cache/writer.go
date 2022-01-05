package cache

import (
	"net/http"
)

type writer struct {
	Code      int                 `json:"code"`
	Data      []byte              `json:"data"`
	HdrToSend map[string][]string `json:"headers"`
}

func (w *writer) WriteHeader(code int) {
	w.Code = code
}

func (w *writer) Write(b []byte) (int, error) {
	w.Data = make([]byte, len(b))
	copy(w.Data, b)
	return len(w.Data), nil
}

func (w *writer) Header() http.Header {
	return w.HdrToSend
}
