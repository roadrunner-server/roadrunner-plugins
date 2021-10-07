package service

import "time"

// Env variables type alias
type Env map[string]string

// Service represents particular service configuration
type Service struct {
	Command         string        `mapstructure:"command"`
	Output          string        `mapstructure:"log_output"`
	ProcessNum      int           `mapstructure:"process_num"`
	ExecTimeout     time.Duration `mapstructure:"exec_timeout"`
	RemainAfterExit bool          `mapstructure:"remain_after_exit"`
	RestartSec      uint64        `mapstructure:"restart_sec"`
	Env             Env           `mapstructure:"env"`
}

// Config for the services
type Config struct {
	Services map[string]Service `mapstructure:"service"`
}

func (c *Config) InitDefault() {
	if len(c.Services) > 0 {
		for k, v := range c.Services {
			val := c.Services[k]
			c.Services[k] = val

			if v.ProcessNum == 0 {
				val := c.Services[k]
				val.ProcessNum = 1
				c.Services[k] = val
			}
			if v.RestartSec == 0 {
				val := c.Services[k]
				val.RestartSec = 30
				c.Services[k] = val
			}
		}
	}
}
