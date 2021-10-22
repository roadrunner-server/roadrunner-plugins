package newrelic

import (
	"os"

	"github.com/spiral/errors"
)

const (
	newRelicLK      string = "NEW_RELIC_LICENSE_KEY"
	newRelicAppName string = "NEW_RELIC_APP_NAME"
)

type Config struct {
	AppName    string `mapstructure:"app_name"`
	LicenseKey string `mapstructure:"license_key"`
}

func (c *Config) InitDefaults() error {
	if c.LicenseKey == "" {
		lc := os.Getenv(newRelicLK)

		if lc == "" {
			return errors.Str("license key should not be empty")
		}

		c.LicenseKey = os.Getenv(newRelicLK)
	}

	if c.AppName == "" {
		an := os.Getenv(newRelicAppName)

		if an == "" {
			return errors.Str("application name should not be empty")
		}

		c.AppName = os.Getenv(newRelicAppName)
	}

	return nil
}
