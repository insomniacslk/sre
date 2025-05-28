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

func NewOncallSearchCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "search",
		Aliases: []string{"s"},
		Short:   "Search oncall schedule information",
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// search for an oncall
			scheduleIDs := make([]string, 0)
			query := cfg.Oncall.DefaultQuery
			if len(args) > 0 {
				query = strings.Join(args, " ")
			}
			client := pagerduty.NewClient(cfg.PagerDuty.UserToken)

			// Resolve schedule name into ID(s)
			sOpts := pagerduty.ListSchedulesOptions{
				Query: query,
			}
			fmt.Printf("Searching schedule with query %q\n", query)
			sResp, err := client.ListSchedulesWithContext(ctx, sOpts)
			if err != nil {
				logrus.Fatalf("Failed to get schedules: %v", err)
			}
			for _, sc := range sResp.Schedules {
				scheduleIDs = append(scheduleIDs, sc.ID)
			}

			opts := pagerduty.ListOnCallOptions{
				ScheduleIDs: scheduleIDs,
				Includes:    []string{"users"},
				Until:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			}
			resp, err := client.ListOnCallsWithContext(ctx, opts)
			if err != nil {
				logrus.Fatalf("Failed to get schedules: %v", err)
			}
			oncallBySchedule := make(map[string][]*pagerduty.OnCall)
			for _, oc := range resp.OnCalls {
				_, ok := oncallBySchedule[oc.Schedule.Summary]
				if !ok {
					oncallBySchedule[oc.Schedule.Summary] = []*pagerduty.OnCall{&oc}
				} else {
					oncallBySchedule[oc.Schedule.Summary] = append(
						oncallBySchedule[oc.Schedule.Summary],
						&oc,
					)
				}
			}
			fmt.Printf("Found %d schedules\n", len(oncallBySchedule))
			idx := 0
			for sched, oncalls := range oncallBySchedule {
				schedURL := ""
				if len(oncalls) > 0 {
					// assume that the first schedule URL is the same for
					// all the other items, because they were grouped by
					// schedule name
					schedURL = oncalls[0].Schedule.HTMLURL
					fmt.Printf("%d) %s ", idx, ansi.Bold(ansi.ToURL(sched, schedURL)))
				} else {
					fmt.Printf("%d) %s ", idx, ansi.Bold(sched))
				}
				fmt.Printf("ID: %s oncall: %s (%s)\n",
					oncalls[0].Schedule.ID,
					ansi.ToURL(oncalls[0].User.Summary, oncalls[0].User.HTMLURL),
					oncalls[0].User.Email,
				)
				idx++
			}
		},
	}
}
