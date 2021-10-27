package tcp

type rpc struct {
	p *Plugin
}

func (r *rpc) Close(uuid string, _ *bool) error {
	return r.p.Close(uuid)
}
