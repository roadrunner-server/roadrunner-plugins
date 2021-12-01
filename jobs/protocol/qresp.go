package protocol

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner-plugins/v2/api/jobs"
	"github.com/spiral/roadrunner-plugins/v2/utils"
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
