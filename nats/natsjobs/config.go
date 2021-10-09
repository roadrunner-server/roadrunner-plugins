package natsjobs

import (
	"github.com/nats-io/nats.go"
)

type config struct {
	// global
	// NATS URL
	Addr string `mapstructure:"addr"`

	Queue    string `mapstructure:"queue"`
	Prefetch int    `mapstructure:"prefetch"`
}

func (c *config) InitDefaults() {
	if c.Addr == "" {
		c.Addr = nats.DefaultURL
	}

	if c.Queue == "" {
		c.Queue = "default"
	}

	if c.Prefetch == 0 {
		c.Prefetch = 10
	}
}
