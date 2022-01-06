package directives

import (
	"strconv"
	"strings"

	"github.com/spiral/roadrunner/v2/utils"
	"go.uber.org/zap"
)

/*
Response Cache-Control Directives: https://datatracker.ietf.org/doc/html/rfc7234#section-5.2.2
*/

/*
   Cache-Control   = 1#cache-directive
   cache-directive = token [ "=" ( token / quoted-string ) ]
*/

const (
	maxAge       string = "max-age"
	maxStale     string = "max-stale"
	minFresh     string = "min-fresh"
	noCache      string = "no-cache"
	noStore      string = "no-store"
	noTransform  string = "no-transform"
	onlyIfCached string = "only-if-cached"
)

// Req represents possible Cache-Control request header values
type Req struct {
	MaxAge   *uint64
	MaxStale *uint64
	MinFresh *uint64

	NoCache      bool
	NoStore      bool
	NoTransform  bool
	OnlyIfCached bool
}

func (r *Req) Reset() {
	r.MaxAge = nil
	r.MaxStale = nil
	r.MinFresh = nil

	r.NoCache = false
	r.NoStore = false
	r.NoTransform = false
	r.OnlyIfCached = false
}

func ParseRequest(directives string, log *zap.Logger, r *Req) {
	split := strings.Split(directives, ",")

	// a lot of allocations here - 14 todo(rustatian): FIXME
	for i := 0; i < len(split); i++ {
		// max-age, max-stale, min-fresh
		if idx := strings.IndexByte(split[i], '='); idx != -1 {
			// get token and associated value
			if len(split[i]) < idx+1 {
				log.Warn("bad header", zap.String("value", split[i]))
				continue
			}

			token := strings.Trim(split[i][:idx], " ")
			val := strings.Trim(split[i][idx+1:], " ")
			if val == "" || token == "" {
				log.Warn("bad header", zap.String("value", split[i]))
				continue
			}

			switch token {
			case maxAge:
				valUint, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					log.Error("parse max-age", zap.String("value", val), zap.Error(err))
					continue
				}
				r.MaxAge = utils.Uint64(valUint)
			case maxStale:
				valUint, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					log.Error("parse max-stale", zap.String("value", val), zap.Error(err))
					continue
				}
				r.MaxStale = utils.Uint64(valUint)
			case minFresh:
				valUint, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					log.Error("parse min-fresh", zap.String("value", val), zap.Error(err))
					continue
				}
				r.MinFresh = utils.Uint64(valUint)
			}

			continue
		}

		token := strings.Trim(split[i], " ")

		switch token {
		case noCache:
			r.NoCache = true
		case noStore:
			r.NoStore = true
		case noTransform:
			r.NoTransform = true
		case onlyIfCached:
			r.OnlyIfCached = true
		default:
			continue
		}
	}
}
