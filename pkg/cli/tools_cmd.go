package cli

import (
	"log"

	"github.com/insomniacslk/sre/pkg/config"
	"github.com/insomniacslk/sre/pkg/tools"

	"github.com/spf13/cobra"
)

func NewToolsCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "tools",
		Short: "Manage SRE tools",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := tools.List(&cfg.Tools); err != nil {
				log.Fatalf("Failed to list tools: %v", err)
			}
		},
	}
}
