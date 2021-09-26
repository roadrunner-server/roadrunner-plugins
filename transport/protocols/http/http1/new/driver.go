package new

import (
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner/v2/events"
	"github.com/spiral/roadrunner/v2/pool"
)

type driver struct {
	log logger.Logger
	eh events.Handler
	pool pool.Pool
}

func New(cfgKey string, cfg config.Configurer, log logger.Logger, eh events.Handler, pool pool.Pool) *driver {
	return &driver{}
}

// Serve the driver
func (d *driver) Serve() {

}

// Start listening, blocking
func (d *driver) Listen() {

}


