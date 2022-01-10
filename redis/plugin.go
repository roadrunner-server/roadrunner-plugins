package redis

import (
	"sync"

	"github.com/roadrunner-server/api/v2/plugins/config"
	"github.com/roadrunner-server/api/v2/plugins/kv"
	"github.com/roadrunner-server/api/v2/plugins/pubsub"
	"github.com/spiral/errors"
	redis_kv "github.com/spiral/roadrunner-plugins/v2/redis/kv"
	redis_pubsub "github.com/spiral/roadrunner-plugins/v2/redis/pubsub"
	"go.uber.org/zap"
)

const PluginName = "redis"

type Plugin struct {
	sync.RWMutex
	// config for RR integration
	cfgPlugin config.Configurer
	// logger
	log *zap.Logger
}

func (p *Plugin) Init(cfg config.Configurer, log *zap.Logger) error {
	p.log = new(zap.Logger)
	*p.log = *log
	p.cfgPlugin = cfg

	return nil
}

func (p *Plugin) Name() string {
	return PluginName
}

// KvFromConfig provides KV storage implementation over the redis plugin
func (p *Plugin) KvFromConfig(key string) (kv.Storage, error) {
	const op = errors.Op("redis_plugin_provide")
	st, err := redis_kv.NewRedisDriver(p.log, key, p.cfgPlugin)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return st, nil
}

func (p *Plugin) PubSubFromConfig(key string) (pubsub.PubSub, error) {
	const op = errors.Op("pub_sub_from_config")
	ps, err := redis_pubsub.NewPubSubDriver(p.log, key, p.cfgPlugin)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return ps, nil
}
