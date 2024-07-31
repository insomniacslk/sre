package cli

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/insomniacslk/sre/pkg/ansi"
	"github.com/insomniacslk/sre/pkg/config"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewScheduleCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "schedule",
		Short: "Show the schedule of a given oncall (PagerDuty)",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			client := pagerduty.NewClient(cfg.PagerDuty.UserToken)

			scheduleID := cfg.Schedule.DefaultSchedule
			if len(args) > 0 {
				scheduleID = args[0]
			}
			if scheduleID == "" {
				logrus.Fatalf("No schedule ID specified")
			}
			log.Printf("Searching schedules matching %q", scheduleID)
			// search for a schedule
			now := time.Now()
			opts := pagerduty.GetScheduleOptions{
				Since:    now.Format(time.RFC3339),
				Until:    now.Add(24 * time.Hour).Format(time.RFC3339),
				TimeZone: cfg.Timezone,
			}
			sched, err := client.GetScheduleWithContext(ctx, scheduleID, opts)
			if err != nil {
				logrus.Fatalf("Failed to get schedules: %v", err)
			}
			fmt.Printf("%s\n", ansi.Bold(sched.Name))
			fmt.Printf("    Summary: %s\n", ansi.ToURL(sched.Summary, sched.HTMLURL))
			fmt.Printf("    Description: %s\n", sched.Description)
			fmt.Printf("    Users:\n")
			for _, user := range sched.Users {
				fmt.Printf("        %s\n", ansi.ToURL(user.Summary, user.HTMLURL))
			}
			for _, entry := range sched.FinalSchedule.RenderedScheduleEntries {
				start, err := time.Parse(time.RFC3339, entry.Start)
				if err != nil {
					log.Fatalf("Start time %q is not in RFC3339 format: %v", entry.Start, err)
				}
				end, err := time.Parse(time.RFC3339, entry.End)
				if err != nil {
					log.Fatalf("End time %q is not in RFC3339 format: %v", entry.End, err)
				}
				timeFmt := "Mon 02 Jan 2006"
				fmt.Printf("%s - %s (%s) %+v\n",
					start.Format(timeFmt),
					end.Format(timeFmt),
					end.Sub(start),
					ansi.ToURL(entry.User.Summary, entry.User.HTMLURL),
				)
			}
		},
	}
}
