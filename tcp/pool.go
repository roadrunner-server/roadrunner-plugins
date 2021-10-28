package tcp

import (
	"bytes"
)

func (p *Plugin) getServInfo(event, serverName, id, remoteAddr string) *ServerInfo {
	si := p.servInfoPool.Get().(*ServerInfo)
	si.Event = event
	si.Server = serverName
	si.UUID = id
	si.RemoteAddr = remoteAddr
	return si
}

func (p *Plugin) putServInfo(si *ServerInfo) {
	si.Event = ""
	si.RemoteAddr = ""
	si.Server = ""
	si.UUID = ""
	p.servInfoPool.Put(si)
}

func (p *Plugin) getReadBuf() *[]byte {
	return p.readBufPool.Get().(*[]byte)
}

func (p *Plugin) putReadBuf(buf *[]byte) {
	p.readBufPool.Put(buf)
}

func (p *Plugin) getResBuf() *bytes.Buffer {
	return p.resBufPool.Get().(*bytes.Buffer)
}

func (p *Plugin) putResBuf(buf *bytes.Buffer) {
	buf.Reset()
	p.resBufPool.Put(buf)
}
