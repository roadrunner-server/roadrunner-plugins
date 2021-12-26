package memcached

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/kv"
	"github.com/spiral/roadrunner-plugins/v2/memcached/memcachedkv"
	"go.uber.org/zap"
)

const (
	PluginName     string = "memcached"
	RootPluginName string = "kv"
)

type Plugin struct {
	// config plugin
	cfgPlugin config.Configurer
	// logger
	log *zap.Logger
}

func (p *Plugin) Init(log *zap.Logger, cfg config.Configurer) error {
	if !cfg.Has(RootPluginName) {
		return errors.E(errors.Disabled)
	}

	p.cfgPlugin = cfg
	p.log = log
	return nil
}

// Name returns plugin user-friendly name
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) KvFromConfig(key string) (kv.Storage, error) {
	const op = errors.Op("boltdb_plugin_provide")
	st, err := memcachedkv.NewMemcachedDriver(p.log, key, p.cfgPlugin)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return st, nil
}
