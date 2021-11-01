package handler

var (
	CLOSE      = []byte("CLOSE")
	CONTINUE   = []byte("CONTINUE")
	WRITECLOSE = []byte("WRITECLOSE")
	WRITE      = []byte("WRITE")
)

type ServerInfo struct {
	RemoteAddr string `json:"remote_addr"`
	Server     string `json:"server"`
	UUID       string `json:"uuid"`
	Event      string `json:"event"`
}

const (
	EventConnected    string = "CONNECTED"
	EventIncomingData string = "DATA"
	EventClose        string = "CLOSE"
)
