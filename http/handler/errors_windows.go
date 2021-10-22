//go:build windows
// +build windows

package handler

import (
	"errors"
)

//Software caused connection abort.
//An established connection was aborted by the software in your host computer,
//possibly due to a data transmission time-out or protocol error.
var errEPIPE = errors.New("WSAECONNABORTED (10053) ->  an established connection was aborted by peer")
