package nats

import (
	"github.com/spiral/roadrunner-plugins/v2/api/jobs"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/jobs/pipeline"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/nats/natsjobs"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
)

const (
	pluginName string = "nats"
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

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) ConsumerFromConfig(configKey string, queue priorityqueue.Queue) (jobs.Consumer, error) {
	return natsjobs.FromConfig(configKey, p.log, p.cfg, queue)
}

func (p *Plugin) ConsumerFromPipeline(pipe *pipeline.Pipeline, queue priorityqueue.Queue) (jobs.Consumer, error) {
	return natsjobs.FromPipeline(pipe, p.log, p.cfg, queue)
}
