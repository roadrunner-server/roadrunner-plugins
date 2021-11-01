package tcp

type rpc struct {
	p *Plugin
}

func (r *rpc) Close(uuid string, ret *bool) error {
	*ret = true
	return r.p.Close(uuid)
}
