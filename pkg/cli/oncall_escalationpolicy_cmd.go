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

func NewOncallEscalationPolicyCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "escalationpolicy",
		Aliases: []string{"ep", "escalation-policy", "escalation_policy"},
		Short:   "Show oncall information",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// search for an escalation policy
			client := pagerduty.NewClient(cfg.PagerDuty.UserToken)
			query := strings.Join(args, " ")
			separator := strings.Repeat("-", 80)
			opts := pagerduty.ListEscalationPoliciesOptions{
				Query:    query,
				Includes: []string{"services", "targets", "teams"},
			}
			resp, err := client.ListEscalationPoliciesWithContext(ctx, opts)
			if err != nil {
				logrus.Fatalf("Failed to get escalation policies: %v", err)
			}
			// cache teams to minimize API calls
			teams := make(map[string]*pagerduty.Team)
			users := make(map[string]*pagerduty.User)
			for _, ep := range resp.EscalationPolicies {
				fmt.Printf(ansi.Bold("Name:")+" %s\n", ansi.ToURL(ep.Name, ep.HTMLURL))
				fmt.Printf(ansi.Bold("Description:")+" %s\n", ep.Description)
				fmt.Print(ansi.Bold("Services:") + "\n")
				for _, s := range ep.Services {
					fmt.Printf("    %s\n", ansi.ToURL(s.Summary, s.HTMLURL))
				}
				fmt.Print(ansi.Bold("Teams:") + "\n")
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
							logrus.Fatalf("Failed to fetch team with ID %q: %v", t.ID, err)
						}
						teams[team.ID] = team
					}
					fmt.Printf("    %s (Description: %q)\n", ansi.ToURL(team.Summary, team.HTMLURL), team.Description)
				}
				fmt.Print(ansi.Bold("Escalation rules:") + "\n")
				for _, r := range ep.EscalationRules {
					for _, t := range r.Targets {
						fmt.Printf("    %s\n", ansi.ToURL(t.Summary, t.HTMLURL))
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
									logrus.Fatalf("Failed to get escalation policies: %v", err)
								}
								users[user.ID] = user
							}
							fmt.Printf("        %s (%s)\n", ansi.ToURL(user.Name, user.HTMLURL), ansi.ToURL(user.Email, "mailto:"+user.Email))
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
								fmt.Printf("        %s (%s)\n", ansi.ToURL(u.Name, u.HTMLURL), ansi.ToURL(u.Email, "mailto:"+u.Email))
								alreadyPrinted[u.ID] = struct{}{}
							}
							for _, u := range nextOncalls {
								if _, ok := alreadyPrinted[u.ID]; ok {
									// already printed, don't print it again
									continue
								}
								fmt.Printf("        %s (%s)\n", ansi.ToURL(u.Name, u.HTMLURL), ansi.ToURL(u.Email, "mailto:"+u.Email))
							}
						}
					}
				}
				fmt.Println(separator)
			}
		},
	}
}
