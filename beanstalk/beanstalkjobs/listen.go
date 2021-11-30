package beanstalkjobs

import (
	"context"
	"sync/atomic"

	"github.com/beanstalkd/go-beanstalk"
)

func (c *consumer) listen() {
	for {
		select {
		case <-c.stopCh:
			c.log.Debug("beanstalk listener stopped")
			return
		default:
			id, body, err := c.pool.Reserve(c.reserveTimeout)
			if err != nil {
				if errB, ok := err.(beanstalk.ConnError); ok {
					switch errB.Err { //nolint:gocritic
					case beanstalk.ErrTimeout:
						c.log.Info("beanstalk reserve timeout", "warn", errB.Op)
						continue
					}
				}
				// in case of other error - continue
				c.log.Warn("beanstalk reserve", "error", err)
				continue
			}

			// todo(rustatian): to sync pool
			item := &Item{}
			err = c.unpack(id, body, item)
			if err != nil {
				c.log.Error("beanstalk unpack item", "error", err)
				_ = c.pool.Delete(context.Background(), id)
				continue
			}

			atomic.StoreUint64(&c.items, 1)
			// insert job into the priority queue
			c.pq.Insert(item)
		}
	}
}
