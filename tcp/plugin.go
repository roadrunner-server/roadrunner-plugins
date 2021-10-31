package tcp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/google/uuid"
	rrErrors "github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner/v2/payload"
	"github.com/spiral/roadrunner/v2/pool"
	"github.com/spiral/roadrunner/v2/utils"
)

const (
	pluginName string = "tcp"
	RrMode     string = "RR_MODE"
)

var (
	bufferSize int = 1024 * 1024 * 1
)

/*
	rsp.Context -> `WRITECLOSE` -> write back and close
	---
	rsp.Context -> `CLOSE` -> close connection w/o writing to it
	---
	rsp.Context -> WRITE -> conn.Write -> loop
	---
	rsp.Context -> CONTINUE -> w/o write loop
*/

var (
	CLOSE      = []byte("CLOSE")
	CONTINUE   = []byte("CONTINUE")
	WRITECLOSE = []byte("WRITECLOSE")
	WRITE      = []byte("WRITE")
)

const (
	EventConnected    string = "CONNECTED"
	EventIncomingData string = "DATA"
	EventClose        string = "CLOSE"
)

type ServerInfo struct {
	RemoteAddr string `json:"remote_addr"`
	Server     string `json:"server"`
	UUID       string `json:"uuid"`
	Event      string `json:"event"`
}

type Plugin struct {
	sync.RWMutex
	cfg         *Config
	log         logger.Logger
	server      server.Server
	connections sync.Map // uuid -> conn

	wPool pool.Pool

	resBufPool   sync.Pool
	readBufPool  sync.Pool
	servInfoPool sync.Pool
}

func (p *Plugin) Init(log logger.Logger, cfg config.Configurer, server server.Server) error {
	const op = rrErrors.Op("tcp_plugin_init")

	if !cfg.Has(pluginName) {
		return rrErrors.E(op, rrErrors.Disabled)
	}

	err := cfg.UnmarshalKey(pluginName, &p.cfg)
	if err != nil {
		return rrErrors.E(op, err)
	}

	err = p.cfg.InitDefault()
	if err != nil {
		return err
	}

	// buffer sent to the user
	p.resBufPool = sync.Pool{
		New: func() interface{} {
			buf := new(bytes.Buffer)
			buf.Grow(bufferSize)
			return buf
		},
	}

	// cyclic buffer to readLoop the data from the connection
	p.readBufPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, bufferSize)
			return &buf
		},
	}

	p.servInfoPool = sync.Pool{
		New: func() interface{} {
			return new(ServerInfo)
		},
	}

	p.log = log
	p.server = server
	bufferSize = p.cfg.ReadBufferSize
	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	var err error
	p.wPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{RrMode: pluginName})
	if err != nil {
		errCh <- err
		return errCh
	}

	for k := range p.cfg.Servers {
		go func(addr string, delim []byte, name string) {
			// create a TCP listener
			l, err := utils.CreateListener(addr)
			if err != nil {
				errCh <- err
				return
			}

			for {
				conn, err := l.Accept()
				if err != nil {
					p.log.Warn("listener error, stopping", "error", err)
					// just stop
					return
				}

				go p.handleConnection(conn, delim, name)
			}
		}(p.cfg.Servers[k].Addr, p.cfg.Servers[k].delimBytes, k)
	}

	return errCh
}

func (p *Plugin) Stop() error {
	// close all connections
	p.connections.Range(func(_, value interface{}) bool {
		conn := value.(net.Conn)
		if conn != nil {
			_ = conn.Close()
		}
		return true
	})
	return nil
}

func (p *Plugin) Reset() error {
	p.Lock()
	defer p.Unlock()

	var err error
	p.wPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{RrMode: pluginName})
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) Close(uuid string) error {
	if c, ok := p.connections.LoadAndDelete(uuid); ok {
		conn := c.(net.Conn)
		if conn != nil {
			return conn.Close()
		}
	}

	return nil
}

func (p *Plugin) RPC() interface{} {
	return &rpc{
		p: p,
	}
}

func (p *Plugin) handleConnection(conn net.Conn, delim []byte, serverName string) {
	// generate id for the connection
	id := uuid.NewString()
	// store connection to close from outside
	p.connections.Store(id, conn)
	defer p.connections.Delete(id)

	pldCtxConnected, err := p.generate(EventConnected, serverName, id, conn.RemoteAddr().String())
	if err != nil {
		p.log.Error("payload marshaling error", "error", err)
		return
	}

	pld := &payload.Payload{
		Context: pldCtxConnected,
	}

	// send connected
	p.RLock()
	rsp, err := p.wPool.Exec(pld)
	p.RUnlock()
	if err != nil {
		p.log.Error("execute error", "error", err)
		_ = conn.Close()
		return
	}

	// handleAndContinue return true if the RR needs to return from the loop, or false to continue
	if p.handleAndContinue(rsp, conn, serverName, id) {
		p.readLoop(conn, delim, id, serverName)
	}
}

func (p *Plugin) readLoop(conn net.Conn, delim []byte, id, serverName string) {
	rbuf := p.getReadBuf()
	defer p.putReadBuf(rbuf)
	resbuf := p.getResBuf()
	defer p.putResBuf(resbuf)

	pldCtxData, err := p.generate(EventIncomingData, serverName, id, conn.RemoteAddr().String())
	if err != nil {
		p.log.Error("generate payload error", "error", err)
		return
	}

	// start readLoop loop
	for {
		// readLoop a data from the connection
		for {
			n, errR := conn.Read(*rbuf)
			if errR != nil {
				if errors.Is(errR, io.EOF) {
					p.sendClose(serverName, id, conn.RemoteAddr().String())
					break
				}
				p.log.Warn("readLoop error, connection closed", "error", errR)
				_ = conn.Close()

				p.sendClose(serverName, id, conn.RemoteAddr().String())
				return
			}

			if n < len(delim) {
				p.log.Error("received small payload from the connection. less than delimiter")
				_ = conn.Close()

				p.sendClose(serverName, id, conn.RemoteAddr().String())
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
			if bytes.Equal((*rbuf)[:n][n-len(delim):], delim) {
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

		pld := &payload.Payload{
			Context: pldCtxData,
			Body:    resbuf.Bytes(),
		}

		// reset protection
		p.RLock()
		rsp, err := p.wPool.Exec(pld)
		p.RUnlock()
		if err != nil {
			p.log.Error("execute error", "error", err)
			_ = conn.Close()
			return
		}

		// handleAndContinue return true if the RR needs to return from the loop, or false to continue
		if p.handleAndContinue(rsp, conn, serverName, id) {
			// reset the readLoop-buffer
			resbuf.Reset()
			continue
		}
		return
	}
}
