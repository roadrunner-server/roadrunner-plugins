package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestValidateVersion(t *testing.T) {
	ver := "2"
	err := validateVersion(ver)
	require.NoError(t, err)

	ver = "2.a.b"
	err = validateVersion(ver)
	require.Error(t, err)

	ver = ""
	err = validateVersion(ver)
	require.Error(t, err)

	ver = "2.4"
	err = validateVersion(ver)
	require.NoError(t, err)

	ver = "2.4.4"
	err = validateVersion(ver)
	require.NoError(t, err)

	ver = "2a.4b.4c"
	err = validateVersion(ver)
	require.Error(t, err)

	ver = "Ⅷ.Ⅷ.Ⅷ"
	err = validateVersion(ver)
	require.Error(t, err)

	ver = "2.Ⅷ.Ⅷ"
	err = validateVersion(ver)
	require.Error(t, err)
}

func TestTransition(t *testing.T) {
	from := "2"
	to := "2.7"

	v := viper.New()

	err := transition(from, to, v)
	require.Error(t, err)

	from = "2.6.3"
	to = "2.7"

	err = transition(from, to, v)
	require.Error(t, err)

	from = "2.6"
	to = "2.7"

	err = transition(from, to, v)
	require.NoError(t, err)
}
