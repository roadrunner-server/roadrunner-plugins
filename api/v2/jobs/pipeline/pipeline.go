package pipeline

import (
	json "github.com/goccy/go-json"
	"github.com/spiral/roadrunner/v2/utils"
)

// Pipeline defines pipeline options.
type Pipeline map[string]interface{}

const (
	priority string = "priority"
	driver   string = "driver"
	name     string = "name"
	queue    string = "queue"

	// config
	config string = "config"
)

// With pipeline value
func (p *Pipeline) With(name string, value interface{}) {
	(*p)[name] = value
}

// Name returns pipeline name.
func (p Pipeline) Name() string {
	// https://github.com/spiral/roadrunner-jobs/blob/master/src/Queue/CreateInfo.php#L81
	// In the PHP client library used the wrong key name
	// should be "name" instead of "queue"
	if p.String(name, "") != "" {
		return p.String(name, "")
	}

	return p.String(queue, "")
}

// Driver associated with the pipeline.
func (p Pipeline) Driver() string {
	return p.String(driver, "")
}

// Has checks if value presented in pipeline.
func (p Pipeline) Has(name string) bool {
	if _, ok := p[name]; ok {
		return true
	} else {
		// check the config section if exists
		if val, ok := p[config]; ok {
			if rv, ok := val.(map[string]interface{}); ok {
				if _, ok := rv[name]; ok {
					return true
				}
				return false
			}
		}
	}

	return false
}

// String must return option value as string or return default value.
func (p Pipeline) String(name string, d string) string {
	if value, ok := p[name]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	} else {
		// check the config section if exists
		if val, ok := p[config]; ok {
			if rv, ok := val.(map[string]interface{}); ok {
				if rv[name] != "" {
					return rv[name].(string)
				}
			}
		}
	}

	return d
}

// Int must return option value as string or return default value.
func (p Pipeline) Int(name string, d int) int {
	if value, ok := p[name]; ok {
		if i, ok := value.(int); ok {
			return i
		}
	} else {
		// check the config section if exists
		if val, ok := p[config]; ok {
			if rv, ok := val.(map[string]interface{}); ok {
				if rv[name] != nil {
					return rv[name].(int)
				}
			}
		}
	}

	return d
}

// Bool must return option value as bool or return default value.
func (p Pipeline) Bool(name string, d bool) bool {
	if value, ok := p[name]; ok {
		if i, ok := value.(string); ok {
			switch i {
			case "true":
				return true
			case "false":
				return false
			default:
				return false
			}
		}
	} else {
		// check the config section if exists
		if val, ok := p[config]; ok {
			if rv, ok := val.(map[string]interface{}); ok {
				if rv[name] != nil {
					if i, ok := value.(string); ok {
						switch i {
						case "true":
							return true
						case "false":
							return false
						default:
							return false
						}
					}
				}
			}
		}
	}

	return d
}

// Map must return nested map value or empty config.
// Here might be sqs attributes or tags for example
func (p Pipeline) Map(name string, out map[string]string) error {
	if value, ok := p[name]; ok {
		if m, ok := value.(string); ok {
			err := json.Unmarshal(utils.AsBytes(m), &out)
			if err != nil {
				return err
			}
		}
	} else {
		// check the config section if exists
		if val, ok := p[config]; ok {
			if rv, ok := val.(map[string]interface{}); ok {
				if val, ok := rv[name]; ok {
					if m, ok := val.(string); ok {
						err := json.Unmarshal(utils.AsBytes(m), &out)
						if err != nil {
							return err
						}
					}
					return nil
				}
			}
		}
	}

	return nil
}

// Priority returns default pipeline priority
func (p Pipeline) Priority() int64 {
	if value, ok := p[priority]; ok {
		if v, ok := value.(int64); ok {
			return v
		}
	} else {
		// check the config section if exists
		if val, ok := p[config]; ok {
			if rv, ok := val.(map[string]interface{}); ok {
				if rv[name] != nil {
					return rv[name].(int64)
				}
			}
		}
	}

	return 10
}
