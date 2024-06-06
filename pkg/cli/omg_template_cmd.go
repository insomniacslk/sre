package cli

import (
	_ "embed"
	"fmt"

	"github.com/insomniacslk/sre/pkg/config"

	"github.com/spf13/cobra"
)

//go:embed omg.template.example
var omgTemplateExample string

func NewOmgTemplateCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "omg-template-example",
		Short: "Print an example omg.template file",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(omgTemplateExample)
		},
	}
}
