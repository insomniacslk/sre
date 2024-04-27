package cli

import (
	"context"
	"fmt"
	"insomniac/sre/pkg/config"
	"log"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func pagerParseTime(s string) (*time.Time, error) {
	now := time.Now()
	if s == "now" {
		return &now, nil
	}
	// try to parse in RFC3339 format first
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return &t, nil
	}
	// then try to parse a duration
	d, err := time.ParseDuration(s)
	if err == nil {
		t := now.Add(-d)
		return &t, nil
	}
	return nil, fmt.Errorf("failed to parse time string, it is neither RFC3339 nor duration")
}

func NewNotificationsCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "notifications",
		Short: "Show notificationss via the oncall tool (PagerDuty)",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			now := time.Now()
			start := now.Add(-time.Hour)
			end := now
			loc, err := time.LoadLocation(cfg.Timezone)
			if err != nil {
				log.Fatalf("Cannot load timezone %q: %v", cfg.Timezone, err)
			}
			if len(args) > 0 {
				t, err := pagerParseTime(args[0])
				if err != nil {
					logrus.Fatalf("Failed to parse start time: %v", err)
				}
				start = *t
			}
			if len(args) > 1 {
				t, err := pagerParseTime(args[1])
				if err != nil {
					logrus.Fatalf("Failed to parse start time: %v", err)
				}
				end = *t
			}
			ctx := context.Background()
			client := pagerduty.NewClient(cfg.PagerDuty.UserToken)
			opts := pagerduty.ListNotificationOptions{
				Since: start.String(),
				Until: end.String(),
				Limit: 100, // 100 is the maximum allowed by PagerDuty's API
			}
			allNotifications := make([]pagerduty.Notification, 0)
			for {
				resp, err := client.ListNotificationsWithContext(ctx, opts)
				if err != nil {
					logrus.Fatalf("Failed to list notifications: %v", err)
				}
				allNotifications = append(allNotifications, resp.Notifications...)
				if !resp.More {
					break
				}
				opts.Offset += 1
			}
			for _, n := range allNotifications {
				startedAt, err := pagerParseTime(n.StartedAt)
				if err != nil {
					logrus.Fatalf("Failed to parse time string %q: %v", n.StartedAt, err)
				}
				start := startedAt.In(loc)
				switch n.Type {
				case "email_notification":
					fmt.Printf(bold("[%s]")+": E-mail to %s (%s)\n", start.String(), toAnsiURL(n.User.Summary, n.User.HTMLURL), toAnsiURL(n.Address, "mailto:"+n.Address))
				case "phone_notification":
					fmt.Printf(bold("[%s]")+": Phone call to %s (%s)\n", start.String(), toAnsiURL(n.User.Summary, n.User.HTMLURL), toAnsiURL(n.Address, "phone:"+n.Address))
				case "sms_notification":
					fmt.Printf(bold("[%s]")+": SMS to %s (%s)\n", start.String(), toAnsiURL(n.User.Summary, n.User.HTMLURL), toAnsiURL(n.Address, "sms:"+n.Address))
				case "push_notification":
					fmt.Printf(bold("[%s]")+": Push to %s\n", start.String(), toAnsiURL(n.User.Summary, n.User.HTMLURL))
				default:
					fmt.Printf(bold("[%s]")+": %s to %s (%s)\n", start.String(), n.Type, toAnsiURL(n.User.Summary, n.User.HTMLURL), n.Address)
					fmt.Printf("%+v\n", n)
				}
			}
			fmt.Printf("Found %d notifications between %s and %s", len(allNotifications), start, end)
		},
	}
}
