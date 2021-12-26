package amqpjobs

import (
	"time"

	"github.com/cenkalti/backoff/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/jobs/pipeline"
	"go.uber.org/zap"
)

// redialer used to redial to the rabbitmq in case of the connection interrupts
func (c *consumer) redialer() { //nolint:gocognit
	go func() {
		const op = errors.Op("rabbitmq_redial")

		for {
			select {
			case err := <-c.conn.NotifyClose(make(chan *amqp.Error)):
				if err == nil {
					return
				}

				c.Lock()

				// trash the broken publishing channel
				<-c.publishChan

				t := time.Now().UTC()
				pipe := c.pipeline.Load().(*pipeline.Pipeline)

				c.log.Error("pipeline connection was closed, redialing", zap.Error(err), zap.String("pipeline", pipe.Name()), zap.String("driver", pipe.Driver()), zap.Time("start", t))

				expb := backoff.NewExponentialBackOff()
				// set the retry timeout (minutes)
				expb.MaxElapsedTime = c.retryTimeout
				operation := func() error {
					c.log.Warn("reconnecting", zap.Error(err))
					var dialErr error
					c.conn, dialErr = amqp.Dial(c.connStr)
					if dialErr != nil {
						return errors.E(op, dialErr)
					}

					c.log.Info("rabbitmq dial was succeed. trying to redeclare queues and subscribers")

					// re-init connection
					errInit := c.initRabbitMQ()
					if errInit != nil {
						c.log.Error("rabbitmq dial", zap.Error(errInit))
						return errInit
					}

					// redeclare consume channel
					var errConnCh error
					c.consumeChan, errConnCh = c.conn.Channel()
					if errConnCh != nil {
						return errors.E(op, errConnCh)
					}

					// redeclare publish channel
					pch, errPubCh := c.conn.Channel()
					if errPubCh != nil {
						return errors.E(op, errPubCh)
					}

					// start reading messages from the channel
					deliv, err := c.consumeChan.Consume(
						c.queue,
						c.consumeID,
						false,
						false,
						false,
						false,
						nil,
					)
					if err != nil {
						return errors.E(op, err)
					}

					// put the fresh publishing channel
					c.publishChan <- pch
					// restart listener
					c.listener(deliv)

					c.log.Info("queues and subscribers was redeclared successfully")

					return nil
				}

				retryErr := backoff.Retry(operation, expb)
				if retryErr != nil {
					c.Unlock()
					c.log.Error("backoff operation failed", zap.Error(retryErr))
					return
				}

				c.log.Info("connection was successfully restored", zap.String("pipeline", pipe.Name()), zap.String("driver", pipe.Driver()), zap.Time("start", t), zap.Duration("elapsed", time.Since(t)))
				c.Unlock()

			case <-c.stopCh:
				pch := <-c.publishChan
				err := pch.Close()
				if err != nil {
					c.log.Error("publish channel close", zap.Error(err))
				}

				if c.consumeChan != nil {
					err = c.consumeChan.Close()
					if err != nil {
						c.log.Error("consume channel close", zap.Error(err))
					}
				}

				err = c.conn.Close()
				if err != nil {
					c.log.Error("amqp connection close", zap.Error(err))
				}

				return
			}
		}
	}()
}
