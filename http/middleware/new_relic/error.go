package newrelic

import (
	"strings"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/spiral/roadrunner/v2/utils"
)

const (
	message    string = "message"
	class      string = "class"
	attributes string = "attributes"
)

/*
Error headers format:

message:foo
class:bar
attributes:baz

*/
func handleErr(headers []string) error {
	nrErr := newrelic.Error{}

	for i := 0; i < len(headers); i++ {
		key, value := split(utils.AsBytes(headers[i]))
		switch {
		case strings.EqualFold(utils.AsString(key), message):
			nrErr.Message = utils.AsString(value)
		case strings.EqualFold(utils.AsString(key), class):
			nrErr.Class = utils.AsString(value)
		case strings.EqualFold(utils.AsString(key), attributes):
			nrErr.Attributes = map[string]interface{}{
				"attributes": utils.AsString(value),
			}
		}
	}

	return nrErr
}
