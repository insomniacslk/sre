package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/insomniacslk/sre/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xhit/go-str2duration/v2"
	"gopkg.in/yaml.v3"
)

type CountryName string

type ByCountry map[CountryName][]CustomTime

type HolidayCalendar struct {
	Years map[int]ByCountry `json:"years" yaml:"years"`
}

func (h *HolidayCalendar) ToText() string {
	var ret string
	for year, byCountry := range h.Years {
		ret += fmt.Sprintf("%d:\n", year)
		for loc, dates := range byCountry {
			ret += fmt.Sprintf("    %s\n", loc)
			for _, d := range dates {
				ret += fmt.Sprintf("        %s\n", d.Format(customTimeDefaultFormat))
			}
		}
	}
	return ret
}

func (h *HolidayCalendar) ToJSON() (string, error) {
	ret, err := json.MarshalIndent(h, "", "    ")
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

func (h *HolidayCalendar) ToYAML() (string, error) {
	ret, err := yaml.Marshal(h)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

func NewOncallGeneratorCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "weekendgenerator",
		Aliases: []string{"g", "gen", "weekend"},
		Short:   "Generate oncall schedule",
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			dur, err := str2duration.ParseDuration(cfg.OncallGenerator.ScheduleDuration)
			if err != nil {
				logrus.Fatalf("Invalid schedule duration: %v", err)
			}

			// read the holiday calendar file
			data, err := os.ReadFile(cfg.OncallGenerator.PublicHolidayCalendarFile)
			if err != nil {
				logrus.Fatalf("Failed to read public holiday calendar file %q: %v", cfg.OncallGenerator.PublicHolidayCalendarFile, err)
			}
			var hc HolidayCalendar
			if err := yaml.Unmarshal(data, &hc); err != nil {
				logrus.Fatalf("Failed to unmarshal holiday calendar from YAML: %v", err)
			}

			fmt.Printf("Generating oncall schedule for the next %s\n", dur)
			fmt.Printf("Members:\n")
			for idx, m := range cfg.OncallGenerator.Members {
				fmt.Printf("%d) %s <%s>\n", idx+1, m.Name, m.Email)
				fmt.Printf("    Constraints:\n")
				fmt.Printf("    - timezone        : %s\n", m.Constraints.Timezone)
				fmt.Printf("    - earliest hour   : %d\n", m.Constraints.EarliestOncallHour)
				fmt.Printf("    - latest hour     : %d\n", m.Constraints.LatestOncallHour)
				fmt.Printf("    - public holidays :\n")
				fmt.Printf("            country name: %s\n", m.Constraints.PublicHolidays.CountryName)
				fmt.Printf("            include_dates:\n")
				for _, h := range m.Constraints.PublicHolidays.IncludeDates {
					fmt.Printf("            - %s\n", h.Format("2006-01-02"))
				}
				fmt.Printf("            exclude_dates:\n")
				for _, h := range m.Constraints.PublicHolidays.ExcludeDates {
					fmt.Printf("            - %s\n", h.Format("2006-01-02"))
				}
			}
			schedule, err := GenerateSchedule(&cfg.OncallGenerator)
			if err != nil {
				logrus.Errorf("Failed to generate schedule: %v", err)
			}
			fmt.Printf("Generated schedule: %+v\n", schedule)

			/*
				ctx := context.Background()
				client := pagerduty.NewClient(pagerdutyAPIKey)

				opts := pagerduty.ListOverridesOptions{
					Since: "2025-05-13",
					Until: "2025-05-27",
				}
				overrides, err := client.ListOverridesWithContext(ctx, pagerdutySchedule, opts)
				if err != nil {
					logrus.Fatalf("Failed to get overrides: %v", err)
				}
				logrus.Infof("Overrides: %+v", overrides)
			*/

			// create overrides: see:
			// API docs:    https://developer.pagerduty.com/api-reference/41d0a7c3c3a01-create-one-or-more-overrides
			// Go bindings: https://github.com/PagerDuty/go-pagerduty/blob/9831333ebe6bbe82a890c9aa0a9fd462a2a8d3e8/schedule.go#L243
			// CreateOverrideWithContext

		},
	}
}
