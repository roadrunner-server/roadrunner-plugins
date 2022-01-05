package pool

import (
	"sync"

	json "github.com/goccy/go-json"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/pubsub"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/websockets/connection"
	"github.com/spiral/roadrunner/v2/utils"
	"go.uber.org/zap"
)

// number of the WS pollers
const workersNum int = 10

type WorkersPool struct {
	subscriber  pubsub.Subscriber
	connections *sync.Map
	resPool     sync.Pool
	log         *zap.Logger

	queue chan *pubsub.Message
	exit  chan struct{}
}

// NewWorkersPool constructs worker pool for the websocket connections
func NewWorkersPool(subscriber pubsub.Subscriber, connections *sync.Map, log *zap.Logger) *WorkersPool {
	wp := &WorkersPool{
		connections: connections,
		queue:       make(chan *pubsub.Message, 100),
		subscriber:  subscriber,
		log:         log,
		exit:        make(chan struct{}),
	}

	wp.resPool.New = func() interface{} {
		return make(map[string]struct{}, 10)
	}

	// start 10 workers
	for i := 0; i < workersNum; i++ {
		wp.do()
	}

	return wp
}

func (wp *WorkersPool) Queue(msg *pubsub.Message) {
	wp.queue <- msg
}

func (wp *WorkersPool) Stop() {
	for i := 0; i < workersNum; i++ {
		wp.exit <- struct{}{}
	}

	close(wp.exit)
}

func (wp *WorkersPool) put(res map[string]struct{}) {
	// optimized
	// https://go-review.googlesource.com/c/go/+/110055/
	// not O(n), but O(1)
	for k := range res {
		delete(res, k)
	}
}

func (wp *WorkersPool) get() map[string]struct{} {
	return wp.resPool.Get().(map[string]struct{})
}

// Response from the server
type Response struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

func (wp *WorkersPool) do() { //nolint:gocognit
	go func() {
		for {
			select {
			case msg, ok := <-wp.queue:
				if !ok {
					return
				}
				if msg == nil || msg.Topic == "" {
					continue
				}

				// get free map
				res := wp.get()

				// get connections for the particular topic
				wp.subscriber.Connections(msg.Topic, res)

				if len(res) == 0 {
					wp.log.Info("no connections associated with provided topic", zap.String("topic", msg.Topic))
					wp.put(res)
					continue
				}

				// res is a map with a connectionsID
				for connID := range res {
					c, ok := wp.connections.Load(connID)
					if !ok {
						wp.log.Warn("websocket was disconnected before the message being written to it", zap.String("topics", msg.Topic))
						wp.put(res)
						continue
					}

					d, err := json.Marshal(&Response{
						Topic:   msg.Topic,
						Payload: utils.AsString(msg.Payload),
					})

					if err != nil {
						wp.log.Error("error marshaling response", zap.Error(err))
						wp.put(res)
						break
					}

					// put data into the bytes buffer
					err = c.(*connection.Connection).Write(d)
					if err != nil {
						wp.log.Error("error sending payload over the connection", zap.String("topic", msg.Topic), zap.Error(err))
						wp.put(res)
						continue
					}
				}
			case <-wp.exit:
				wp.log.Info("get exit signal, exiting from the workers pool")
				return
			}
		}
	}()
}
