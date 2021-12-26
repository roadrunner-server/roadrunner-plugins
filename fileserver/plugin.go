package fileserver

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/config"
	"github.com/spiral/roadrunner/v2/utils"
	"go.uber.org/zap"
)

const (
	pluginName string = "fileserver"
)

type Plugin struct {
	sync.Mutex
	config *Config

	log *zap.Logger
	app *fiber.App
}

func (p *Plugin) Init(cfg config.Configurer, log *zap.Logger) error {
	const op = errors.Op("file_server_init")

	if !cfg.Has(pluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(pluginName, &p.config)
	if err != nil {
		return errors.E(op, err)
	}

	p.log = log

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	p.Lock()
	p.app = fiber.New(fiber.Config{
		Prefork:                      false,
		BodyLimit:                    0,
		ReadTimeout:                  time.Second * 5,
		WriteTimeout:                 0,
		IdleTimeout:                  0,
		DisableKeepalive:             false,
		DisableDefaultDate:           false,
		DisableDefaultContentType:    false,
		DisableHeaderNormalizing:     false,
		DisableStartupMessage:        true,
		StreamRequestBody:            p.config.StreamRequestBody,
		DisablePreParseMultipartForm: false,
		ReduceMemoryUsage:            false,
	})

	if p.config.CalculateEtag {
		p.app.Use(etag.New(etag.Config{
			Weak: p.config.Weak,
		}))
	}

	for i := 0; i < len(p.config.Configuration); i++ {
		p.app.Static(p.config.Configuration[i].Prefix, p.config.Configuration[i].Root, fiber.Static{
			Compress:      p.config.Configuration[i].Compress,
			ByteRange:     p.config.Configuration[i].BytesRange,
			Browse:        false,
			CacheDuration: time.Second * time.Duration(p.config.Configuration[i].CacheDuration),
			MaxAge:        p.config.Configuration[i].MaxAge,
		})
	}

	ln, err := utils.CreateListener(p.config.Address)
	if err != nil {
		errCh <- err
		return errCh
	}

	go func() {
		p.Unlock()
		p.log.Info("file server started", zap.String("address", p.config.Address))
		err = p.app.Listener(ln)
		if err != nil {
			errCh <- err
			return
		}
	}()

	return errCh
}

func (p *Plugin) Stop() error {
	p.Lock()
	defer p.Unlock()

	err := p.app.Shutdown()
	if err != nil {
		return err
	}
	return nil
}

func (p *Plugin) Name() string {
	return pluginName
}
