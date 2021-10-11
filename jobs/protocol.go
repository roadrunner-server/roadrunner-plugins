package jobs

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	pq "github.com/spiral/roadrunner/v2/priority_queue"
)

type Type uint32

const (
	NoError Type = iota
	Error
)

// internal worker protocol (jobs mode)
type protocol struct {
	// message type, see Type
	T Type `json:"type"`
	// Payload
	Data json.RawMessage `json:"data"`
}

type errorResp struct {
	Msg     string              `json:"message"`
	Requeue bool                `json:"requeue"`
	Delay   int64               `json:"delay_seconds"`
	Headers map[string][]string `json:"headers"`
}

func handleResponse(resp []byte, jb pq.Item, log logger.Logger) error {
	const op = errors.Op("jobs_handle_response")
	// todo(rustatian) to sync.Pool
	p := &protocol{}

	err := json.Unmarshal(resp, p)
	if err != nil {
		return errors.E(op, err)
	}

	switch p.T {
	// likely case
	case NoError:
		err = jb.Ack()
		if err != nil {
			return errors.E(op, err)
		}
	case Error:
		// todo(rustatian) to sync.Pool
		er := &errorResp{}

		err = json.Unmarshal(p.Data, er)
		if err != nil {
			return errors.E(op, err)
		}

		log.Error("jobs protocol error", "error", er.Msg, "delay", er.Delay, "requeue", er.Requeue)

		// requeue the job
		if er.Requeue {
			err = jb.Requeue(er.Headers, er.Delay)
			if err != nil {
				return errors.E(op, err)
			}
			return nil
		}

		// user don't want to requeue the job - silently ACK and return nil
		errAck := jb.Ack()
		if errAck != nil {
			log.Error("job acknowledge", "message", er.Msg, "error", errAck)
		}

		return nil

	default:
		err = jb.Ack()
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}
