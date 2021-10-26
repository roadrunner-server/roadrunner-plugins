package tcp

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pool"
	"github.com/spiral/roadrunner/v2/utils"
)

type TCPServer struct {
	Addr       string `mapstructure:"addr"`
	Delimiter  string `mapstructure:"delimiter"`
	delimBytes []byte
}

type Config struct {
	Servers map[string]*TCPServer `mapstructure:"servers"`
	Pool    *pool.Config          `mapstructure:"pool"`
}

func (c *Config) InitDefault() error {
	if len(c.Servers) == 0 {
		return errors.Str("no servers registered")
	}

	for k, v := range c.Servers {
		if v.Delimiter == "" {
			return errors.Errorf("no delimiter for the server: %s", k)
		}

		if v.Addr == "" {
			return errors.Errorf("empty address for the server: %s", k)
		}

		// delimiter bytes used here to compare with the connection data
		v.delimBytes = utils.AsBytes(v.Delimiter)
	}

	if c.Pool == nil {
		c.Pool = &pool.Config{}
	}

	c.Pool.InitDefaults()

	return nil
}
