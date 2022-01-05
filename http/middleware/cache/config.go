package cache

import (
	"net/http"
)

type Config struct {
	CacheMinUses int `mapstructure:"cache_min_uses"`
	// Only get by default
	CacheMethods []string `mapstructure:"cache_methods"`
}

func (c *Config) InitDefaults() {
	// save every request by default
	if c.CacheMinUses == 0 {
		c.CacheMinUses = 1
	}

	// cache only GET requests
	if len(c.CacheMethods) == 0 {
		c.CacheMethods = append(c.CacheMethods, http.MethodGet)
	}
}
