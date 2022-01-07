package logger

import (
	"github.com/roadrunner-server/api/plugins/v2/config"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"go.uber.org/zap"
)

// PluginName declares plugin name.
const PluginName = "logs"

// ZapLogger manages zap logger.
type ZapLogger struct {
	base     *zap.Logger
	cfg      *Config
	channels ChannelConfig
}

// Init logger service.
func (z *ZapLogger) Init(cfg config.Configurer) error {
	const op = errors.Op("config_plugin_init")
	var err error
	// if not configured, configure with default params
	if !cfg.Has(PluginName) {
		z.cfg = &Config{}
		z.cfg.InitDefault()

		z.base, err = z.cfg.BuildLogger()
		if err != nil {
			return errors.E(op, err)
		}

		return nil
	}

	err = cfg.UnmarshalKey(PluginName, &z.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	err = cfg.UnmarshalKey(PluginName, &z.channels)
	if err != nil {
		return errors.E(op, err)
	}

	z.cfg.InitDefault()
	z.base, err = z.cfg.BuildLogger()

	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (z *ZapLogger) Serve() chan error {
	return make(chan error, 1)
}

func (z *ZapLogger) Stop() error {
	_ = z.base.Sync()
	return nil
}

// NamedLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
func (z *ZapLogger) NamedLogger(name string) (*zap.Logger, error) {
	if cfg, ok := z.channels.Channels[name]; ok {
		l, err := cfg.BuildLogger()
		if err != nil {
			return nil, err
		}
		return l.Named(name), nil
	}

	return z.base.Named(name), nil
}

// ServiceLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
func (z *ZapLogger) ServiceLogger(n endure.Named) (*zap.Logger, error) {
	return z.NamedLogger(n.Name())
}

// Provides declares factory methods.
func (z *ZapLogger) Provides() []interface{} {
	return []interface{}{
		z.ServiceLogger,
	}
}

// Name returns user-friendly plugin name
func (z *ZapLogger) Name() string {
	return PluginName
}
