package tcp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"

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

// {
// "name": "server1",
// "uuid": "foo"
// }

const (
	pluginName string = "tcp"
)

type ServerInfo struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

type Plugin struct {
	cfg    *Config
	log    logger.Logger
	server server.Server

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
	p.wPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{})
	if err != nil {
		panic(err)
	}

	go func() {
		for k := range p.cfg.Servers {
			go func(addr string, delim []byte) {
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
						p.handleConnection(conn, delim)
					}()
				}
			}(p.cfg.Servers[k].Addr, p.cfg.Servers[k].delimBytes)
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

func (p *Plugin) handleConnection(conn net.Conn, delim []byte) {
	// generate id for the connection
	id := uuid.NewString()
	// todo(rustatian): TO sync.Pool
	rbuf := make([]byte, 1024)
	resbuf := make([]byte, 0, 1024)

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

		rbuf = rbuf[:n]
		// aaaaaaaa\n\r
		/*
			n -> aaaaaaaa
			total -> aaaaaaaa -> \n\r
		*/
		if bytes.Equal(rbuf[n-len(delim):], delim) {
			resbuf = append(resbuf, rbuf[:n-len(delim)]...)
			break
		}

		resbuf = append(resbuf, rbuf[:n]...)
	}

	// TODO(rustatian): to sync.Pool
	si := &ServerInfo{
		Name: "Server1",
		UUID: id,
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

	_, err = conn.Write(rsp.Body)
	if err != nil {
		p.log.Error("write response error", "error", err)
		return
	}
}
