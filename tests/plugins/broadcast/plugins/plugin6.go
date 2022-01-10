package plugins

import (
	"context"
	"fmt"

	"github.com/roadrunner-server/api/v2/plugins/pubsub"
	"github.com/spiral/errors"
	"go.uber.org/zap"
)

const Plugin6Name = "plugin6"

type Plugin6 struct {
	log    *zap.Logger
	b      pubsub.Broadcaster
	driver pubsub.SubReader
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *Plugin6) Init(log *zap.Logger, b pubsub.Broadcaster) error {
	p.log = new(zap.Logger)
	*p.log = *log
	p.b = b
	p.ctx, p.cancel = context.WithCancel(context.Background())
	return nil
}

func (p *Plugin6) Serve() chan error {
	errCh := make(chan error, 1)

	var err error
	p.driver, err = p.b.GetDriver("test")
	if err != nil {
		panic(err)
	}

	err = p.driver.Subscribe("6", "foo")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			msg, errD := p.driver.Next(p.ctx)
			if errD != nil {
				if errors.Is(errors.TimeOut, errD) {
					return
				}
				errCh <- errD
				return
			}

			if msg == nil {
				continue
			}

			p.log.Info(fmt.Sprintf("%s: %s", Plugin6Name, *msg))
		}
	}()

	return errCh
}

func (p *Plugin6) Stop() error {
	_ = p.driver.Unsubscribe("6", "foo")
	p.cancel()
	return nil
}

func (p *Plugin6) Name() string {
	return Plugin6Name
}
