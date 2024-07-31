package cli

import (
	"fmt"

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
		rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Configuration file")
		rootCmd.PersistentFlags().BoolVarP(&flagDebug, "debug", "d", false, "Print debug messages")
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
