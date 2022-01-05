package cache

import (
	"go.uber.org/zap"
)

type Cache interface {
	Get(id uint64) ([]byte, error)
	Set(id uint64, value []byte) error
}

type HTTPCacheFromConfig interface {
	FromConfig(log *zap.Logger) (Cache, error)
}
