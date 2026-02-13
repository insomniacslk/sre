package cli

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		}
		// build list of log level strings
		logrusLogLevels := make([]string, 0, len(logrus.AllLevels))
		for _, l := range logrus.AllLevels {
			ls, _ := l.MarshalText()
			logrusLogLevels = append(logrusLogLevels, string(ls))
		}
		mll, _ := logrus.InfoLevel.MarshalText()
		defaultLogLevel := string(mll)
		rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Configuration file")
		rootCmd.PersistentFlags().StringVarP(&flagLogLevel, "log-level", "L", defaultLogLevel, fmt.Sprintf("Set log level. One of %v", logrusLogLevels))
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
