package amqpjobs

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func (c *consumer) listener(deliv <-chan amqp.Delivery) {
	go func() {
		for { //nolint:gosimple
			select {
			case msg, ok := <-deliv:
				if !ok {
					c.log.Debug("delivery channel was closed, leaving the rabbit listener")
					return
				}

				d, err := c.fromDelivery(msg)
				if err != nil {
					c.log.Error("delivery convert", zap.Error(err))
					/*
						Acknowledge failed job to prevent endless loo;
					*/
					err = msg.Ack(false)
					if err != nil {
						c.log.Error("nack failed", zap.Error(err))
					}

					if d != nil {
						d.Headers = nil
						d.Options = nil
					}
					continue
				}

				// insert job into the main priority queue
				c.pq.Insert(d)
			}
		}
	}()
}
