package tcp

import (
	"bytes"
	"net"

	"github.com/spiral/roadrunner/v2/payload"
)

func (p *Plugin) handleAndContinue(rsp *payload.Payload, conn net.Conn, serverName, id string) bool {
	switch {
	case bytes.Equal(rsp.Context, CONTINUE):
		// cont
		return true
	case bytes.Equal(rsp.Context, WRITE):
		_, err := conn.Write(rsp.Body)
		if err != nil {
			p.log.Error("write response error", "error", err)
			_ = conn.Close()
			p.sendClose(serverName, id, conn.RemoteAddr().String())
			// stop
			return false
		}

		// cont
		return true
	case bytes.Equal(rsp.Context, WRITECLOSE):
		_, err := conn.Write(rsp.Body)
		if err != nil {
			p.log.Error("write response error", "error", err)
			_ = conn.Close()
			p.sendClose(serverName, id, conn.RemoteAddr().String())
			// stop
			return false
		}

		err = conn.Close()
		if err != nil {
			p.log.Error("close connection error", "error", err)
		}

		p.sendClose(serverName, id, conn.RemoteAddr().String())
		// stop
		return false

	case bytes.Equal(rsp.Context, CLOSE):
		err := conn.Close()
		if err != nil {
			p.log.Error("close connection error", "error", err)
		}

		p.sendClose(serverName, id, conn.RemoteAddr().String())
		// stop
		return false

	default:
		err := conn.Close()
		if err != nil {
			p.log.Error("close connection error", "error", err)
		}

		p.sendClose(serverName, id, conn.RemoteAddr().String())
		// stop
		return false
	}
}
