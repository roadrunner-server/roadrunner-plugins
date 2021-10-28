package tcp

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pool"
	"github.com/spiral/roadrunner/v2/utils"
)

type Server struct {
	Addr       string `mapstructure:"addr"`
	Delimiter  string `mapstructure:"delimiter"`
	delimBytes []byte
}

type Config struct {
	Servers map[string]*Server `mapstructure:"servers"`
	Pool    *pool.Config       `mapstructure:"pool"`
}

func (c *Config) InitDefault() error {
	if len(c.Servers) == 0 {
		return errors.Str("no servers registered")
	}

	for k, v := range c.Servers {
		if v.Delimiter == "" {
			v.Delimiter = "\r\n"
			v.delimBytes = []byte{'\r', '\n'}
		}

		if v.Addr == "" {
			return errors.Errorf("empty address for the server: %s", k)
		}

		// already written
		if len(v.delimBytes) > 0 {
			continue
		}

		v.delimBytes = utils.AsBytes(v.Delimiter)
	}

	if c.Pool == nil {
		c.Pool = &pool.Config{}
	}

	c.Pool.InitDefaults()

	return nil
}
