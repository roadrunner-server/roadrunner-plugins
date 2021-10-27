package tcp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/google/uuid"
	json "github.com/json-iterator/go"
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

type ServerInfo struct {
	RemoteAddr string `json:"remote_addr"`
	Server     string `json:"server"`
	UUID       string `json:"uuid"`
}

type Plugin struct {
	sync.RWMutex
	cfg         *Config
	log         logger.Logger
	server      server.Server
	connections sync.Map // uuid -> conn

	wPool pool.Pool
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

	p.log = log
	p.server = server

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	var err error
	p.wPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{RrMode: pluginName})
	if err != nil {
		panic(err)
	}

	go func() {
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

					go func() {
						p.handleConnection(conn, delim, name)
					}()
				}
			}(p.cfg.Servers[k].Addr, p.cfg.Servers[k].delimBytes, k)
		}
	}()

	return errCh
}

func (p *Plugin) Stop() error {
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
	p.connections.Store(id, conn)
	defer p.connections.Delete(id)

	// todo(rustatian): TO sync.Pool
	rbuf := make([]byte, 1024)
	resbuf := make([]byte, 0, 1024)

start:
	for {
		n, err := conn.Read(rbuf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			p.log.Error("read data error", "error", err)
			break
		}

		if n < len(delim) {
			p.log.Error("received small payload from the connection. less than delimiter")
			panic("boom")
		}

		/*
			rbuf -> 1024
			10 bytes
			rbuf = rbuf[:n]
			n >= delim
		*/

		//
		if n == len(delim) {
			if bytes.Equal(rbuf[:n], delim) {
				// end of the data
				break
			}
			resbuf = append(resbuf, rbuf[:n]...)
			continue
		}

		/*
			n -> aaaaaaaa
			total -> aaaaaaaa -> \n\r
		*/
		if bytes.Equal(rbuf[:n][n-len(delim):], delim) {
			resbuf = append(resbuf, rbuf[:n][:n-len(delim)]...)
			break
		}

		resbuf = append(resbuf, rbuf[:n]...)
	}

	// TODO(rustatian): to sync.Pool
	si := &ServerInfo{
		RemoteAddr: conn.RemoteAddr().String(),
		Server:     serverName,
		UUID:       id,
	}

	pldCtx, err := json.Marshal(si)
	if err != nil {
		panic(err)
	}

	pld := &payload.Payload{
		Context: pldCtx,
		Body:    resbuf,
	}

	rsp, err := p.wPool.Exec(pld)
	if err != nil {
		p.log.Error("execute error", "error", err)
		return
	}

	switch {
	case bytes.Equal(rsp.Context, CONTINUE):
		goto start
	case bytes.Equal(rsp.Context, WRITE):
		_, err = conn.Write(rsp.Body)
		if err != nil {
			p.log.Error("write response error", "error", err)
			return
		}
		goto start
	case bytes.Equal(rsp.Context, WRITECLOSE):
		_, err = conn.Write(rsp.Body)
		if err != nil {
			p.log.Error("write response error", "error", err)
			return
		}
		err = conn.Close()
		if err != nil {
			p.log.Error("close connection error", "error", err)
		}
		return
	case bytes.Equal(rsp.Context, CLOSE):
		err = conn.Close()
		if err != nil {
			p.log.Error("close connection error", "error", err)
		}
		return
	}
}
