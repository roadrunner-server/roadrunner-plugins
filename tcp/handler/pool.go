package handler

import (
	"bytes"

	"github.com/spiral/roadrunner/v2/payload"
)

func (h *handler) getServInfo(event string) *ServerInfo {
	si := h.servInfoPool.Get().(*ServerInfo)
	si.Event = event
	si.Server = h.serverName
	si.UUID = h.uuid
	si.RemoteAddr = h.conn.RemoteAddr().String()
	return si
}

func (h *handler) putServInfo(si *ServerInfo) {
	si.Event = ""
	si.RemoteAddr = ""
	si.Server = ""
	si.UUID = ""
	h.servInfoPool.Put(si)
}

func (h *handler) getReadBuf() *[]byte {
	return h.readBufPool.Get().(*[]byte)
}

func (h *handler) putReadBuf(buf *[]byte) {
	h.readBufPool.Put(buf)
}

func (h *handler) getResBuf() *bytes.Buffer {
	return h.resBufPool.Get().(*bytes.Buffer)
}

func (h *handler) putResBuf(buf *bytes.Buffer) {
	buf.Reset()
	h.resBufPool.Put(buf)
}

func (h *handler) getPayload() *payload.Payload {
	return h.pldPool.Get().(*payload.Payload)
}

func (h *handler) putPayload(pld *payload.Payload) {
	pld.Body = nil
	pld.Context = nil
	h.pldPool.Put(pld)
}
