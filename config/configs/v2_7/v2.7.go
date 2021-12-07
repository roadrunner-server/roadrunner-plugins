package v2_7 //nolint:stylecheck

import (
	"github.com/spiral/roadrunner/v2/pool"
)

type (
	Env      map[string]string
	Pipeline map[string]interface{}
)

type Config struct {
	// --------------
	Nats *struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"nats"`
	// --------------
	Boltdb *struct {
		Permissions int `mapstructure:"permissions"`
	} `mapstructure:"boltdb"`
	// --------------
	AMQP *struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"amqp"`
	// --------------
	Beanstalk *struct {
		Addr    string `mapstructure:"addr"`
		Timeout string `mapstructure:"timeout"`
	} `mapstructure:"beanstalk"`
	// --------------
	Sqs *struct {
		Key          string `mapstructure:"key"`
		Secret       string `mapstructure:"secret"`
		Region       string `mapstructure:"region"`
		SessionToken string `mapstructure:"session_token"`
		Endpoint     string `mapstructure:"endpoint"`
	} `mapstructure:"sqs"`
	// --------------
	Jobs *struct {
		NumPollers   int                  `mapstructure:"num_pollers"`
		PipelineSize int                  `mapstructure:"pipeline_size"`
		Pool         *pool.Config         `mapstructure:"pool"`
		Pipelines    map[string]*Pipeline `mapstructure:"pipelines"`
		Consume      []string             `mapstructure:"consume"`
	} `mapstructure:"jobs"`
}
