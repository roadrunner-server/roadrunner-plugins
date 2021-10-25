package newrelic

import (
	"testing"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	target1 := []string{"message:world", "class:foo", "attributes:bar"}
	err := handleErr(target1)
	require.Equal(t, newrelic.Error{
		Message: "world",
		Class:   "foo",
		Attributes: map[string]interface{}{
			"attributes": "bar",
		},
		Stack: nil,
	}, err)

	target2 := []string{"message:", "class:foo", "attributes:bar"}
	err = handleErr(target2)
	require.Equal(t, newrelic.Error{
		Message: "",
		Class:   "foo",
		Attributes: map[string]interface{}{
			"attributes": "bar",
		},
		Stack: nil,
	}, err)

	target3 := []string{"message:world", "class:", "attributes:bar"}
	err = handleErr(target3)
	require.Equal(t, newrelic.Error{
		Message: "world",
		Class:   "",
		Attributes: map[string]interface{}{
			"attributes": "bar",
		},
		Stack: nil,
	}, err)

	target4 := []string{"message:world", "class:foo", ""}
	err = handleErr(target4)
	require.Equal(t, newrelic.Error{
		Message:    "world",
		Class:      "foo",
		Attributes: nil,
		Stack:      nil,
	}, err)

	target5 := []string{":", ":", ":"}
	err = handleErr(target5)
	require.Equal(t, newrelic.Error{
		Message:    "",
		Class:      "",
		Attributes: nil,
		Stack:      nil,
	}, err)

	target6 := []string{": ", " :", ": "}
	err = handleErr(target6)
	require.Equal(t, newrelic.Error{
		Message:    "",
		Class:      "",
		Attributes: nil,
		Stack:      nil,
	}, err)

	target7 := []string{"message:что", "class:это", "attributes:за язык"}
	err = handleErr(target7)
	require.Equal(t, newrelic.Error{
		Message: "что",
		Class:   "это",
		Attributes: map[string]interface{}{
			"attributes": "за язык",
		},
		Stack: nil,
	}, err)
}
