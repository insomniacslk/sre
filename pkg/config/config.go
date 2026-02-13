package config

import (
	"fmt"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/xhit/go-str2duration/v2"
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
	OncallGenerator OncallGeneratorConfig `mapstructure:"oncall_generator"`

	// this field is not coming from the config file and is set from the outside
	ConfigDir string `mapstructure:"-"`
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

type OncallGeneratorConfig struct {
	ScheduleDuration          string `mapstructure:"schedule_duration"`
	PublicHolidayCalendarFile string `mapstructure:"public_holiday_calendar_file"`
	Members                   []struct {
		Name        string `mapstructure:"name"`
		Email       string `mapstructure:"email"`
		Constraints struct {
			Timezone           string `mapstructure:"timezone"`
			EarliestOncallHour uint   `mapstructure:"earliest_oncall_hour"`
			LatestOncallHour   uint   `mapstructure:"latest_oncall_hour"`
			PublicHolidays     struct {
				CountryName  string      `mapstructure:"country_name"`
				IncludeDates []time.Time `mapstructure:"include_dates"`
				ExcludeDates []time.Time `mapstructure:"exclude_dates"`
			} `mapstructure:"public_holidays"`
		} `mapstructure:"constraints"`
	} `mapstructure:"members"`
}

func (o *OncallGeneratorConfig) Validate(cfg *Config) error {
	if _, err := str2duration.ParseDuration(o.ScheduleDuration); err != nil {
		return fmt.Errorf("invalid schedule_duration: %w", err)
	}
	cf, err := homedir.Expand(o.PublicHolidayCalendarFile)
	if err != nil {
		return fmt.Errorf("failed to expand public_holiday_calendar_file path: %v", err)
	}
	if !filepath.IsAbs(cf) {
		cf = filepath.Join(cfg.ConfigDir, cf)
	}
	o.PublicHolidayCalendarFile = cf
	for idx, m := range o.Members {
		if m.Name == "" {
			return fmt.Errorf("empty or missing member name at index %d", idx)
		}
		if m.Email == "" {
			return fmt.Errorf("empty or missing member email at index %d", idx)
		}
		if m.Constraints.Timezone == "" {
			return fmt.Errorf("missing or empty timezone for user %q", m.Name)
		}
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
		return fmt.Errorf("invalid `oncall` config: %w", err)
	}
	if err := c.OncallGenerator.Validate(c); err != nil {
		return fmt.Errorf("invalid `oncall_generator` config: %w", err)
	}

	return nil
}
