package cli

import (
	"fmt"
	"path/filepath"

	"github.com/insomniacslk/sre/pkg/config"

	"github.com/kirsle/configdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	globalConfig *config.Config
	configFile   string
	flagDebug    bool
)

func GetConfig() (*config.Config, error) {
	if globalConfig == nil {
		return nil, fmt.Errorf("config not initialized, must call InitConfig first")
	}
	return globalConfig, nil
}

func InitConfig(progname string) (*config.Config, error) {
	if flagDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	var configDir string
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
		configDir = filepath.Dir(configFile)
	} else {
		configDir = configdir.LocalConfig(progname)
		logrus.Debugf("Searching for config file in directory %q", configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(configDir)
	}
	viper.AutomaticEnv()

	var cfg config.Config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Warningf("Config file not found, using an empty one. Run the `config-example` subcommand for an example config")
			// return an empty config
			return &cfg, nil
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	logrus.Debugf("Successfully unmarshalled config: %+v", cfg)
	cfg.ConfigDir = configDir
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}
	logrus.Debugf("Configuration validated successfully")
	globalConfig = &cfg
	return globalConfig, nil
}
