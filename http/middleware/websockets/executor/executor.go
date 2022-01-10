package executor

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"github.com/goccy/go-json"
	"github.com/roadrunner-server/api/v2/plugins/pubsub"
	websocketsv1 "github.com/roadrunner-server/api/v2/proto/websockets/v1beta"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/websockets/commands"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/websockets/connection"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/websockets/validator"
	"go.uber.org/zap"
)

const (
	joinRq string = "@join"
	joinRs string = "#join"
	leave  string = "@leave"
)

type Response struct {
	Topic   string   `json:"topic"`
	Payload []string `json:"payload"`
}

type Executor struct {
	sync.Mutex
	// raw ws connection
	conn *connection.Connection
	log  *zap.Logger

	// associated connection ID
	connID string

	// subscriber drivers
	sub          pubsub.Subscriber
	actualTopics map[string]struct{}

	req             *http.Request
	accessValidator validator.AccessValidatorFn
}

// NewExecutor creates protected connection and starts command loop
func NewExecutor(conn *connection.Connection, log *zap.Logger,
	connID string, sub pubsub.Subscriber, av validator.AccessValidatorFn, r *http.Request) *Executor {
	return &Executor{
		conn:            conn,
		connID:          connID,
		log:             log,
		sub:             sub,
		accessValidator: av,
		actualTopics:    make(map[string]struct{}, 10),
		req:             r,
	}
}

func (e *Executor) StartCommandLoop() error { //nolint:gocognit
	const op = errors.Op("executor_command_loop")
	for {
		data, opCode, err := e.conn.Read()
		if err != nil {
			return errors.E(op, err)
		}

		if opCode == ws.OpClose {
			e.log.Debug("socket was closed", zap.Error(err))
			return nil
		}

		msg := &websocketsv1.Message{}

		err = json.Unmarshal(data, msg)
		if err != nil {
			e.log.Error("unmarshal message", zap.Error(err))
			continue
		}

		// nil message, continue
		if msg == nil {
			e.log.Debug("nil message, skipping")
			continue
		}

		switch msg.Command {
		// handle leave
		case commands.Join:
			e.log.Debug("join command is received", zap.Any("msg", msg))

			val, err := e.accessValidator(e.req, msg.Topics...)
			if err != nil {
				if val != nil {
					e.log.Debug("validation error", zap.Int("status", val.Status), zap.Any("headers", val.Header), zap.ByteString("body", val.Body))
				}

				resp := &Response{
					Topic:   joinRs,
					Payload: msg.Topics,
				}

				packet, errJ := json.Marshal(resp)
				if errJ != nil {
					e.log.Error("marshal the body", zap.Error(errJ))
					return errors.E(op, fmt.Errorf("%v,%v", err, errJ))
				}

				errW := e.conn.Write(packet)
				if errW != nil {
					e.log.Error("write payload to the connection", zap.ByteString("payload", packet), zap.Error(errW))
					return errors.E(op, fmt.Errorf("%v,%v", err, errW))
				}

				continue
			}

			resp := &Response{
				Topic:   joinRq,
				Payload: msg.Topics,
			}

			packet, errM := json.Marshal(resp)
			if errM != nil {
				e.log.Error("marshal the body", zap.Error(errM))
				return errors.E(op, err)
			}

			err = e.conn.Write(packet)
			if err != nil {
				e.log.Error("write payload to the connection", zap.ByteString("payload", packet), zap.Error(err))
				return errors.E(op, err)
			}

			// subscribe to the topic
			err = e.Set(msg.Topics)
			if err != nil {
				return errors.E(op, err)
			}

		// handle leave
		case commands.Leave:
			e.log.Debug("received leave command", zap.Any("msg", msg))

			// prepare response
			resp := &Response{
				Topic:   leave,
				Payload: msg.Topics,
			}

			packet, err := json.Marshal(resp)
			if err != nil {
				e.log.Error("marshal the body", zap.Error(err))
				return errors.E(op, err)
			}

			err = e.conn.Write(packet)
			if err != nil {
				e.log.Error("write payload to the connection", zap.ByteString("payload", packet), zap.Error(err))
				return errors.E(op, err)
			}

			err = e.Leave(msg.Topics)
			if err != nil {
				return errors.E(op, err)
			}

		case commands.Headers:

		default:
			e.log.Warn("unknown command", zap.String("command", msg.Command))
		}
	}
}

func (e *Executor) Set(topics []string) error {
	// associate connection with topics
	err := e.sub.Subscribe(e.connID, topics...)
	if err != nil {
		e.log.Error("subscribe to the provided topics", zap.Strings("topics", topics), zap.Error(err))
		// in case of error, unsubscribe connection from the dead topics
		_ = e.sub.Unsubscribe(e.connID, topics...)
		return err
	}

	// save topics for the connection
	for i := 0; i < len(topics); i++ {
		e.actualTopics[topics[i]] = struct{}{}
	}

	return nil
}

func (e *Executor) Leave(topics []string) error {
	// remove associated connections from the storage
	err := e.sub.Unsubscribe(e.connID, topics...)
	if err != nil {
		e.log.Error("subscribe to the provided topics", zap.Strings("topics", topics), zap.Error(err))
		return err
	}

	// remove topics for the connection
	for i := 0; i < len(topics); i++ {
		delete(e.actualTopics, topics[i])
	}

	return nil
}

func (e *Executor) CleanUp() {
	// unsubscribe particular connection from the topics
	for topic := range e.actualTopics {
		_ = e.sub.Unsubscribe(e.connID, topic)
	}

	// clean up the actualTopics data
	for k := range e.actualTopics {
		delete(e.actualTopics, k)
	}
}
