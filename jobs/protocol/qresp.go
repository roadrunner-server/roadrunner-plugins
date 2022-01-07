package protocol

import (
	json "github.com/goccy/go-json"
	"github.com/roadrunner-server/api/plugins/v2/jobs"
	"github.com/spiral/roadrunner/v2/utils"
)

// data - data to redirect to the queue
func (rh *RespHandler) handleQueueResp(data []byte, jb jobs.Acknowledger) error {
	qs := rh.getQResp()
	defer rh.putQResp(qs)

	err := json.Unmarshal(data, qs)
	if err != nil {
		return err
	}

	err = jb.Respond(utils.AsBytes(qs.Payload), qs.Queue)
	if err != nil {
		return err
	}

	return nil
}
