package config

import (
	"os"
)

type AcmeConfig struct {
	// directory to save the certificates, le_certs default
	CacheDir string `mapstructure:"certs_dir"`
	// User email, mandatory
	Email string `mapstructure:"email"`

	//
	PrivateKeyName string `mapstructure:"private_key_name"`

	CertificateName string `mapstructure:"certificate_name"`

	// supported values: http-01, tlsalpn-01
	ChallengeType string `mapstructure:"challenge_type"`

	ChallengePort string `mapstructure:"challenge_port"`

	ChallengeIface string `mapstructure:"challenge_iface"`

	UseProductionEndpoint bool `mapstructure:"use_production_endpoint"`

	// host white list, default will be
	// Linux: /proc/sys/kernel/hostname
	// Android: localhost
	Domains []string `mapstructure:"domains"`

	// if true - RR will obtain the certificates and put them into the certs_dir
	FirstTime bool
}

func (ac *AcmeConfig) InitDefaults() {
	if ac.CacheDir == "" {
		ac.CacheDir = "le_certs"
	}

	if len(ac.Domains) == 0 {
		name, _ := os.Hostname()
		ac.Domains = append(ac.Domains, name)
	}
}
