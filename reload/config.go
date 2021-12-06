package reload

import (
	"time"

	"github.com/spiral/errors"
)

// Config is a Reload configuration point.
type Config struct {
	// Interval is a global refresh interval
	Interval time.Duration `mapstructure:"interval"`

	// Patterns is a global file patterns to watch. It will be applied to every directory in project
	Patterns []string `mapstructure:"patterns"`

	// Plugins is set of services which would be reloaded in case of FS changes
	Plugins map[string]ServiceConfig `mapstructure:"services"`
}

type ServiceConfig struct {
	// Recursive is options to use nested files from root folder
	Recursive bool `mapstructure:"recursive"`

	// Patterns is per-service specific files to watch
	Patterns []string `mapstructure:"patterns"`

	// Dirs is per-service specific dirs which will be combined with Patterns
	Dirs []string `mapstructure:"dirs"`

	// Ignore is set of files which would not be watched
	Ignore []string `mapstructure:"ignore"`
}

// InitDefaults sets missing values to their default values.
func (c *Config) InitDefaults() {
	if c.Interval == 0 {
		c.Interval = time.Second
	}
	if c.Patterns == nil {
		c.Patterns = []string{".php"}
	}
}

// Valid validates the configuration.
func (c *Config) Valid() error {
	const op = errors.Op("reload_plugin_valid")
	if c.Interval < time.Second {
		return errors.E(op, errors.Str("too short interval"))
	}

	if c.Plugins == nil {
		return errors.E(op, errors.Str("should add at least 1 service"))
	} else if len(c.Plugins) == 0 {
		return errors.E(op, errors.Str("service initialized, however, no config added"))
	}

	return nil
}
