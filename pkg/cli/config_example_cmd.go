package cli

import (
	"fmt"

	"github.com/insomniacslk/sre/pkg/config"
	"github.com/sirupsen/logrus"

	_ "embed"

	"github.com/spf13/cobra"
)

//go:embed config.yml.example
var configFileExample string

func NewConfigExampleCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "config-example",
		Short: "Print an example config file",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Debugf("Running config-example command")
			fmt.Println(configFileExample)
		},
	}
}
