package beanstalk

import (
	cfgPlugin "github.com/roadrunner-server/api/v2/plugins/config"
	"github.com/roadrunner-server/api/v2/plugins/jobs"
	"github.com/roadrunner-server/api/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner-plugins/v2/beanstalk/beanstalkjobs"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
	"go.uber.org/zap"
)

const (
	pluginName string = "beanstalk"
)

type Plugin struct {
	log *zap.Logger
	cfg cfgPlugin.Configurer
}

func (p *Plugin) Init(log *zap.Logger, cfg cfgPlugin.Configurer) error {
	p.log = new(zap.Logger)
	*p.log = *log
	p.cfg = cfg
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error)
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) ConsumerFromConfig(configKey string, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return beanstalkjobs.NewBeanstalkConsumer(configKey, p.log, p.cfg, pq)
}

func (p *Plugin) ConsumerFromPipeline(pipe *pipeline.Pipeline, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return beanstalkjobs.FromPipeline(pipe, p.log, p.cfg, pq)
}
