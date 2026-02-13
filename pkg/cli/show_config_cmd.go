package cli

import (
	"fmt"

	"github.com/insomniacslk/sre/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	_ "embed"

	"github.com/spf13/cobra"
)

func NewShowConfigCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "show-config",
		Aliases: []string{"config", "sc"},
		Short:   "Print the current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Debugf("Running show-config command")
			output, err := yaml.Marshal(cfg)
			if err != nil {
				logrus.Fatalf("Failed to marshal config to YAML: %v", err)
			}
			fmt.Println(string(output))
		},
	}
}
