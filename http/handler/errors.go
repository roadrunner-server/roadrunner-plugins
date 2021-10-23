//go:build !windows
// +build !windows

package handler

import (
	"errors"
)

// Broken pipe
var errEPIPE = errors.New("EPIPE(32) -> connection reset by peer")
