package cli

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/insomniacslk/sre/pkg/ansi"
	"github.com/insomniacslk/sre/pkg/config"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewIncidentsCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "incidents",
		Short: "Show incidents via the oncall tool (PagerDuty)",
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
			teamsMap := make(map[string]*pagerduty.Team, 0)
			for _, teamStr := range cfg.PagerDuty.Teams {
				if teamStr == "" {
					continue
				}
				if teamStr[0] != '+' {
					// this is a team name, not a team ID
					// fetch the team ID
					// TODO remove duplicates before fetching teams via API
					opts := pagerduty.ListTeamOptions{
						Query: teamStr,
					}
					resp, err := client.ListTeamsWithContext(ctx, opts)
					if err != nil {
						logrus.Warningf("Failed to get team with pattern %q, skipping. Error was: %v", teamStr, err)
					}
					for _, team := range resp.Teams {
						teamsMap[team.ID] = &team
					}
				} else {
					teamID := teamStr[1:]
					teamsMap[teamID] = nil
				}
			}
			teamIDs := make([]string, 0, len(teamsMap))
			for teamID := range teamsMap {
				teamIDs = append(teamIDs, teamID)
			}
			if len(teamIDs) == 0 {
				logrus.Fatalf("No teams found")
			}
			opts := pagerduty.ListIncidentsOptions{
				Since:   start.String(),
				Until:   end.String(),
				Limit:   100, // 100 is the maximum allowed by PagerDuty's API
				TeamIDs: teamIDs,
			}
			allIncidents := make([]pagerduty.Incident, 0)
			for {
				resp, err := client.ListIncidentsWithContext(ctx, opts)
				if err != nil {
					logrus.Fatalf("Failed to list incidents: %v", err)
				}
				allIncidents = append(allIncidents, resp.Incidents...)
				if !resp.More {
					break
				}
				opts.Offset += 1
			}
			for _, incident := range allIncidents {
				createdAt, err := pagerParseTime(incident.CreatedAt)
				if err != nil {
					logrus.Fatalf("Failed to parse time %q: %v", incident.CreatedAt, err)
				}
				fmt.Printf(ansi.Bold("[%s]")+" %s\n", createdAt.In(loc), ansi.ToURL(incident.Summary, incident.HTMLURL))
				fmt.Printf("     Urgency: %s\n", incident.Urgency)
				var statusIcon string
				switch incident.Status {
				case "resolved":
					statusIcon = "✅"
				case "acknowledged":
					statusIcon = "⚠️"
				case "triggered":
					statusIcon = "❌"
				default:
					logrus.Warningf("Unknown incident status %q", incident.Status)
				}
				fmt.Printf("     Status: %s %s\n", incident.Status, statusIcon)
				if incident.Status == "resolved" {
					resolvedAt, err := pagerParseTime(incident.ResolvedAt)
					if err != nil {
						logrus.Fatalf("Failed to parse time %q: %v", incident.ResolvedAt, err)
					}
					fmt.Printf("     Resolved at %s\n", resolvedAt)
				} else {
					fmt.Printf("     Not resolved\n")
				}
				fmt.Printf("     Service: %s\n", ansi.ToURL(incident.Service.Summary, incident.Service.HTMLURL))
				fmt.Printf("     Last changed by: %s\n", ansi.ToURL(incident.LastStatusChangeBy.Summary, incident.LastStatusChangeBy.HTMLURL))
				fmt.Printf("     Trigger: %s\n", incident.FirstTriggerLogEntry.Summary)
				var teams []string
				for _, team := range incident.Teams {
					teams = append(teams, ansi.ToURL(team.Summary, team.HTMLURL))
				}
				fmt.Printf("     Teams: %s\n", strings.Join(teams, ", "))
				fmt.Printf("     Escalation policy: %s\n", ansi.ToURL(incident.EscalationPolicy.Summary, incident.EscalationPolicy.HTMLURL))
			}
			fmt.Printf("Found %d incidents for teams matching %q between %s and %s", len(allIncidents), cfg.PagerDuty.Teams, start, end)
		},
	}
}
