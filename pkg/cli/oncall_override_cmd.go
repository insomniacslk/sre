package cli

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/insomniacslk/sre/pkg/ansi"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	flagOncallOverrideStart string
	flagOncallOverrideEnd   string
	flagOncallOverrideYes   bool
)

func init() {
	OncallCmd.AddCommand(OncallOverrideCmd)
	OncallOverrideCmd.PersistentFlags().StringVarP(&flagOncallOverrideStart, "start-time", "s", "", "Start time of the override. Accepted formats: RFC3999, RFC3999Nano, RFC822, RFC822Z, Unix time")
	OncallOverrideCmd.PersistentFlags().StringVarP(&flagOncallOverrideEnd, "end-time", "e", "", "End time of the override. Accepted formats: RFC3339, RFC3339Nano, RFC822, RFC822Z, Unix time")
	OncallOverrideCmd.PersistentFlags().BoolVarP(&flagOncallOverrideYes, "yes", "y", false, "Do not ask for confirmation before creating the overrride")
}

func parseOverrideTimeString(s string) (*time.Time, error) {
	for _, timefmt := range []string{time.RFC3339, time.RFC3339Nano, time.RFC822, time.RFC822Z} {
		t, err := time.Parse(timefmt, s)
		if err != nil {
			continue
		}
		return &t, nil
	}
	// try unix time
	ut, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		t := time.Unix(ut, 0)
		return &t, nil
	}
	return nil, fmt.Errorf("time string %q not in any supported format", s)
}

var OncallOverrideCmd = &cobra.Command{
	Use:     "override",
	Aliases: []string{"o", "ov", "over"},
	Short:   "Create an override in the oncall schedule (PagerDuty)",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.Debugf("Running override command")
		ctx := context.Background()
		cfg, err := GetConfig()
		if err != nil {
			return err
		}

		userID := args[0]
		// validate flags
		if flagOncallOverrideStart == "" {
			return fmt.Errorf("start time not specified")
		}
		start, err := parseOverrideTimeString(flagOncallOverrideStart)
		if err != nil {
			return fmt.Errorf("failed to parse start time: %w", err)
		}
		if flagOncallOverrideEnd == "" {
			return fmt.Errorf("end time not specified")
		}
		end, err := parseOverrideTimeString(flagOncallOverrideEnd)
		if err != nil {
			return fmt.Errorf("failed to parse end time: %w", err)
		}
		if !end.After(*start) {
			return fmt.Errorf("end time must be after start time")
		}
		fmt.Printf("Creating override for user %q from %s to %s (duration: %s)\n", userID, start, end, end.Sub(*start))

		client := pagerduty.NewClient(cfg.PagerDuty.UserToken)

		scheduleID := cfg.Oncall.DefaultSchedule
		if len(args) > 1 {
			scheduleID = args[1]
		}
		if scheduleID == "" {
			logrus.Fatalf("No schedule ID specified")
		}

		// search for the specified user
		uopts := pagerduty.ListUsersOptions{
			Query:    userID,
			Includes: []string{"contact_methods"},
		}
		resp, err := client.ListUsersWithContext(ctx, uopts)
		if err != nil {
			return fmt.Errorf("failed to find users matching %q: %w", userID, err)
		}
		if len(resp.Users) == 0 {
			return fmt.Errorf("no user found matching %q", userID)
		}
		if len(resp.Users) > 1 {
			fmt.Printf("Found more than one user matching %q, try using a narrower search\n", userID)
			for idx, u := range resp.Users {
				fmt.Printf("%d)  %s <%s> %s\n", idx+1, u.Name, u.Email, u.JobTitle)
			}
			os.Exit(1)
		}
		user := resp.Users[0]

		// search for a schedule
		log.Printf("Searching schedules matching %q", scheduleID)
		sopts := pagerduty.GetScheduleOptions{
			Since:    start.Format(time.RFC3339),
			Until:    end.Format(time.RFC3339),
			TimeZone: cfg.Timezone,
		}
		sched, err := client.GetScheduleWithContext(ctx, scheduleID, sopts)
		if err != nil {
			return fmt.Errorf("failed to get schedule with ID %q: %w", scheduleID, err)
		}
		currentOncalls := []pagerduty.APIObject{}
		fmt.Printf("%s\n", ansi.Bold(sched.Name))
		fmt.Printf("    Summary: %s\n", ansi.ToURL(sched.Summary, sched.HTMLURL))
		fmt.Printf("    Description: %s\n", sched.Description)
		fmt.Println()
		fmt.Printf("%s:\n", ansi.Bold("Current oncall(s)"))
		for _, entry := range sched.FinalSchedule.RenderedScheduleEntries {
			start, err := time.Parse(time.RFC3339, entry.Start)
			if err != nil {
				log.Fatalf("Start time %q is not in RFC3339 format: %v", entry.Start, err)
			}
			end, err := time.Parse(time.RFC3339, entry.End)
			if err != nil {
				log.Fatalf("End time %q is not in RFC3339 format: %v", entry.End, err)
			}
			timeFmt := "Mon 02 Jan 2006 15:04:05 MST"
			fmt.Printf("    %s - %s (%s)\t%+v\n",
				start.Format(timeFmt),
				end.Format(timeFmt),
				end.Sub(start),
				ansi.ToURL(entry.User.Summary, entry.User.HTMLURL),
			)
			currentOncalls = append(currentOncalls, entry.User)
		}
		fmt.Println()

		// check that this user isn't already oncall at that time
		if len(currentOncalls) == 1 && currentOncalls[0].ID == user.ID {
			fmt.Printf(ansi.Bold(fmt.Sprintf("The user %s <%s> is already oncall at that time, aborting\n", user.Name, user.Email)))
			os.Exit(1)
		}

		//
		if !flagOncallOverrideYes {
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("Do you want to override the above schedule with the user %s <%s> %s ? [yN] ", user.Name, user.Email, user.JobTitle)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" {
				fmt.Printf("\nAborting\n")
				os.Exit(0)
			}
		}

		// create the override
		fmt.Printf("Overriding schedule with user %s <%s>\n", user.Name, user.Email)

		override := pagerduty.Override{
			User:  user.APIObject,
			Start: start.Format(time.RFC3339),
			End:   end.Format(time.RFC3339),
		}
		_, err = client.CreateOverrideWithContext(ctx, scheduleID, override)
		if err != nil {
			return fmt.Errorf("Override creation failed: %w", err)
		}
		fmt.Printf("Override created\n")

		return nil
	},
}
