package handler

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner/v2/payload"
)

func (h *handler) generate(event string) ([]byte, error) {
	si := h.getServInfo(event)
	pld, err := json.Marshal(si)
	if err != nil {
		h.putServInfo(si)
		return nil, err
	}

	h.putServInfo(si)
	return pld, nil
}

func (h *handler) sendClose() {
	c, err := h.generate(EventClose)
	if err != nil {
		h.log.Error("payload marshaling error", "error", err)
		return
	}
	pld := &payload.Payload{
		Context: c,
	}

	_, _ = h.wPool(pld)
}
