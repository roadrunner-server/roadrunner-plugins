package transport

type Driver interface {
	Init()
	Serve()
	Listen()
	Shutdown()
}

type Constructor interface {
	DConstruct() (error, Driver)
}
