package natsjobs

import (
	json "github.com/json-iterator/go"
)

// blocking
func (c *consumer) listenerInit() error {
	var err error
	c.sub, err = c.conn.ChanSubscribe(c.natsQ, c.msgCh)
	if err != nil {
		return err
	}

	return nil
}

func (c *consumer) listenerStart() {
	for {
		select {
		case m := <-c.msgCh:
			item := new(Item)

			err := json.Unmarshal(m.Data, item)
			if err != nil {
				c.log.Error("unmarshal nats payload", "error", err)
				continue
			}

			// save the ack, nak and requeue functions
			item.Options.ack = m.Ack
			item.Options.nak = m.Nak
			item.Options.requeueFn = c.requeue

			c.queue.Insert(item)
		case <-c.stopCh:
			return
		}
	}
}
