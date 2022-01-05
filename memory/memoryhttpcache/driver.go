package memoryhttpcache

import (
	"sync"

	"github.com/spiral/errors"
	"go.uber.org/zap"
)

type Cache struct {
	log  *zap.Logger
	data sync.Map // map[string][]byte
}

func NewCacheDriver(log *zap.Logger) (*Cache, error) {
	l := new(zap.Logger)
	*l = *log
	return &Cache{
		log: l,
	}, nil
}

func (c *Cache) Get(id uint64) ([]byte, error) {
	c.log.Debug("cache_get", zap.Uint64("id", id))
	data, ok := c.data.Load(id)
	if !ok {
		return nil, errors.E(errors.EmptyItem)
	}

	return data.([]byte), nil
}

func (c *Cache) Set(id uint64, value []byte) error {
	c.log.Debug("cache_set", zap.Uint64("id", id), zap.ByteString("data", value))
	c.data.Store(id, value)
	return nil
}
