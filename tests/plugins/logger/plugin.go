package logger

import (
	"strings"

	"github.com/roadrunner-server/api/plugins/v2/config"
	"github.com/spiral/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Plugin struct {
	config config.Configurer
	log    *zap.Logger
}

type Loggable struct {
}

func (l *Loggable) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("error", "Example marshaller error")
	return nil
}

func (p1 *Plugin) Init(cfg config.Configurer, log *zap.Logger) error {
	p1.config = cfg
	p1.log = log
	return nil
}

func (p1 *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	p1.log.Error("error", zap.Error(errors.E(errors.Str("test"))))
	p1.log.Info("error", zap.Error(errors.E(errors.Str("test"))))
	p1.log.Debug("error", zap.Error(errors.E(errors.Str("test"))))
	p1.log.Warn("error", zap.Error(errors.E(errors.Str("test"))))

	field := zap.String("error", "Example field error")

	p1.log.Error("error", field)
	p1.log.Info("error", field)
	p1.log.Debug("error", field)
	p1.log.Warn("error", field)

	marshalledObject := &Loggable{}

	p1.log.Error("error", zap.Any("object", marshalledObject))
	p1.log.Info("error", zap.Any("object", marshalledObject))
	p1.log.Debug("error", zap.Any("object", marshalledObject))
	p1.log.Warn("error", zap.Any("object", marshalledObject))

	p1.log.Error("error", zap.String("test", ""))
	p1.log.Info("error", zap.String("test", ""))
	p1.log.Debug("error", zap.String("test", ""))
	p1.log.Warn("error", zap.String("test", ""))

	// test the `raw` mode
	messageJSON := []byte(`{"field": "value"}`)
	p1.log.Debug(strings.TrimRight(string(messageJSON), " \n\t"))

	return errCh
}

func (p1 *Plugin) Stop() error {
	return nil
}

func (p1 *Plugin) Name() string {
	return "logger_plugin"
}
