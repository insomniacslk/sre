package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/insomniacslk/sre/pkg/ansi"
	"github.com/insomniacslk/sre/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	OncallCmd.AddCommand(OncallShortlistCmd)
}

var OncallShortlistCmd = &cobra.Command{
	Use:     "shortlist [filter]",
	Aliases: []string{"sl"},
	Short:   "Show current oncalls for a curated shortlist of components (from config)",
	Long: `Show the current oncall for each component in the configured shortlist.

The shortlist is defined under `+ "`oncall.shortlist`" + ` in the config file. Each
entry maps a component/team name to a PagerDuty schedule, resolved either by a
free-text `+ "`query`" + ` (like `+ "`oncall search`" + `) or a pinned `+ "`schedule_id`" + `.

With an optional [filter] argument, only entries whose name or component contain
the (case-insensitive) filter are shown, e.g. "oncall shortlist security".`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.Debugf("Running oncall shortlist command")
		ctx := context.Background()
		cfg, err := GetConfig()
		if err != nil {
			return err
		}

		entries := cfg.Oncall.Shortlist
		if len(entries) == 0 {
			return fmt.Errorf("no shortlist configured; add entries under `oncall.shortlist` (see the `config-example` subcommand)")
		}

		filter := strings.ToLower(strings.Join(args, " "))
		selected := make([]config.OncallShortlistEntry, 0, len(entries))
		for _, e := range entries {
			if filter == "" ||
				strings.Contains(strings.ToLower(e.Name), filter) ||
				strings.Contains(strings.ToLower(e.Component), filter) {
				selected = append(selected, e)
			}
		}
		if len(selected) == 0 {
			fmt.Printf("No shortlist entries match %q\n", filter)
			return nil
		}

		client := pagerduty.NewClient(cfg.PagerDuty.UserToken)
		until := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

		if filter == "" {
			fmt.Printf("Oncall shortlist (%d entries)\n", len(selected))
		} else {
			fmt.Printf("Oncall shortlist matching %q (%d entries)\n", filter, len(selected))
		}
		for idx, e := range selected {
			label := ansi.Bold(e.Name)
			oncalls, err := resolveShortlistOncalls(ctx, client, e, until)
			if err != nil {
				fmt.Printf("%d) %s — error: %v\n", idx+1, label, err)
				continue
			}
			if len(oncalls) == 0 {
				where := e.Query
				if e.ScheduleID != "" {
					where = "schedule " + e.ScheduleID
				}
				fmt.Printf("%d) %s — no current oncall found (%s)\n", idx+1, label, where)
				continue
			}
			for _, oc := range oncalls {
				sched := oc.Schedule.Summary
				if oc.Schedule.HTMLURL != "" {
					sched = ansi.ToURL(sched, oc.Schedule.HTMLURL)
				}
				user := oc.User.Summary
				if oc.User.HTMLURL != "" {
					user = ansi.ToURL(user, oc.User.HTMLURL)
				}
				fmt.Printf("%d) %s — %s [ID: %s] oncall: %s (%s)\n",
					idx+1, label, sched, oc.Schedule.ID, user, oc.User.Email)
			}
		}
		return nil
	},
}

// resolveShortlistOncalls returns the current oncall entries for a shortlist
// entry, resolving its schedule(s) either by pinned ScheduleID or by searching
// schedules with the entry's Query. Only the currently-active oncall per
// schedule is returned (deduplicated by schedule ID).
func resolveShortlistOncalls(
	ctx context.Context,
	client *pagerduty.Client,
	e config.OncallShortlistEntry,
	until string,
) ([]pagerduty.OnCall, error) {
	var scheduleIDs []string
	if e.ScheduleID != "" {
		scheduleIDs = []string{e.ScheduleID}
	} else {
		sResp, err := client.ListSchedulesWithContext(ctx, pagerduty.ListSchedulesOptions{Query: e.Query})
		if err != nil {
			return nil, fmt.Errorf("failed to search schedules for %q: %w", e.Query, err)
		}
		for _, sc := range sResp.Schedules {
			scheduleIDs = append(scheduleIDs, sc.ID)
		}
		if len(scheduleIDs) == 0 {
			return nil, nil
		}
	}

	resp, err := client.ListOnCallsWithContext(ctx, pagerduty.ListOnCallOptions{
		ScheduleIDs: scheduleIDs,
		Includes:    []string{"users"},
		Until:       until,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get oncalls: %w", err)
	}

	// Keep only the first (current) oncall per schedule so a multi-layer
	// escalation policy doesn't print several rows for one schedule.
	seen := make(map[string]struct{})
	out := make([]pagerduty.OnCall, 0, len(resp.OnCalls))
	for _, oc := range resp.OnCalls {
		if _, ok := seen[oc.Schedule.ID]; ok {
			continue
		}
		seen[oc.Schedule.ID] = struct{}{}
		out = append(out, oc)
	}
	return out, nil
}
