package config

import (
	"fmt"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

var (
	LogLevels       []string
	DefaultLogLevel string
)

func init() {
	// build list of log level strings
	logrusLogLevels := make([]string, 0, len(logrus.AllLevels))
	for _, l := range logrus.AllLevels {
		ls, _ := l.MarshalText()
		logrusLogLevels = append(logrusLogLevels, string(ls))
	}
	LogLevels = logrusLogLevels
	mll, _ := logrus.InfoLevel.MarshalText()
	DefaultLogLevel = string(mll)
}

type Config struct {
	Timezone        string                `mapstructure:"timezone"`
	LogLevel        string                `mapstructure:"loglevel"`
	Omg             OmgConfig             `mapstructure:"omg"`
	Tools           ToolsConfig           `mapstructure:"tools"`
	Vpn             VpnConfig             `mapstructure:"vpn"`
	PagerDuty       PagerDutyConfig       `mapstructure:"pagerduty"`
	Oncall          OncallConfig          `mapstructure:"oncall"`

	// these fields are not coming from the config file and are set from the outside
	ConfigDir     string       `mapstructure:"-"`
	LogLevelValue logrus.Level `mapstructure:"-"`
}

type OmgConfig struct {
	Template string `mapstructure:"template"`
}

func (o *OmgConfig) Validate(cfg *Config) error {
	if o.Template == "" {
		return fmt.Errorf("`omg.template` cannot be empty")
	}
	p, err := homedir.Expand(o.Template)
	if err != nil {
		return fmt.Errorf("failed to expand `omg.template` %q: %w", o.Template, err)
	}
	o.Template = p
	return nil
}

type ToolsConfig []struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
	URL         string `mapstructure:"url"`
}

func (t *ToolsConfig) Validate(cfg *Config) error {
	for idx, tool := range *t {
		if tool.Name == "" {
			return fmt.Errorf("missing tool name at index %d", idx)
		}
	}
	return nil
}

type VpnConfig struct {
	Executable      string              `mapstructure:"executable"`
	Endpoints       []map[string]string `mapstructure:"endpoints"`
	DefaultEndpoint string              `mapstructure:"default_endpoint"`
	PidFile         string              `mapstructure:"pid_file"`
}

func (v *VpnConfig) Validate(cfg *Config) error {
	// no validation at this stage
	return nil
}

type OncallConfig struct {
	DefaultQuery            string `mapstructure:"default_query"`
	DefaultSchedule         string `mapstructure:"default_schedule"`
	DefaultScheduleDuration string `mapstructure:"default_schedule_duration"`
}

func (o *OncallConfig) Validate(cfg *Config) error {
	// no validation at this stage
	return nil
}

type PagerDutyConfig struct {
	UserToken string   `mapstructure:"user_token"`
	Teams     []string `mapstructure:"teams"`
}

func (p *PagerDutyConfig) Validate(cfg *Config) error {
	if p.UserToken == "" {
		return fmt.Errorf("missing or empty `user_token`")
	}
	return nil
}

func (c *Config) Validate() error {
	if err := c.Omg.Validate(c); err != nil {
		return fmt.Errorf("invalid `omg` config: %w", err)
	}
	if err := c.Tools.Validate(c); err != nil {
		return fmt.Errorf("invalid `tools` config: %w", err)
	}
	if err := c.Vpn.Validate(c); err != nil {
		return fmt.Errorf("invalid `vpn` config: %w", err)
	}
	if err := c.Oncall.Validate(c); err != nil {
		return fmt.Errorf("invalid `oncall` config: %w", err)
	}
	if err := c.PagerDuty.Validate(c); err != nil {
		return fmt.Errorf("invalid `pagerduty` config: %w", err)
	}

	return nil
}
