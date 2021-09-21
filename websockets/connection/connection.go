package connection

import (
	"io"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/logger"
)

const EOF string = "EOF"

// Connection represents wrapped and safe to use from the different threads websocket connection
type Connection struct {
	sync.RWMutex
	log  logger.Logger
	conn net.Conn
}

func NewConnection(wsConn net.Conn, log logger.Logger) *Connection {
	return &Connection{
		conn: wsConn,
		log:  log,
	}
}

func (c *Connection) Write(data []byte) error {
	c.Lock()
	defer c.Unlock()

	const op = errors.Op("websocket_write")
	// handle a case when a goroutine tried to write into the closed connection
	defer func() {
		if r := recover(); r != nil {
			c.log.Warn("panic handled, tried to write into the closed connection")
		}
	}()

	header := ws.Header{
		Fin:    true,
		OpCode: ws.OpBinary,
		Masked: false,
		Length: int64(len(data)),
	}

	err := ws.WriteHeader(c.conn, header)
	if err != nil {
		return errors.E(op, err)
	}
	_, err = c.conn.Write(data)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (c *Connection) Read() ([]byte, ws.OpCode, error) {
	const op = errors.Op("websocket_read")

	header, err := ws.ReadHeader(c.conn)
	if err != nil {
		if err.Error() == EOF {
			return nil, ws.OpClose, nil
		}

		return nil, ws.OpContinuation, errors.E(op, err)
	}

	if header.OpCode == ws.OpClose {
		return nil, ws.OpClose, nil
	}

	payload := make([]byte, header.Length)
	_, err = io.ReadFull(c.conn, payload)
	if err != nil {
		return nil, ws.OpContinuation, errors.E(op, err)
	}

	if header.Masked {
		ws.Cipher(payload, header.Mask, 0)
	}

	return payload, header.OpCode, nil
}

func (c *Connection) Close() error {
	c.Lock()
	defer c.Unlock()
	const op = errors.Op("websocket_close")

	err := c.conn.Close()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}
