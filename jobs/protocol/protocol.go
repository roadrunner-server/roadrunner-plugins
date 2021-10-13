package protocol

import (
	"sync"

	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	pq "github.com/spiral/roadrunner/v2/priority_queue"
)

type Type uint32

const (
	NoError Type = iota
	Error
	Response
)

// internal worker protocol (jobs mode)
type protocol struct {
	// message type, see Type
	T Type `json:"type"`
	// Payload
	Data json.RawMessage `json:"data"`
}

type RespHandler struct {
	log logger.Logger
	// response pools
	qPool sync.Pool
	ePool sync.Pool
	pPool sync.Pool
}

func NewResponseHandler(log logger.Logger) *RespHandler {
	return &RespHandler{
		log: log,

		pPool: sync.Pool{
			New: func() interface{} {
				return new(protocol)
			},
		},

		qPool: sync.Pool{
			New: func() interface{} {
				return new(queueResp)
			},
		},

		ePool: sync.Pool{
			New: func() interface{} {
				return new(errorResp)
			},
		},
	}
}

func (rh *RespHandler) Handle(resp []byte, jb pq.Item) error {
	const op = errors.Op("jobs_handle_response")
	p := rh.getProtocol()
	defer rh.putProtocol(p)

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
		return nil
		// error returned from the PHP
	case Error:
		err = rh.handleErrResp(p.Data, jb)
		if err != nil {
			return errors.E(op, err)
		}
		return nil
		// RR should send a response to the queue/tube/subject
	case Response:
		qs := rh.getQResp()
		defer rh.putQResp(qs)
	default:
		err = jb.Ack()
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

func (rh *RespHandler) getProtocol() *protocol {
	return rh.pPool.Get().(*protocol)
}

func (rh *RespHandler) putProtocol(p *protocol) {
	p.T = 0
	p.Data = nil
	rh.pPool.Put(p)
}

func (rh *RespHandler) getErrResp() *errorResp {
	return rh.ePool.Get().(*errorResp)
}

func (rh *RespHandler) putErrResp(p *errorResp) {
	p.Msg = ""
	p.Headers = nil
	p.Delay = 0
	p.Requeue = false
	rh.ePool.Put(p)
}

func (rh *RespHandler) getQResp() *queueResp {
	return rh.qPool.Get().(*queueResp)
}

func (rh *RespHandler) putQResp(p *queueResp) {
	p.Queue = ""
	rh.qPool.Put(p)
}
