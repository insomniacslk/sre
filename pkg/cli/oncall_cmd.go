package cli

import (
	"github.com/spf13/cobra"
)

var OncallCmd = &cobra.Command{
	Use:     "oncall",
	Aliases: []string{"oc"},
	Short:   "Interact with the oncall tool (PagerDuty)",
	Args:    cobra.MinimumNArgs(1),
}
