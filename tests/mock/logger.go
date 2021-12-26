package mock_logger //nolint:stylecheck

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLoggerMock struct {
	l *zap.Logger
}

func ZapTestLogger(enab zapcore.LevelEnabler) (*ZapLoggerMock, *ObservedLogs) {
	core, logs := New(enab)
	obsLog := zap.New(core, zap.Development())

	return &ZapLoggerMock{
		l: obsLog,
	}, logs
}

func (z *ZapLoggerMock) Init() error {
	return nil
}

func (z *ZapLoggerMock) Serve() chan error {
	return make(chan error, 1)
}

func (z *ZapLoggerMock) Stop() error {
	return z.l.Sync()
}

func (z *ZapLoggerMock) Provides() []interface{} {
	return []interface{}{
		z.ProvideZapLogger,
	}
}

func (z *ZapLoggerMock) ProvideZapLogger() *zap.Logger {
	return z.l
}
