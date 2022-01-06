package cache

import (
	"hash"

	cacheV1beta "github.com/spiral/roadrunner-plugins/v2/api/proto/cache/v1beta"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/cache/directives"
)

func (p *Plugin) getHash() hash.Hash64 {
	return p.hashPool.Get().(hash.Hash64)
}

func (p *Plugin) putHash(h hash.Hash64) {
	h.Reset()
	p.hashPool.Put(h)
}

func (p *Plugin) getWriter() *writer {
	wr := p.writersPool.Get().(*writer)
	return wr
}

func (p *Plugin) putWriter(w *writer) {
	w.Code = -1
	w.Data = nil

	for k := range w.HdrToSend {
		delete(w.HdrToSend, k)
	}

	p.writersPool.Put(w)
}

func (p *Plugin) getRsp() *cacheV1beta.Response {
	return p.rspPool.Get().(*cacheV1beta.Response)
}

func (p *Plugin) putRsp(r *cacheV1beta.Response) {
	r.Reset()
	p.rspPool.Put(r)
}

func (p *Plugin) getRq() *directives.Req {
	return p.rqPool.Get().(*directives.Req)
}

func (p *Plugin) putRq(r *directives.Req) {
	r.Reset()
	p.rqPool.Put(r)
}
