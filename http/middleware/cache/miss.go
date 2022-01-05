package cache

import (
	"net/http"

	cacheV1beta "github.com/spiral/roadrunner-plugins/v2/api/proto/cache/v1beta"
	"google.golang.org/protobuf/proto"
)

func (p *Plugin) handleCacheMiss(wr *writer, id uint64) {
	// cache only 200OK responses
	if wr.Code == http.StatusOK {
		payload := p.getRsp()
		defer p.putRsp(payload)

		payload.Headers = make(map[string]*cacheV1beta.HeaderValue, len(wr.HdrToSend))
		payload.Code = uint64(wr.Code)
		payload.Data = make([]byte, len(wr.Data))
		copy(payload.Data, wr.Data)

		for k := range wr.HdrToSend {
			for i := 0; i < len(wr.HdrToSend[k]); i++ {
				payload.Headers[k].Value = wr.HdrToSend[k]
			}
		}

		data, err := proto.Marshal(payload)
		if err != nil {
			panic(err)
		}

		_ = p.cache.Set(id, data)
	}
}
