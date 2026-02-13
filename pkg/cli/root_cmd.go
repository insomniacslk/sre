package cli

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/insomniacslk/sre/pkg/config"
)

var globalRootCmd *cobra.Command

func InitRootCmd(progname string) {
	if globalRootCmd == nil {
		rootCmd := &cobra.Command{
			Use:   progname,
			Short: fmt.Sprintf("%q is the uber-CLI for an SRE team.", progname),
			Long:  fmt.Sprintf("%s is the uber-CLI for an SRE team. It includes subcommands that help with incident response, documentation, oncall, and much more", progname),
			Args:  cobra.MinimumNArgs(1),
			Run:   nil,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := GetConfig()
				if err != nil {
					return fmt.Errorf("failed to get config: %w", err)
				}
				// override log level if set via command line.
				// If not set, use the one from config file, and if that one is not set either, use the default
				var logLevel string
				changed := cmd.Flags().Changed("log-level")
				if changed {
					lls, err := cmd.Flags().GetString("log-level")
					if err != nil {
						return fmt.Errorf("failed to get string value for --log-level: %w", err)
					}
					logLevel = lls
				} else {
					if cfg.LogLevel != "" {
						logLevel = cfg.LogLevel
					} else {
						logLevel = config.DefaultLogLevel
					}
				}
				ll, err := logrus.ParseLevel(logLevel)
				if err != nil {
					return fmt.Errorf("invalid log level %q passed via command line: %w", logLevel, err)
				}
				logrus.SetLevel(ll)
				cfg.LogLevelValue = ll
				return nil
			},
		}
		rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Configuration file")
		rootCmd.PersistentFlags().StringVarP(&flagLogLevel, "log-level", "L", config.DefaultLogLevel, fmt.Sprintf("Set log level. One of %v", config.LogLevels))
		globalRootCmd = rootCmd
	}
}

func GetRootCmd() (*cobra.Command, error) {
	if globalRootCmd == nil {
		return nil, fmt.Errorf("root command not initialized, must call InitRootCmd")
	}
	return globalRootCmd, nil
}

func WithSubcommands(progname string, cmds ...*cobra.Command) (*cobra.Command, error) {
	InitRootCmd(progname)
	rootCmd, err := GetRootCmd()
	if err != nil {
		return nil, err
	}
	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}
	return rootCmd, nil
}
