package newrelic

import (
	"net/http"
)

type writer struct {
	w         http.ResponseWriter
	code      int
	hdrToSend map[string][]string
}

func (w writer) WriteHeader(code int) {
	w.w.WriteHeader(code)
}

func (w writer) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w writer) Header() http.Header {
	return w.hdrToSend
}
