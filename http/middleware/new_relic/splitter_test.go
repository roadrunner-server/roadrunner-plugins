package newrelic

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitter(t *testing.T) {
	target1 := []byte("hello:world")
	key, value := split(target1)
	require.Equal(t, []byte("hello"), key)
	require.Equal(t, []byte("world"), value)

	target2 := []byte(":")
	key, value = split(target2)
	require.Equal(t, []byte(nil), key)
	require.Equal(t, []byte(nil), value)

	target3 := []byte(" : ")
	key, value = split(target3)
	require.Equal(t, []byte(" "), key)
	require.Equal(t, []byte(" "), value)

	target4 := []byte(" :")
	key, value = split(target4)
	require.Equal(t, []byte(" "), key)
	require.Equal(t, []byte{}, value)

	target5 := []byte(": ")
	key, value = split(target5)
	require.Equal(t, []byte(nil), key)
	require.Equal(t, []byte(nil), value)
}
