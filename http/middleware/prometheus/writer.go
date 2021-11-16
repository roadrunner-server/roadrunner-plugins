package prometheus

import (
	"net/http"
)

type writer struct {
	w    http.ResponseWriter
	code int
}

func (w *writer) WriteHeader(code int) {
	w.code = code
	w.w.WriteHeader(code)
}

func (w *writer) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *writer) Header() http.Header {
	return w.w.Header()
}
