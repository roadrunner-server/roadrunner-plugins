package boltdb

import (
	"github.com/roadrunner-server/api/plugins/v2/config"
	"github.com/roadrunner-server/api/plugins/v2/jobs"
	"github.com/roadrunner-server/api/plugins/v2/jobs/pipeline"
	"github.com/roadrunner-server/api/plugins/v2/kv"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/boltdb/boltjobs"
	"github.com/spiral/roadrunner-plugins/v2/boltdb/boltkv"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
	"go.uber.org/zap"
)

const (
	PluginName string = "boltdb"
)

// Plugin BoltDB K/V storage.
type Plugin struct {
	cfg config.Configurer
	// logger
	log *zap.Logger
}

func (p *Plugin) Init(log *zap.Logger, cfg config.Configurer) error {
	p.log = new(zap.Logger)
	*p.log = *log
	p.cfg = cfg
	return nil
}

// Name returns plugin name
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) KvFromConfig(key string) (kv.Storage, error) {
	const op = errors.Op("boltdb_plugin_provide")
	st, err := boltkv.NewBoltDBDriver(p.log, key, p.cfg)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return st, nil
}

// JOBS bbolt implementation

func (p *Plugin) ConsumerFromConfig(configKey string, queue priorityqueue.Queue) (jobs.Consumer, error) {
	return boltjobs.NewBoltDBJobs(configKey, p.log, p.cfg, queue)
}

func (p *Plugin) ConsumerFromPipeline(pipe *pipeline.Pipeline, queue priorityqueue.Queue) (jobs.Consumer, error) {
	return boltjobs.FromPipeline(pipe, p.log, p.cfg, queue)
}
