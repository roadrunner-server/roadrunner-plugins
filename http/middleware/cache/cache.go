package cache

import (
	"net/http"
	"time"

	cacheV1beta "github.com/roadrunner-server/api/v2/proto/cache/v1beta"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (p *Plugin) writeCache(wr *writer, id uint64) {
	/*
		First - check the status code, should be only 200, 203, 204, 206, 300, 301, 404, 405, 410, 414, and 501
	*/
	switch wr.Code {
	case http.StatusOK:
		payload := p.getRsp()
		defer p.putRsp(payload)

		payload.Headers = make(map[string]*cacheV1beta.HeaderValue, len(wr.HdrToSend))
		payload.Code = uint64(wr.Code)
		payload.Data = make([]byte, len(wr.Data))
		payload.Timestamp = time.Now().Format(time.RFC3339)
		copy(payload.Data, wr.Data)

		for k := range wr.HdrToSend {
			for i := 0; i < len(wr.HdrToSend[k]); i++ {
				payload.Headers[k].Value = wr.HdrToSend[k]
			}
		}

		data, err := proto.Marshal(payload)
		if err != nil {
			p.log.Error("cache write", zap.Error(err))
			return
		}

		_ = p.cache.Set(id, data)
	case http.StatusNonAuthoritativeInfo:
	case http.StatusNoContent:
	case http.StatusPartialContent:
	case http.StatusMultipleChoices:
	case http.StatusMovedPermanently:
	case http.StatusNotFound:
	case http.StatusMethodNotAllowed:
	case http.StatusGone:
	case http.StatusRequestURITooLong:
	case http.StatusNotImplemented:
	}
}
