package protocol

type errorResp struct {
	Msg     string              `json:"message"`
	Requeue bool                `json:"requeue"`
	Delay   int64               `json:"delay_seconds"`
	Headers map[string][]string `json:"headers"`
}

type queueResp struct {
	Queue   string `json:"queue"`
	Payload string `json:"payload"`
}
