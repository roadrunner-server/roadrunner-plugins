package sqs

import (
	"github.com/spiral/roadrunner-plugins/v2/common/jobs"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/jobs/pipeline"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner/v2/events"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
)

const (
	pluginName string = "sqs"
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

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) JobsConstruct(configKey string, e events.Handler, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return NewSQSConsumer(configKey, p.log, p.cfg, e, pq)
}

func (p *Plugin) FromPipeline(pipe *pipeline.Pipeline, e events.Handler, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return FromPipeline(pipe, p.log, p.cfg, e, pq)
}
