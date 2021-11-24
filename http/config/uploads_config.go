package config

import (
	"os"
)

// Uploads describes file location and controls access to them.
type Uploads struct {
	// Dir contains name of directory to control access to.
	Dir string `mapstructure:"dir"`

	// Forbid specifies list of file extensions which are forbidden for access.
	// Example: .php, .exe, .bat, .htaccess and etc.
	Forbid []string `mapstructure:"forbid"`

	// Allowed files
	Allow []string `mapstructure:"allow"`

	// internal
	Forbidden map[string]struct{} `mapstructure:"-"`
	Allowed   map[string]struct{} `mapstructure:"-"`
}

// InitDefaults sets missing values to their default values.
func (cfg *Uploads) InitDefaults() error {
	if len(cfg.Forbid) == 0 {
		cfg.Forbid = []string{".php", ".exe", ".bat"}
	}

	if cfg.Dir == "" {
		cfg.Dir = os.TempDir()
	}

	cfg.Forbidden = make(map[string]struct{})
	cfg.Allowed = make(map[string]struct{})

	for i := 0; i < len(cfg.Forbid); i++ {
		cfg.Forbidden[cfg.Forbid[i]] = struct{}{}
	}

	for i := 0; i < len(cfg.Allow); i++ {
		cfg.Allowed[cfg.Allow[i]] = struct{}{}
	}

	for k := range cfg.Forbidden {
		delete(cfg.Allowed, k)
	}

	cfg.Forbid = nil
	cfg.Allow = nil

	return nil
}
