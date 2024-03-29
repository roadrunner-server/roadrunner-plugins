package config

import (
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pool"
)

// HTTP configures RoadRunner HTTP server.
type HTTP struct {
	// Host and port to handle as http server.
	Address string `mapstructure:"address"`

	// AccessLogs turn on/off, logged at Info log level, default: false
	AccessLogs bool `mapstructure:"access_logs"`

	// Pool configures worker pool.
	Pool *pool.Config `mapstructure:"pool"`

	// InternalErrorCode used to override default 500 (InternalServerError) http code
	InternalErrorCode uint64 `mapstructure:"internal_error_code"`

	// SSLConfig defines https server options.
	SSLConfig *SSL `mapstructure:"ssl"`

	// FCGIConfig configuration. You can use FastCGI without HTTP server.
	FCGIConfig *FCGI `mapstructure:"fcgi"`

	// HTTP2Config configuration
	HTTP2Config *HTTP2 `mapstructure:"http2"`

	// Uploads configures uploads configuration.
	Uploads *Uploads `mapstructure:"uploads"`

	// MaxRequestSize specified max size for payload body in megabytes, set 0 to unlimited.
	MaxRequestSize uint64 `mapstructure:"max_request_size"`

	// TrustedSubnets declare IP subnets which are allowed to set ip using X-Real-Ip and X-Forwarded-For
	TrustedSubnets []string `mapstructure:"trusted_subnets"`

	// Env is environment variables passed to the  http pool
	Env map[string]string

	// List of the middleware names (order will be preserved)
	Middleware []string

	// slice of net.IPNet
	Cidrs Cidrs
}

// EnableHTTP is true when http server must run.
func (c *HTTP) EnableHTTP() bool {
	return c.Address != ""
}

// EnableTLS returns true if pool must listen TLS connections.
func (c *HTTP) EnableTLS() bool {
	if c.SSLConfig == nil {
		return false
	}
	if c.SSLConfig.Acme != nil {
		return true
	}
	return c.SSLConfig.Key != "" || c.SSLConfig.Cert != ""
}

// EnableH2C when HTTP/2 extension must be enabled on TCP.
func (c *HTTP) EnableH2C() bool {
	if c.HTTP2Config == nil {
		return false
	}
	return c.HTTP2Config.H2C
}

// EnableFCGI is true when FastCGI server must be enabled.
func (c *HTTP) EnableFCGI() bool {
	if c.FCGIConfig == nil {
		return false
	}
	return c.FCGIConfig.Address != ""
}

func (c *HTTP) EnableACME() bool {
	if c.SSLConfig == nil {
		return false
	}
	return c.SSLConfig.Acme != nil
}

// InitDefaults must populate HTTP values using given HTTP source. Must return error if HTTP is not valid.
func (c *HTTP) InitDefaults() error {
	if c.Pool == nil {
		// default pool
		c.Pool = &pool.Config{
			Debug:           false,
			NumWorkers:      uint64(runtime.NumCPU()),
			MaxJobs:         0,
			AllocateTimeout: time.Second * 60,
			DestroyTimeout:  time.Second * 60,
			Supervisor:      nil,
		}
	}

	if c.InternalErrorCode == 0 {
		c.InternalErrorCode = 500
	}

	if c.HTTP2Config != nil {
		err := c.HTTP2Config.InitDefaults()
		if err != nil {
			return err
		}
	}

	if c.SSLConfig != nil {
		err := c.SSLConfig.InitDefaults()
		if err != nil {
			return err
		}
	}

	if c.Uploads == nil {
		c.Uploads = &Uploads{}
	}

	err := c.Uploads.InitDefaults()
	if err != nil {
		return err
	}

	if c.TrustedSubnets == nil {
		// @see https://en.wikipedia.org/wiki/Reserved_IP_addresses
		c.TrustedSubnets = []string{
			"10.0.0.0/8",
			"127.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			"::1/128",
			"fc00::/7",
			"fe80::/10",
		}
	}

	cidrs, err := ParseCIDRs(c.TrustedSubnets)
	if err != nil {
		return err
	}
	c.Cidrs = cidrs

	return c.Valid()
}

// ParseCIDRs parse IPNet addresses and return slice of its
func ParseCIDRs(subnets []string) (Cidrs, error) {
	c := make(Cidrs, 0, len(subnets))
	for _, cidr := range subnets {
		_, cr, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}

		c = append(c, cr)
	}

	return c, nil
}

// Valid validates the configuration.
func (c *HTTP) Valid() error {
	const op = errors.Op("validation")
	if c.Uploads == nil {
		return errors.E(op, errors.Str("malformed uploads config"))
	}

	if c.Pool == nil {
		return errors.E(op, "malformed pool config")
	}

	if !c.EnableHTTP() && !c.EnableTLS() && !c.EnableFCGI() {
		return errors.E(op, errors.Str("unable to run http service, no method has been specified (http, https, http/2 or FastCGI)"))
	}

	if c.Address != "" && !strings.Contains(c.Address, ":") {
		return errors.E(op, errors.Str("malformed http server address"))
	}

	if c.EnableTLS() {
		err := c.SSLConfig.Valid()
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}
