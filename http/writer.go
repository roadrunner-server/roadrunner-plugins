package http

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

func (p *Plugin) getWriter(w http.ResponseWriter) *writer {
	wr := p.writersPool.Get().(*writer)
	wr.w = w
	return wr
}

func (p *Plugin) putWriter(w *writer) {
	w.code = -1
	w.w = nil
	p.writersPool.Put(w)
}
