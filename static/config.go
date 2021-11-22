package static

import (
	"time"

	"github.com/spiral/errors"
)

type Config struct {
	// Address to serve
	Address string `mapstructure:"address"`

	// CalculateEtag can be true/false and used to calculate etag for the static
	CalculateEtag bool `mapstructure:"calculate_etag"`

	// Weak etag `W/`
	Weak bool `mapstructure:"weak"`

	// per-root configuration
	Configuration []*Cfg `mapstructure:"serve"`

	// StreamRequestBody ...
	StreamRequestBody bool `mapstructure:"stream_request_body"`
}

type Cfg struct {
	// Prefix HTTP
	Prefix string `mapstructure:"prefix"`

	// Dir contains name of directory to control access to.
	// Default - "."
	Root string `mapstructure:"root"`

	BytesRange    bool          `mapstructure:"bytes_range"`
	Compress      bool          `mapstructure:"compress"`
	CacheDuration time.Duration `mapstructure:"cache_duration"`
	MaxAge        int           `mapstructure:"max_age"`
}

func (c *Config) Valid() error {
	const op = errors.Op("static_validation")
	if c.Address == "" {
		return errors.E(op, errors.Str("empty address"))
	}

	return nil
}
