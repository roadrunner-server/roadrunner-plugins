package newrelic

import (
	"os"

	"github.com/spiral/errors"
)

const newRelicLK string = "NEW_RELIC_LICENSE_KEY"

type Config struct {
	FromEnv    bool   `mapstructure:"from_env"`
	AppName    string `mapstructure:"app_name"`
	LicenseKey string `mapstructure:"license_key"`
}

// 1 - config
// 2 - env variable

func (c *Config) InitDefaults() error {
	if c.LicenseKey == "" {
		lc := os.Getenv(newRelicLK)

		if lc == "" {
			return errors.Str("license key should not be empty")
		}

		c.LicenseKey = os.Getenv(newRelicLK)
	}

	if c.AppName == "" {
		c.AppName = "roadrunner"
	}

	return nil
}
