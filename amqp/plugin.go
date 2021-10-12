package amqp

import (
	"github.com/spiral/roadrunner-plugins/v2/amqp/amqpjobs"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/internal/common/jobs"
	"github.com/spiral/roadrunner-plugins/v2/jobs/pipeline"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner/v2/events"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
)

const (
	pluginName string = "amqp"
)

type Plugin struct {
	log logger.Logger
	cfg config.Configurer
}

func (p *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	p.log = log
	p.cfg = cfg
	return nil
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) Available() {}

func (p *Plugin) ConsumerFromConfig(configKey string, e events.Handler, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return amqpjobs.NewAMQPConsumer(configKey, p.log, p.cfg, e, pq)
}

// ConsumerFromPipeline constructs AMQP driver from pipeline
func (p *Plugin) ConsumerFromPipeline(pipe *pipeline.Pipeline, e events.Handler, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return amqpjobs.FromPipeline(pipe, p.log, p.cfg, e, pq)
}
