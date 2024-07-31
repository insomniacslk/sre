package config

import (
	"fmt"

	homedir "github.com/mitchellh/go-homedir"
)

type Config struct {
	Timezone  string          `mapstructure:"timezone"`
	Omg       OmgConfig       `mapstructure:"omg"`
	Tools     ToolsConfig     `mapstructure:"tools"`
	Vpn       VpnConfig       `mapstructure:"vpn"`
	PagerDuty PagerDutyConfig `mapstructure:"pagerduty"`
	Oncall    OncallConfig    `mapstructure:"oncall"`
	Schedule  ScheduleConfig `mapstructure:"schedule"`
	Page      PageConfig `mapstructure:"page"`
}

type OmgConfig struct {
	Template string `mapstructure:"template"`
}

type ToolsConfig []struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
	URL         string `mapstructure:"url"`
}

type VpnConfig struct {
	Executable      string              `mapstructure:"executable"`
	Endpoints       []map[string]string `mapstructure:"endpoints"`
	DefaultEndpoint string              `mapstructure:"default_endpoint"`
	PidFile         string              `mapstructure:"pid_file"`
}

type OncallConfig struct {
	DefaultAction string `mapstructure:"default_action"`
	DefaultQuery  string `mapstructure:"default_query"`
}

type ScheduleConfig struct {
	DefaultSchedule string `mapstructure:"default_schedule"`
}

type PageConfig struct {
	Client string `mapstructure:"client"`
}

type PagerDutyConfig struct {
	UserToken string   `mapstructure:"user_token"`
	Teams     []string `mapstructure:"teams"`
}

func (c *Config) Validate() error {
	// TODO move validation to the `omg` command
	// validate `omg`
	if c.Omg.Template == "" {
		return fmt.Errorf("`omg.template` cannot be empty")
	}
	p, err := homedir.Expand(c.Omg.Template)
	if err != nil {
		return fmt.Errorf("faild to expand `omg.template` %q: %w", c.Omg.Template, err)
	}
	c.Omg.Template = p

	// validate `tools`
	// nothing to do for now

	return nil
}
