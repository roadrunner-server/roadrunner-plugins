package v2_6

import (
	"time"
)

type BeanstalkConfig struct {
	PipePriority   int64         `mapstructure:"priority"`
	TubePriority   *uint32       `mapstructure:"tube_priority"`
	Tube           string        `mapstructure:"tube"`
	ReserveTimeout time.Duration `mapstructure:"reserve_timeout"`
}

type BoltDBConfig struct {
	File     string `mapstructure:"file"`
	Prefetch int    `mapstructure:"prefetch"`
}

type AMQPConfig struct {
	Prefetch      int    `mapstructure:"prefetch"`
	Queue         string `mapstructure:"queue"`
	Exchange      string `mapstructure:"exchange"`
	ExchangeType  string `mapstructure:"exchange_type"`
	RoutingKey    string `mapstructure:"routing_key"`
	Exclusive     bool   `mapstructure:"exclusive"`
	MultipleAck   bool   `mapstructure:"multiple_ask"`
	RequeueOnFail bool   `mapstructure:"requeue_on_fail"`
}

type NATSConfig struct {
	Subject            string `mapstructure:"subject"`
	Stream             string `mapstructure:"stream"`
	Prefetch           int    `mapstructure:"prefetch"`
	RateLimit          uint64 `mapstructure:"rate_limit"`
	DeleteAfterAck     bool   `mapstructure:"delete_after_ack"`
	DeliverNew         bool   `mapstructure:"deliver_new"`
	DeleteStreamOnStop bool   `mapstructure:"delete_stream_on_stop"`
}

type SQSConfig struct {
	VisibilityTimeout int32             `mapstructure:"visibility_timeout"`
	WaitTimeSeconds   int32             `mapstructure:"wait_time_seconds"`
	Prefetch          int32             `mapstructure:"prefetch"`
	Queue             *string           `mapstructure:"queue"`
	Attributes        map[string]string `mapstructure:"attributes"`
	Tags              map[string]string `mapstructure:"tags"`
}
