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
	Page      PageConfig      `mapstructure:"page"`
}

type OmgConfig struct {
	Template string `mapstructure:"template"`
}

type ToolsConfig struct {
	InstallDir string `mapstructure:"install_dir"`
	SrcDir     string `mapstructure:"src_dir"`
}

type VpnConfig struct {
	Executable      string              `mapstructure:"executable"`
	Endpoints       []map[string]string `mapstructure:"endpoints"`
	DefaultEndpoint string              `mapstructure:"default_endpoint"`
	PidFile         string              `mapstructure:"pid_file"`
}

type OncallConfig struct {
	EscalationPolicy string `mapstructure:"escalation_policy"`
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
		return fmt.Errorf("faild to expand `tools.install_dir` %q: %w", c.Tools.InstallDir, err)
	}
	c.Omg.Template = p

	// TODO move validation to the `tools` command
	// validate `tools`
	if c.Tools.InstallDir == "" {
		return fmt.Errorf("`tools.install_dir` cannot be empty")
	}
	p, err = homedir.Expand(c.Tools.InstallDir)
	if err != nil {
		return fmt.Errorf("faild to expand `tools.install_dir` %q: %w", c.Tools.InstallDir, err)
	}
	c.Tools.InstallDir = p
	if c.Tools.SrcDir == "" {
		return fmt.Errorf("`tools.src_dir` cannot be empty")
	}
	p, err = homedir.Expand(c.Tools.SrcDir)
	if err != nil {
		return fmt.Errorf("faild to expand `tools.src_dir` %q: %w", c.Tools.SrcDir, err)
	}
	c.Tools.SrcDir = p

	return nil
}
