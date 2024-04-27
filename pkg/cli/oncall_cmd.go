package cli

import (
	"context"
	"fmt"
	"insomniac/sre/pkg/config"
	"log"
	"strings"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewOncallCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "oncall",
		Short: "Interact with the oncall tool (PagerDuty)",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			client := pagerduty.NewClient(cfg.PagerDuty.UserToken)

			separator := strings.Repeat("-", 80)
			query := cfg.Oncall.EscalationPolicy
			if len(args) != 0 {
				query = args[0]
			}
			if query == "" {
				log.Fatalf("No schedule pattern specified")
			}
			if query[0] == '@' {
				// searching for a user
				query = query[1:]
				opts := pagerduty.ListUsersOptions{
					Query:    query,
					Includes: []string{"contact_methods"},
				}
				resp, err := client.ListUsersWithContext(ctx, opts)
				if err != nil {
					log.Fatalf("Failed to get escalation policies: %v", err)
				}
				for _, u := range resp.Users {
					fmt.Printf(bold("Name      :")+" %s\n", u.Name)
					fmt.Printf(bold("Email     :")+" %s\n", toAnsiURL(u.Email, "mailto:"+u.Email))
					fmt.Printf(bold("Title     :")+" %s\n", u.JobTitle)
					fmt.Printf(bold("Time zone :")+" %s\n", u.Timezone)
					fmt.Print(bold("Teams     :\n"))
					for _, t := range u.Teams {
						fmt.Printf("  %s\n", toAnsiURL(t.Summary, t.HTMLURL))
					}
					fmt.Print(bold("Contacts  :\n"))
					for _, c := range u.ContactMethods {
						switch c.Type {
						case "email_contact_method_reference":
							fmt.Printf("  E-mail: %s (%s)\n", toAnsiURL(c.Address, "mailto:"+c.Address), c.Summary)
						case "phone_contact_method":
							phoneNum := fmt.Sprintf("+%d%s", c.CountryCode, c.Address)
							fmt.Printf("  Phone : %s (%s)\n", toAnsiURL(phoneNum, "tel:"+phoneNum), c.Summary)
						case "push_notification_contact_method":
							fmt.Printf("  Push  : %s\n", c.Summary)
						case "sms_contact_method":
							phoneNum := fmt.Sprintf("+%d%s", c.CountryCode, c.Address)
							fmt.Printf("  SMS   : %s (%s)\n", toAnsiURL(phoneNum, "sms:"+phoneNum), c.Summary)
						default:
							fmt.Printf("  %s: %s (%s)\n", c.Address, c.Type, c.Summary)
						}
					}
					fmt.Println(separator)
				}
			} else {
				// search for an escalation policy
				opts := pagerduty.ListEscalationPoliciesOptions{
					Query:    query,
					Includes: []string{"services", "targets", "teams"},
				}
				resp, err := client.ListEscalationPoliciesWithContext(ctx, opts)
				if err != nil {
					log.Fatalf("Failed to get escalation policies: %v", err)
				}
				// cache teams to minimize API calls
				teams := make(map[string]*pagerduty.Team)
				users := make(map[string]*pagerduty.User)
				for _, ep := range resp.EscalationPolicies {
					fmt.Printf(bold("Name:")+" %s\n", toAnsiURL(ep.Name, ep.HTMLURL))
					fmt.Printf(bold("Description:")+" %s\n", ep.Description)
					fmt.Print(bold("Services:") + "\n")
					for _, s := range ep.Services {
						fmt.Printf("    %s\n", toAnsiURL(s.Summary, s.HTMLURL))
					}
					fmt.Print(bold("Teams:") + "\n")
					for _, t := range ep.Teams {
						// FIXME: for some reason the escalation policies API
						// does not return team details, despite being specified
						// in `includes`. So, for now, fetch each team
						// individually.
						team, ok := teams[t.ID]
						if !ok {
							// fetch team
							team, err = client.GetTeamWithContext(ctx, t.ID)
							if err != nil {
								log.Fatalf("Failed to fetch team with ID %q: %v", t.ID, err)
							}
							teams[team.ID] = team
						}
						fmt.Printf("    %s (Description: %q)\n", toAnsiURL(team.Summary, team.HTMLURL), team.Description)
					}
					fmt.Print(bold("Escalation rules:") + "\n")
					for _, r := range ep.EscalationRules {
						for _, t := range r.Targets {
							fmt.Printf("    %s\n", toAnsiURL(t.Summary, t.HTMLURL))
							now := time.Now()
							pastHour := now.Add(-time.Hour)
							nextHour := now.Add(time.Hour)
							oopts := pagerduty.ListOnCallUsersOptions{
								Since: pastHour.String(),
								Until: now.String(),
							}
							// FIXME this approach is not great. I get the
							// oncalls for the past hour, and for the next hour,
							// and print them without repetitions. Need to find
							// a way to get subsequent oncalls from the
							// PagerDuty API instead.
							if t.Type == "user" {
								user, ok := users[t.ID]
								if !ok {
									// fetch the user
									opts := pagerduty.GetUserOptions{
										Includes: []string{"contact_methods"},
									}
									user, err = client.GetUserWithContext(ctx, t.ID, opts)
									if err != nil {
										log.Fatalf("Failed to get escalation policies: %v", err)
									}
									users[user.ID] = user
								}
								fmt.Printf("        %s (%s)\n", toAnsiURL(user.Name, user.HTMLURL), toAnsiURL(user.Email, "mailto:"+user.Email))
							} else {
								currentOncalls, err := client.ListOnCallUsersWithContext(ctx, t.ID, oopts)
								if err != nil {
									logrus.Warningf("Failed to get users for schedule %s, skipping. Error was: %v", t.Summary, err)
								}
								oopts.Since = now.String()
								oopts.Until = nextHour.String()
								nextOncalls, err := client.ListOnCallUsersWithContext(ctx, t.ID, oopts)
								if err != nil {
									logrus.Warningf("Failed to get users for schedule %s (ID: %s), skipping. Error was: %v", t.Summary, t.ID, err)
								}
								alreadyPrinted := make(map[string]struct{})
								for _, u := range currentOncalls {
									fmt.Printf("        %s (%s)\n", toAnsiURL(u.Name, u.HTMLURL), toAnsiURL(u.Email, "mailto:"+u.Email))
									alreadyPrinted[u.ID] = struct{}{}
								}
								for _, u := range nextOncalls {
									if _, ok := alreadyPrinted[u.ID]; ok {
										// already printed, don't print it again
										continue
									}
									fmt.Printf("        %s (%s)\n", toAnsiURL(u.Name, u.HTMLURL), toAnsiURL(u.Email, "mailto:"+u.Email))
								}
							}
						}
					}
					fmt.Println(separator)
				}
			}
		},
	}
}
