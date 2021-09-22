package protocol

type Protocol interface {
	Listen() error
}

// Protocol constructor used to construct any external protocol
type Constructor interface {
	Construct() Protocol
}
