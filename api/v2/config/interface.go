package config

import (
	"time"

	"github.com/hashicorp/go-version"
)

type Configurer interface {
	// UnmarshalKey takes a single key and unmarshals it into a Struct.
	//
	// func (h *HttpService) Init(cp config.Configurer) error {
	//     h.config := &HttpConfig{}
	//     if err := configProvider.UnmarshalKey("http", h.config); err != nil {
	//         return err
	//     }
	// }
	UnmarshalKey(name string, out interface{}) error

	// Unmarshal unmarshal the config into a Struct. Make sure that the tags
	// on the fields of the structure are properly set.
	Unmarshal(out interface{}) error

	// Get used to get config section
	Get(name string) interface{}

	// Overwrite used to overwrite particular values in the unmarshalled config
	Overwrite(values map[string]interface{}) error

	// Has checks if config section exists.
	Has(name string) bool

	// GetCommonConfig returns General section. Read-only
	GetCommonConfig() *General
}

// General is the part of the config plugin which contains general for the whole RR2 parameters
// For example - http timeouts, headers sizes etc and also graceful shutdown timeout should be the same across whole application
type General struct {
	// GracefulTimeout for the temporal and http
	GracefulTimeout time.Duration

	// RRVersion passed from the rr-binary
	RRVersion *version.Version
}
