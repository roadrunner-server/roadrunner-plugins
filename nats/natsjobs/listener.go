package natsjobs

import (
	json "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// blocking
func (c *consumer) listenerInit() error {
	var err error

	opts := make([]nats.SubOpt, 0)
	if c.deliverNew {
		opts = append(opts, nats.DeliverNew())
	}

	opts = append(opts, nats.RateLimit(c.rateLimit))
	opts = append(opts, nats.AckExplicit())
	c.sub, err = c.js.ChanSubscribe(c.subject, c.msgCh, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (c *consumer) listenerStart() {
	for {
		select {
		case m := <-c.msgCh:
			// only JS messages
			meta, err := m.Metadata()
			if err != nil {
				c.log.Info("can't get message metadata", zap.Error(err))
				continue
			}

			err = m.InProgress()
			if err != nil {
				c.log.Error("failed to send InProgress state", zap.Error(err))
				continue
			}

			item := new(Item)
			err = json.Unmarshal(m.Data, item)
			if err != nil {
				c.log.Error("unmarshal nats payload", zap.Error(err))
				continue
			}

			// save the ack, nak and requeue functions
			item.Options.ack = m.Ack
			item.Options.nak = m.Nak
			item.Options.requeueFn = c.requeue
			item.Options.respondFn = c.respond
			// sequence needed for the requeue
			item.Options.seq = meta.Sequence.Stream

			// needed only if delete after ack is true
			if c.deleteAfterAck {
				item.Options.stream = c.stream
				item.Options.sub = c.js
				item.Options.deleteAfterAck = c.deleteAfterAck
			}

			if item.Priority() == 0 {
				item.Options.Priority = c.priority
			}

			c.queue.Insert(item)
		case <-c.stopCh:
			return
		}
	}
}
