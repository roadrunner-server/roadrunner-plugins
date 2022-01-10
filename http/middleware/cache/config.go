package cache

import (
	"net/http"
)

type Config struct {
	// Driver name
	Driver string `mapstructure:"driver"`

	// Only get by default
	CacheMethods []string `mapstructure:"cache_methods"`

	// Driver specific configuration
	Config interface{} `mapstructure:"config"`
}

func (c *Config) InitDefaults() {
	if c.Driver == "" {
		c.Driver = "memory"
	}

	// cache only GET requests
	if len(c.CacheMethods) == 0 {
		c.CacheMethods = append(c.CacheMethods, http.MethodGet)
	}
}
