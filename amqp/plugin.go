package amqp

import (
	"github.com/roadrunner-server/api/v2/plugins/config"
	"github.com/roadrunner-server/api/v2/plugins/jobs"
	"github.com/roadrunner-server/api/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner-plugins/v2/amqp/amqpjobs"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
	"go.uber.org/zap"
)

const (
	pluginName string = "amqp"
)

type Plugin struct {
	log *zap.Logger
	cfg config.Configurer
}

func (p *Plugin) Init(log *zap.Logger, cfg config.Configurer) error {
	p.log = new(zap.Logger)
	*p.log = *log
	p.cfg = cfg
	return nil
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) ConsumerFromConfig(configKey string, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return amqpjobs.NewAMQPConsumer(configKey, p.log, p.cfg, pq)
}

// ConsumerFromPipeline constructs AMQP driver from pipeline
func (p *Plugin) ConsumerFromPipeline(pipe *pipeline.Pipeline, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return amqpjobs.FromPipeline(pipe, p.log, p.cfg, pq)
}
