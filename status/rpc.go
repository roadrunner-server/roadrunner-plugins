package status

import (
	"github.com/roadrunner-server/api/plugins/v2/status"
	"github.com/spiral/errors"
	"go.uber.org/zap"
)

type rpc struct {
	srv *Plugin
	log *zap.Logger
}

// Status return current status of the provided plugin
func (rpc *rpc) Status(service string, status *status.Status) error {
	const op = errors.Op("checker_rpc_status")
	rpc.log.Debug("Status method was invoked", zap.String("plugin", service))
	st, err := rpc.srv.status(service)
	if err != nil {
		return errors.E(op, err)
	}

	if st != nil {
		*status = *st
	}

	rpc.log.Debug("status code", zap.Int("code", st.Code))
	rpc.log.Debug("successfully finished the Status method")
	return nil
}

// Ready return the readiness check of the provided plugin
func (rpc *rpc) Ready(service string, status *status.Status) error {
	const op = errors.Op("checker_rpc_ready")
	rpc.log.Debug("Ready method was invoked", zap.String("plugin", service))
	st, err := rpc.srv.ready(service)
	if err != nil {
		return errors.E(op, err)
	}

	if st != nil {
		*status = *st
	}

	rpc.log.Debug("status code", zap.Int("code", st.Code))
	rpc.log.Debug("successfully finished the Ready method")
	return nil
}
