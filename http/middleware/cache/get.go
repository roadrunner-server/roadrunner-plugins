package cache

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/cache/directives"
	"github.com/spiral/roadrunner/v2/utils"
	"google.golang.org/protobuf/proto"
)

func (p *Plugin) handleGET(w http.ResponseWriter, r *http.Request, next http.Handler, rq *directives.Req) {
	h := p.getHash()
	defer p.putHash(h)

	wr := p.getWriter()
	defer p.putWriter(wr)

	// write the data to the hash function
	_, err := h.Write(utils.AsBytes(r.RequestURI))
	if err != nil {
		http.Error(w, "failed to write the hash", http.StatusInternalServerError)
		return
	}

	// try to get the data from cache
	out, err := p.cache.Get(h.Sum64())
	if err != nil {
		// cache miss, no data
		if errors.Is(errors.EmptyItem, err) {
			// forward the request to the worker
			next.ServeHTTP(wr, r)

			// ------- RESPONSE STARTS HERE ----------

			// send original data to the receiver
			for k := range wr.HdrToSend {
				for kk := range wr.HdrToSend[k] {
					w.Header().Add(k, wr.HdrToSend[k][kk])
				}
			}

			// write the original status code
			w.WriteHeader(wr.Code)
			// write the data
			_, _ = w.Write(wr.Data)

			// handle the response (decide to cache or not)
			p.writeCache(wr, h.Sum64())
			return
		}

		http.Error(w, "get hash", http.StatusInternalServerError)
		return
	}

	msg := p.getRsp()
	defer p.putRsp(msg)

	err = proto.Unmarshal(out, msg)
	if err != nil {
		http.Error(w, "cache data unpack", http.StatusInternalServerError)
		return
	}

	ts := msg.GetTimestamp()
	parsed, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		http.Error(w, "timestamp parse", http.StatusInternalServerError)
		return
	}

	ageHdr := time.Since(parsed).Seconds()
	if rq.MaxAge != nil {
		// request should not be accepted
		if uint64(ageHdr) > *rq.MaxAge {
			// delete prev data from the cache
			p.cache.Delete(h.Sum64())

			next.ServeHTTP(wr, r)

			for k := range wr.HdrToSend {
				for kk := range wr.HdrToSend[k] {
					w.Header().Add(k, wr.HdrToSend[k][kk])
				}
			}

			// write the original status code
			w.WriteHeader(wr.Code)
			// write the data
			_, _ = w.Write(wr.Data)

			p.writeCache(wr, h.Sum64())
			return
		}
	}

	// send original data
	for k := range msg.Headers {
		for i := 0; i < len(msg.Headers[k].Value); i++ {
			w.Header().Add(k, msg.Headers[k].Value[i])
		}
	}

	// write Age header
	w.Header().Add(age, fmt.Sprintf("%.0f", ageHdr))

	w.WriteHeader(int(msg.Code))
	_, _ = w.Write(msg.Data)
}
