package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/insomniacslk/sre/pkg/ansi"
	"github.com/insomniacslk/sre/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewOncallUserCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "user",
		Aliases: []string{"u"},
		Short:   "Show user information",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			client := pagerduty.NewClient(cfg.PagerDuty.UserToken)

			// search for a user
			query := strings.Join(args, " ")[1:]
			separator := strings.Repeat("-", 80)
			opts := pagerduty.ListUsersOptions{
				Query:    query,
				Includes: []string{"contact_methods"},
			}
			resp, err := client.ListUsersWithContext(ctx, opts)
			if err != nil {
				logrus.Fatalf("Failed to get escalation policies: %v", err)
			}
			for _, u := range resp.Users {
				fmt.Printf(ansi.Bold("Name      :")+" %s\n", u.Name)
				fmt.Printf(ansi.Bold("Email     :")+" %s\n", ansi.ToURL(u.Email, "mailto:"+u.Email))
				fmt.Printf(ansi.Bold("Title     :")+" %s\n", u.JobTitle)
				fmt.Printf(ansi.Bold("Time zone :")+" %s\n", u.Timezone)
				fmt.Print(ansi.Bold("Teams     :\n"))
				for _, t := range u.Teams {
					fmt.Printf("  %s\n", ansi.ToURL(t.Summary, t.HTMLURL))
				}
				fmt.Print(ansi.Bold("Contacts  :\n"))
				for _, c := range u.ContactMethods {
					switch c.Type {
					case "email_contact_method_reference":
						fmt.Printf("  E-mail: %s (%s)\n", ansi.ToURL(c.Address, "mailto:"+c.Address), c.Summary)
					case "phone_contact_method":
						phoneNum := fmt.Sprintf("+%d%s", c.CountryCode, c.Address)
						fmt.Printf("  Phone : %s (%s)\n", ansi.ToURL(phoneNum, "tel:"+phoneNum), c.Summary)
					case "push_notification_contact_method":
						fmt.Printf("  Push  : %s\n", c.Summary)
					case "sms_contact_method":
						phoneNum := fmt.Sprintf("+%d%s", c.CountryCode, c.Address)
						fmt.Printf("  SMS   : %s (%s)\n", ansi.ToURL(phoneNum, "sms:"+phoneNum), c.Summary)
					default:
						fmt.Printf("  %s: %s (%s)\n", c.Address, c.Type, c.Summary)
					}
				}
				fmt.Println(separator)
			}
		},
	}
}
