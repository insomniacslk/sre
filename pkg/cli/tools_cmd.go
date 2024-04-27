package cli

import (
	"insomniac/sre/pkg/config"
	"insomniac/sre/pkg/tools"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewToolsCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "tools",
		Short: "Manage SRE tools",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			subcmd := args[0]
			switch subcmd {
			case "install":
				err = tools.Install(&cfg.Tools, args[1:])
			case "update":
				err = tools.Update(&cfg.Tools, args[1:])
			case "list":
				err = tools.List(&cfg.Tools)
			default:
				logrus.Fatalf("Unknown subcommand %q", subcmd)
			}
			if err != nil {
				log.Fatalf("Subcommand %q failed: %v", subcmd, err)
			}
		},
	}
}
