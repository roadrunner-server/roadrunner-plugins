package protocol

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner-plugins/v2/internal/common/jobs"
)

func (rh *RespHandler) handleErrResp(data []byte, jb jobs.Acknowledger) error {
	er := rh.getErrResp()
	defer rh.putErrResp(er)

	err := json.Unmarshal(data, er)
	if err != nil {
		return err
	}

	rh.log.Error("jobs protocol error", "error", er.Msg, "delay", er.Delay, "requeue", er.Requeue)

	// requeue the job
	if er.Requeue {
		err = jb.Requeue(er.Headers, er.Delay)
		if err != nil {
			return err
		}
		return nil
	}

	// user don't want to requeue the job - silently ACK and return nil
	errAck := jb.Ack()
	if errAck != nil {
		rh.log.Error("job acknowledge", "message", er.Msg, "error", errAck)
		// do not return any error
	}

	return nil
}
