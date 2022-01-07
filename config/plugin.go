package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/roadrunner-server/api/plugins/v2/config"
	"github.com/spf13/viper"
	"github.com/spiral/errors"
)

const (
	PluginName       string = "config"
	versionKey       string = "version"
	defaultVersion   string = "2.6"
	defaultRRVersion string = "2.7"
)

type Plugin struct {
	viper     *viper.Viper
	Path      string
	Prefix    string
	Type      string
	ReadInCfg []byte
	// user defined Flags in the form of <option>.<key> = <value>
	// which overwrites initial config key
	Flags []string

	// RRVersion passed from the Endure.
	RRVersion string

	// All plugins common parameters
	CommonConfig *config.General
}

// Init config provider.
func (p *Plugin) Init() error {
	const op = errors.Op("config_plugin_init")
	p.viper = viper.New()
	// If user provided []byte data with config, read it and ignore Path and Prefix
	if p.ReadInCfg != nil && p.Type != "" {
		p.viper.SetConfigType("yaml")
		return p.viper.ReadConfig(bytes.NewBuffer(p.ReadInCfg))
	}

	// read in environment variables that match
	p.viper.AutomaticEnv()
	if p.Prefix == "" {
		return errors.E(op, errors.Str("prefix should be set"))
	}

	p.viper.SetEnvPrefix(p.Prefix)
	if p.Path == "" {
		return errors.E(op, errors.Str("path should be set"))
	}

	p.viper.SetConfigFile(p.Path)
	p.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := p.viper.ReadInConfig()
	if err != nil {
		return errors.E(op, err)
	}

	// get configuration version
	ver := p.viper.Get(versionKey)
	if ver == nil {
		// default version (versioning start point is 2.6)
		ver = defaultVersion
	}

	if _, ok := ver.(string); !ok {
		return errors.E(op, errors.Errorf("version should be a string, actual type: %T", ver))
	}

	// RR gets config feature starting v2.7, so, it's default
	// but this only needed for tests, because starting v2.7 rr-binary will pass the version automatically.
	if p.RRVersion == "" || p.RRVersion == "local" {
		p.RRVersion = defaultRRVersion
	}

	// configuration version
	cfgV, err := version.NewSemver(ver.(string))
	if err != nil {
		return errors.E(op, err)
	}

	// RR version
	rrV, err := version.NewSemver(p.RRVersion)
	if err != nil {
		return errors.E(op, err)
	}

	// default version (2.6.0)
	defV, err := version.NewSemver(defaultVersion)
	if err != nil {
		return errors.E(op, err)
	}

	// probably user set too old version
	if cfgV.LessThan(defV) {
		return errors.E(op, errors.Errorf("too old configuration version used, should be at least 2.6"))
	}

	// if RR version is less than configuration version (2.6 RR and 2.7 config)
	if rrV.LessThan(cfgV) {
		return errors.E(op, errors.Errorf("RR version is older than configuration version, RR version: %s, configuration version: %s", p.RRVersion, ver.(string)))
	}

	if !rrV.GreaterThanOrEqual(cfgV) {
		return errors.Str("configuration version can't be greater that RR version")
	}

	// if rr version is equal to the configuration version, skip transition
	// but we can have versions like 2.7.0 (config) and 2.7.2 (RR version), they are not equal, but we don't change config
	// in the bugfix versions. We should additionally check the minor versions
	if !rrV.Equal(cfgV) {
		// minor RR and minor config
		if (rrV.Segments64()[0] == cfgV.Segments64()[0]) && rrV.Segments64()[1] != cfgV.Segments64()[1] {
			// transform from the older config to the recent RR version
			err = transition(cfgV, rrV, p.viper)
			if err != nil {
				return errors.E(op, err)
			}
		}
	}

	if p.CommonConfig == nil {
		p.CommonConfig = &config.General{}
	}

	p.CommonConfig.RRVersion = rrV

	// automatically inject ENV variables using ${ENV} pattern
	for _, key := range p.viper.AllKeys() {
		val := p.viper.Get(key)
		p.viper.Set(key, parseEnv(val))
	}

	// override config Flags
	if len(p.Flags) > 0 {
		for _, f := range p.Flags {
			key, val, err := parseFlag(f)
			if err != nil {
				return errors.E(op, err)
			}

			p.viper.Set(key, val)
		}
	}

	return nil
}

// Overwrite overwrites existing config with provided values
func (p *Plugin) Overwrite(values map[string]interface{}) error {
	for key, value := range values {
		p.viper.Set(key, value)
	}

	return nil
}

// UnmarshalKey reads configuration section into configuration object.
func (p *Plugin) UnmarshalKey(name string, out interface{}) error {
	const op = errors.Op("config_plugin_unmarshal_key")
	err := p.viper.UnmarshalKey(name, &out)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (p *Plugin) Unmarshal(out interface{}) error {
	const op = errors.Op("config_plugin_unmarshal")
	err := p.viper.Unmarshal(&out)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Get raw config in a form of config section.
func (p *Plugin) Get(name string) interface{} {
	return p.viper.Get(name)
}

// Has checks if config section exists.
func (p *Plugin) Has(name string) bool {
	return p.viper.IsSet(name)
}

// GetCommonConfig Returns common config parameters
func (p *Plugin) GetCommonConfig() *config.General {
	return p.CommonConfig
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop() error {
	return nil
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}

func parseFlag(flag string) (string, string, error) {
	const op = errors.Op("parse_flag")
	if !strings.Contains(flag, "=") {
		return "", "", errors.E(op, errors.Errorf("invalid flag `%s`", flag))
	}

	parts := strings.SplitN(strings.TrimLeft(flag, " \"'`"), "=", 2)
	if len(parts) < 2 {
		return "", "", errors.Str("usage: -o key=value")
	}

	if parts[0] == "" {
		return "", "", errors.Str("key should not be empty")
	}

	if parts[1] == "" {
		return "", "", errors.Str("value should not be empty")
	}

	return strings.Trim(parts[0], " \n\t"), parseValue(strings.Trim(parts[1], " \n\t")), nil
}

func parseValue(value string) string {
	escape := []rune(value)[0]

	if escape == '"' || escape == '\'' || escape == '`' {
		value = strings.Trim(value, string(escape))
		value = strings.ReplaceAll(value, fmt.Sprintf("\\%s", string(escape)), string(escape))
	}

	return value
}

func parseEnv(value interface{}) interface{} {
	str, ok := value.(string)
	if !ok || len(str) <= 3 {
		return value
	}

	if str[0:2] == "${" && str[len(str)-1:] == "}" {
		if v, ok := os.LookupEnv(str[2 : len(str)-1]); ok {
			return v
		}
	}

	return str
}
