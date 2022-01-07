package memory

import (
	"github.com/roadrunner-server/api/plugins/v2/cache"
	"github.com/roadrunner-server/api/plugins/v2/config"
	"github.com/roadrunner-server/api/plugins/v2/jobs"
	"github.com/roadrunner-server/api/plugins/v2/jobs/pipeline"
	"github.com/roadrunner-server/api/plugins/v2/kv"
	"github.com/roadrunner-server/api/plugins/v2/pubsub"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/memory/memoryhttpcache"
	"github.com/spiral/roadrunner-plugins/v2/memory/memoryjobs"
	"github.com/spiral/roadrunner-plugins/v2/memory/memorykv"
	"github.com/spiral/roadrunner-plugins/v2/memory/memorypubsub"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
	"go.uber.org/zap"
)

const PluginName string = "memory"

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
	return PluginName
}

// Drivers implementation

func (p *Plugin) FromConfig(log *zap.Logger) (cache.Cache, error) {
	return memoryhttpcache.NewCacheDriver(log)
}

func (p *Plugin) PubSubFromConfig(key string) (pubsub.PubSub, error) {
	return memorypubsub.NewPubSubDriver(p.log, key)
}

func (p *Plugin) KvFromConfig(key string) (kv.Storage, error) {
	const op = errors.Op("memory_plugin_construct")
	st, err := memorykv.NewInMemoryDriver(key, p.log, p.cfg)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return st, nil
}

// ConsumerFromConfig creates new ephemeral consumer from the configuration
func (p *Plugin) ConsumerFromConfig(configKey string, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return memoryjobs.FromConfig(configKey, p.log, p.cfg, pq)
}

// ConsumerFromPipeline creates new ephemeral consumer from the provided pipeline
func (p *Plugin) ConsumerFromPipeline(pipeline *pipeline.Pipeline, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return memoryjobs.FromPipeline(pipeline, p.log, pq)
}
