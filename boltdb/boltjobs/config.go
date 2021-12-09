package boltjobs

const (
	file     string = "file"
	prefetch string = "prefetch"
)

type config struct {
	// global
	Permissions int `mapstructure:"permissions"`

	// local
	File     string `mapstructure:"file"`
	Priority int64  `mapstructure:"priority"`
	Prefetch int    `mapstructure:"prefetch"`
}

func (c *config) InitDefaults() {
	if c.File == "" {
		c.File = "rr.db"
	}

	if c.Priority == 0 {
		c.Priority = 10
	}

	if c.Prefetch == 0 {
		c.Prefetch = 1000
	}

	if c.Permissions == 0 {
		c.Permissions = 0777
	}
}
