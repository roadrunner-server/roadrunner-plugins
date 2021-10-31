package handler

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner/v2/payload"
)

type handler struct {
	conn        net.Conn
	connections *sync.Map
	serverName  string
	delim       []byte
	uuid        string
	wPool       func(*payload.Payload) (*payload.Payload, error)
	log         logger.Logger

	servInfoPool *sync.Pool
	readBufPool  *sync.Pool
	resBufPool   *sync.Pool
	pldPool      *sync.Pool
}

func NewHandler(conn net.Conn, delim []byte, serverName string, wPool func(*payload.Payload) (*payload.Payload, error),
	pldPool *sync.Pool, siPool *sync.Pool, readBufPool *sync.Pool, resBufPool *sync.Pool, connections *sync.Map, log logger.Logger) *handler {
	return &handler{
		conn:         conn,
		connections:  connections,
		serverName:   serverName,
		uuid:         uuid.NewString(),
		delim:        delim,
		wPool:        wPool,
		pldPool:      pldPool,
		servInfoPool: siPool,
		readBufPool:  readBufPool,
		resBufPool:   resBufPool,
		log:          log,
	}
}

func (h *handler) Start() {
	// generate id for the connection
	id := uuid.NewString()
	// store connection to close from outside
	h.connections.Store(id, h.conn)
	defer h.connections.Delete(id)

	pldCtxConnected, err := h.generate(EventConnected)
	if err != nil {
		h.log.Error("payload marshaling error", "error", err)
		return
	}

	pld := h.getPayload()
	pld.Context = pldCtxConnected

	// send connected
	rsp, err := h.wPool(pld)
	if err != nil {
		h.log.Error("execute error", "error", err)
		_ = h.conn.Close()
		h.putPayload(pld)
		return
	}

	h.putPayload(pld)

	// handleAndContinue return true if the RR needs to return from the loop, or false to continue
	if h.handleAndContinue(rsp) {
		h.readLoop()
	}
}

func (h *handler) Release() {
	// noop at the moment
}

func (h *handler) readLoop() {
	rbuf := h.getReadBuf()
	resbuf := h.getResBuf()
	defer h.putReadBuf(rbuf)
	defer h.putResBuf(resbuf)

	pldCtxData, err := h.generate(EventIncomingData)
	if err != nil {
		h.log.Error("generate payload error", "error", err)
		return
	}

	// start read loop
	for {
		// read a data from the connection
		for {
			n, errR := h.conn.Read(*rbuf)
			if errR != nil {
				if errors.Is(errR, io.EOF) {
					h.sendClose()
					break
				}
				h.log.Warn("read error, connection closed", "error", errR)
				_ = h.conn.Close()

				h.sendClose()
				return
			}

			if n < len(h.delim) {
				h.log.Error("received small payload from the connection. less than delimiter")
				_ = h.conn.Close()

				h.sendClose()
				return
			}

			/*
				n -> aaaaaaaa
				total -> aaaaaaaa -> \n\r
			*/
			// BCE ??
			/*
				check delimiter algo:
				check the ending of the payload
			*/
			if bytes.Equal((*rbuf)[:n][n-len(h.delim):], h.delim) {
				// write w/o delimiter
				resbuf.Write((*rbuf)[:n])
				break
			}

			resbuf.Write((*rbuf)[:n])
		}

		// connection closed
		if resbuf.Len() == 0 {
			return
		}

		pld := h.getPayload()
		pld.Context = pldCtxData
		pld.Body = resbuf.Bytes()

		// reset protection
		rsp, err := h.wPool(pld)
		if err != nil {
			h.log.Error("execute error", "error", err)
			_ = h.conn.Close()
			h.putPayload(pld)
			return
		}

		// handleAndContinue return true if the RR needs to return from the loop, or false to continue
		if h.handleAndContinue(rsp) {
			// reset the read-buffer
			resbuf.Reset()
			h.putPayload(pld)
			continue
		}

		h.putPayload(pld)
		return
	}
}

func (h *handler) handleAndContinue(rsp *payload.Payload) bool {
	switch {
	case bytes.Equal(rsp.Context, CONTINUE):
		// cont
		return true
	case bytes.Equal(rsp.Context, WRITE):
		_, err := h.conn.Write(rsp.Body)
		if err != nil {
			h.log.Error("write response error", "error", err)
			_ = h.conn.Close()
			h.sendClose()
			// stop
			return false
		}

		// cont
		return true
	case bytes.Equal(rsp.Context, WRITECLOSE):
		_, err := h.conn.Write(rsp.Body)
		if err != nil {
			h.log.Error("write response error", "error", err)
			_ = h.conn.Close()
			h.sendClose()
			// stop
			return false
		}

		err = h.conn.Close()
		if err != nil {
			h.log.Error("close connection error", "error", err)
		}

		h.sendClose()
		// stop
		return false

	case bytes.Equal(rsp.Context, CLOSE):
		err := h.conn.Close()
		if err != nil {
			h.log.Error("close connection error", "error", err)
		}

		h.sendClose()
		// stop
		return false

	default:
		err := h.conn.Close()
		if err != nil {
			h.log.Error("close connection error", "error", err)
		}

		h.sendClose()
		// stop
		return false
	}
}
