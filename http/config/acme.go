package config

import (
	"github.com/spiral/errors"
)

type AcmeConfig struct {
	// directory to save the certificates, le_certs default
	CacheDir string `mapstructure:"cache_dir"`
	// User email, mandatory
	Email string `mapstructure:"email"`
	// PK name
	PrivateKeyName string `mapstructure:"private_key_name"`
	// CRT name
	CertificateName string `mapstructure:"certificate_name"`
	// supported values: http-01, tlsalpn-01
	ChallengeType string `mapstructure:"challenge_type"`
	// ChallengePort
	ChallengePort         string `mapstructure:"challenge_port"`
	ChallengeIface        string `mapstructure:"challenge_iface"`
	UseProductionEndpoint bool   `mapstructure:"use_production_endpoint"`
	// host white list, default will be
	// Linux: /proc/sys/kernel/hostname
	// Android: localhost
	Domains []string `mapstructure:"domains"`
	// if true - RR will obtain the certificates and put them into the certs_dir
	ObtainCertificates bool `mapstructure:"obtain_certificates"`
}

func (ac *AcmeConfig) InitDefaults() error {
	if ac.CacheDir == "" {
		ac.CacheDir = "rr_cache_dir"
	}

	if len(ac.Domains) == 0 {
		return errors.Str("should be at least 1 domain")
	}

	if ac.PrivateKeyName == "" {
		ac.PrivateKeyName = "private.key"
	}

	if ac.CertificateName == "" {
		ac.CertificateName = "certificate.crt"
	}

	if ac.ChallengeType == "" {
		ac.ChallengeType = "http-01"
		if ac.ChallengePort == "" {
			ac.ChallengePort = "80"
		}
	}

	return nil
}
