package cli

import (
	"github.com/insomniacslk/sre/pkg/config"

	"github.com/spf13/cobra"
)

func NewOncallCmd(cfg *config.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:     "oncall",
		Aliases: []string{"oc"},
		Short:   "Interact with the oncall tool (PagerDuty). Commands: ep (escalation policy, shortcut: %), user (shortcut: @), sc (schedule, shortcut: +), oncall",
		Args:    cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(NewOncallSearchCmd(cfg))
	cmd.AddCommand(NewOncallScheduleCmd(cfg))
	cmd.AddCommand(NewOncallUserCmd(cfg))
	cmd.AddCommand(NewOncallEscalationPolicyCmd(cfg))
	return &cmd
}
