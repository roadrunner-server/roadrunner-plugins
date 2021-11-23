package grpc

import (
	"crypto/tls"
	"math"
	"os"
	"strings"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pool"
)

type ClientAuthType string

const (
	NoClientCert               string = "no_client_cert"
	RequestClientCert          string = "request_client_cert"
	RequireAnyClientCert       string = "require_any_client_cert"
	VerifyClientCertIfGiven    string = "verify_client_cert_if_given"
	RequireAndVerifyClientCert string = "require_and_verify_client_cert"
)

type Config struct {
	Listen string   `mapstructure:"listen"`
	Proto  []string `mapstructure:"proto"`

	TLS *TLS `mapstructure:"tls"`

	// Env is environment variables passed to the http pool
	Env map[string]string `mapstructure:"env"`

	GrpcPool              *pool.Config  `mapstructure:"pool"`
	MaxSendMsgSize        int64         `mapstructure:"max_send_msg_size"`
	MaxRecvMsgSize        int64         `mapstructure:"max_recv_msg_size"`
	MaxConnectionIdle     time.Duration `mapstructure:"max_connection_idle"`
	MaxConnectionAge      time.Duration `mapstructure:"max_connection_age"`
	MaxConnectionAgeGrace time.Duration `mapstructure:"max_connection_age_grace"`
	MaxConcurrentStreams  int64         `mapstructure:"max_concurrent_streams"`
	PingTime              time.Duration `mapstructure:"ping_time"`
	Timeout               time.Duration `mapstructure:"timeout"`
}

type TLS struct {
	Key      string `mapstructure:"key"`
	Cert     string `mapstructure:"cert"`
	RootCA   string `mapstructure:"root_ca"`
	AuthType string `mapstructure:"client_auth_type"`
	// auth type
	auth tls.ClientAuthType
}

func (c *Config) InitDefaults() error { //nolint:gocognit
	const op = errors.Op("grpc_plugin_config")
	if c.GrpcPool == nil {
		c.GrpcPool = &pool.Config{}
	}

	c.GrpcPool.InitDefaults()

	if !strings.Contains(c.Listen, ":") {
		return errors.E(op, errors.Errorf("mailformed grpc address, provided: %s", c.Listen))
	}

	for i := 0; i < len(c.Proto); i++ {
		if c.Proto[i] == "" {
			continue
		}

		if _, err := os.Stat(c.Proto[i]); err != nil {
			if os.IsNotExist(err) {
				return errors.E(op, errors.Errorf("proto file '%s' does not exists", c.Proto[i]))
			}

			return errors.E(op, err)
		}
	}

	if c.EnableTLS() {
		if _, err := os.Stat(c.TLS.Key); err != nil {
			if os.IsNotExist(err) {
				return errors.E(op, errors.Errorf("key file '%s' does not exists", c.TLS.Key))
			}

			return errors.E(op, err)
		}

		if _, err := os.Stat(c.TLS.Cert); err != nil {
			if os.IsNotExist(err) {
				return errors.E(op, errors.Errorf("cert file '%s' does not exists", c.TLS.Cert))
			}

			return errors.E(op, err)
		}

		// RootCA is optional, but if provided - check it
		if c.TLS.RootCA != "" {
			// auth type used only for the CA
			switch c.TLS.AuthType {
			case NoClientCert:
				c.TLS.auth = tls.NoClientCert
			case RequestClientCert:
				c.TLS.auth = tls.RequestClientCert
			case RequireAnyClientCert:
				c.TLS.auth = tls.RequireAnyClientCert
			case VerifyClientCertIfGiven:
				c.TLS.auth = tls.VerifyClientCertIfGiven
			case RequireAndVerifyClientCert:
				c.TLS.auth = tls.RequireAndVerifyClientCert
			default:
				c.TLS.auth = tls.NoClientCert
			}
			if _, err := os.Stat(c.TLS.RootCA); err != nil {
				if os.IsNotExist(err) {
					return errors.E(op, errors.Errorf("root ca path provided, but key file '%s' does not exists", c.TLS.RootCA))
				}
				return errors.E(op, err)
			}
		}
	}

	// used to set max time
	infinity := time.Duration(math.MaxInt64)

	if c.PingTime == 0 {
		c.PingTime = time.Hour * 2
	}

	if c.Timeout == 0 {
		c.Timeout = time.Second * 20
	}

	if c.MaxConcurrentStreams == 0 {
		c.MaxConcurrentStreams = 10
	}
	// set default
	if c.MaxConnectionAge == 0 {
		c.MaxConnectionAge = infinity
	}

	// set default
	if c.MaxConnectionIdle == 0 {
		c.MaxConnectionIdle = infinity
	}

	if c.MaxConnectionAgeGrace == 0 {
		c.MaxConnectionAgeGrace = infinity
	}

	if c.MaxRecvMsgSize == 0 {
		c.MaxRecvMsgSize = 1024 * 1024 * 50
	} else {
		c.MaxRecvMsgSize = 1024 * 1024 * c.MaxRecvMsgSize
	}

	if c.MaxSendMsgSize == 0 {
		c.MaxSendMsgSize = 1024 * 1024 * 50
	} else {
		c.MaxSendMsgSize = 1024 * 1024 * c.MaxSendMsgSize
	}

	return nil
}

func (c *Config) EnableTLS() bool {
	if c.TLS != nil {
		return (c.TLS.RootCA != "" && c.TLS.Key != "" && c.TLS.Cert != "") || (c.TLS.Key != "" && c.TLS.Cert != "")
	}

	return false
}
