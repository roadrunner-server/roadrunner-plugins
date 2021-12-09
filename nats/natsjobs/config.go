package natsjobs

import (
	"github.com/nats-io/nats.go"
)

const (
	pipeSubject            string = "subject"
	pipeStream             string = "stream"
	pipePrefetch           string = "prefetch"
	pipeDeleteAfterAck     string = "delete_after_ack"
	pipeDeliverNew         string = "deliver_new"
	pipeRateLimit          string = "rate_limit"
	pipeDeleteStreamOnStop string = "delete_stream_on_stop"
)

type config struct {
	// global
	// NATS URL
	Addr string `mapstructure:"addr"`

	Priority           int64  `mapstructure:"priority"`
	Subject            string `mapstructure:"subject"`
	Stream             string `mapstructure:"stream"`
	Prefetch           int    `mapstructure:"prefetch"`
	RateLimit          uint64 `mapstructure:"rate_limit"`
	DeleteAfterAck     bool   `mapstructure:"delete_after_ack"`
	DeliverNew         bool   `mapstructure:"deliver_new"`
	DeleteStreamOnStop bool   `mapstructure:"delete_stream_on_stop"`
}

func (c *config) InitDefaults() {
	if c.Addr == "" {
		c.Addr = nats.DefaultURL
	}

	if c.RateLimit == 0 {
		c.RateLimit = 1000
	}

	if c.Priority == 0 {
		c.Priority = 10
	}

	if c.Stream == "" {
		c.Stream = "default-stream"
	}

	if c.Subject == "" {
		c.Subject = "default"
	}

	if c.Prefetch == 0 {
		c.Prefetch = 10
	}
}
