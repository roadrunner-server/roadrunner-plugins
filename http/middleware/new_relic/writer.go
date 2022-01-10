package newrelic

import (
	"net/http"
)

type writer struct {
	code      int
	data      []byte
	hdrToSend map[string][]string
}

func (w *writer) WriteHeader(code int) {
	w.code = code
}

func (w *writer) Write(b []byte) (int, error) {
	w.data = make([]byte, len(b))
	copy(w.data, b)
	return len(w.data), nil
}

func (w *writer) Header() http.Header {
	return w.hdrToSend
}
