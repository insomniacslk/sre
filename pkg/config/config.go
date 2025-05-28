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
}

type OmgConfig struct {
	Template string `mapstructure:"template"`
}

func (o *OmgConfig) Validate() error {
	if o.Template == "" {
		return fmt.Errorf("`omg.template` cannot be empty")
	}
	p, err := homedir.Expand(o.Template)
	if err != nil {
		return fmt.Errorf("faild to expand `omg.template` %q: %w", o.Template, err)
	}
	o.Template = p
	return nil
}

type ToolsConfig []struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
	URL         string `mapstructure:"url"`
}

func (t *ToolsConfig) Validate() error {
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

func (v *VpnConfig) Validate() error {
	// no validation at this stage
	return nil
}

type OncallConfig struct {
	DefaultQuery            string `mapstructure:"default_query"`
	DefaultSchedule         string `mapstructure:"default_schedule"`
	DefaultScheduleDuration string `mapstructure:"default_schedule_duration"`
}

func (o *OncallConfig) Validate() error {
	// no validation at this stage
	return nil
}

type PagerDutyConfig struct {
	UserToken string   `mapstructure:"user_token"`
	Teams     []string `mapstructure:"teams"`
}

func (p *PagerDutyConfig) Validate() error {
	if p.UserToken == "" {
		return fmt.Errorf("missing or empty `user_token`")
	}
	return nil
}

func (c *Config) Validate() error {
	if err := c.Omg.Validate(); err != nil {
		return fmt.Errorf("invalid `omg` config: %w", err)
	}
	if err := c.Tools.Validate(); err != nil {
		return fmt.Errorf("invalid `tools` config: %w", err)
	}
	if err := c.Vpn.Validate(); err != nil {
		return fmt.Errorf("invalid `vpn` config: %w", err)
	}
	if err := c.Oncall.Validate(); err != nil {
		return fmt.Errorf("invalid `oncall` config: %w", err)
	}
	if err := c.PagerDuty.Validate(); err != nil {
		return fmt.Errorf("invalid `oncall` config: %w", err)
	}

	return nil
}
