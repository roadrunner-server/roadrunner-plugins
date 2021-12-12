package config

import (
	"time"

	"github.com/hashicorp/go-version"
)

// General is the part of the config plugin which contains general for the whole RR2 parameters
// For example - http timeouts, headers sizes etc and also graceful shutdown timeout should be the same across whole application
type General struct {
	// GracefulTimeout for the temporal and http
	GracefulTimeout time.Duration

	// RRVersion passed from the rr-binary
	RRVersion *version.Version
}
