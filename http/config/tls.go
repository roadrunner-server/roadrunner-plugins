package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/spiral/errors"
)

// SSL defines https server configuration.
type SSL struct {
	// Address to listen as HTTPS server, defaults to 0.0.0.0:443.
	Address string

	// ACME configuration
	Acme *AcmeConfig `mapstructure:"acme"`

	// Redirect when enabled forces all http connections to switch to https.
	Redirect bool

	// Key defined private server key.
	Key string

	// Cert is https certificate.
	Cert string

	// Root CA file
	RootCA string `mapstructure:"root_ca"`

	// internal
	host string
	Port int
}

func (s *SSL) InitDefaults() error {
	if s.Acme != nil {
		err := s.Acme.InitDefaults()
		if err != nil {
			return err
		}
	}

	if s.Address == "" {
		s.Address = "127.0.0.1:443"
	}

	return nil
}

func (s *SSL) Valid() error {
	const op = errors.Op("ssl_valid")

	parts := strings.Split(s.Address, ":")
	switch len(parts) {
	// :443 form
	// 127.0.0.1:443 form
	// use 0.0.0.0 as host and 443 as port
	case 2:
		if parts[0] == "" {
			s.host = "127.0.0.1"
		} else {
			s.host = parts[0]
		}

		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return errors.E(op, err)
		}
		s.Port = port
	default:
		return errors.E(op, errors.Errorf("unknown format, accepted format is [:<port> or <host>:<port>], provided: %s", s.Address))
	}

	// the user use they own certificates
	if s.Acme == nil {
		if _, err := os.Stat(s.Key); err != nil {
			if os.IsNotExist(err) {
				return errors.E(op, errors.Errorf("key file '%s' does not exists", s.Key))
			}

			return err
		}

		if _, err := os.Stat(s.Cert); err != nil {
			if os.IsNotExist(err) {
				return errors.E(op, errors.Errorf("cert file '%s' does not exists", s.Cert))
			}

			return err
		}
	}

	// RootCA is optional, but if provided - check it
	if s.RootCA != "" {
		if _, err := os.Stat(s.RootCA); err != nil {
			if os.IsNotExist(err) {
				return errors.E(op, errors.Errorf("root ca path provided, but path '%s' does not exists", s.RootCA))
			}
			return err
		}
	}

	return nil
}
