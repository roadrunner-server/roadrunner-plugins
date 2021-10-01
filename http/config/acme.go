package config

import (
	"github.com/spiral/errors"
)

type AcmeConfig struct {
	// directory to save the certificates, le_certs default
	CacheDir string `mapstructure:"cache_dir"`

	// User email, mandatory
	Email string `mapstructure:"email"`

	// supported values: http-01, tlsalpn-01
	ChallengeType string `mapstructure:"challenge_type"`

	// The alternate port to use for the ACME HTTP challenge
	AltHTTPPort int `mapstructure:"alt_http_port"`

	// The alternate port to use for the ACME TLS-ALPN
	AltTLSALPNPort int `mapstructure:"alt_tlsalpn_port"`

	// Use LE production endpoint or staging
	UseProductionEndpoint bool `mapstructure:"use_production_endpoint"`

	// Domains to obtain certificates
	Domains []string `mapstructure:"domains"`
}

func (ac *AcmeConfig) InitDefaults() error {
	if ac.CacheDir == "" {
		ac.CacheDir = "rr_cache_dir"
	}

	if ac.Email == "" {
		return errors.Str("email could not be empty")
	}

	if len(ac.Domains) == 0 {
		return errors.Str("should be at least 1 domain")
	}

	if ac.ChallengeType == "" {
		ac.ChallengeType = "http-01"
		if ac.AltHTTPPort == 0 {
			ac.AltHTTPPort = 80
		}
	}

	return nil
}
