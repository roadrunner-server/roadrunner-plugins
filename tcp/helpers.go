package tcp

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner/v2/payload"
)

func (p *Plugin) generate(event, serverName, id, remoteAddr string) ([]byte, error) {
	si := p.getServInfo(event, serverName, id, remoteAddr)
	pld, err := json.Marshal(si)
	if err != nil {
		p.putServInfo(si)
		return nil, err
	}

	p.putServInfo(si)
	return pld, nil
}

func (p *Plugin) sendClose(serverName, id, remoteAddr string) {
	c, err := p.generate(EventClose, serverName, id, remoteAddr)
	if err != nil {
		p.log.Error("payload marshaling error", "error", err)
		return
	}
	pld := &payload.Payload{
		Context: c,
	}
	p.RLock()
	_, _ = p.wPool.Exec(pld)
	p.RUnlock()
}
