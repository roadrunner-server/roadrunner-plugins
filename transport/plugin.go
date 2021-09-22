package transport

const (
	name string = "transport"
)

type Plugin struct{}

func (p *Plugin) Init() error {
	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	return errCh
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Name() string {
	return name
}
