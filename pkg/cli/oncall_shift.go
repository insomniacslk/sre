package cli

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/insomniacslk/sre/pkg/config"
	"github.com/sirupsen/logrus"
)

type ShiftType int

// FIXME: this only works in locations where the weekend is sat - sun
const (
	// Unknown shift type, the default when the structure is not populated yet or when the start time is past the end time
	ShiftTypeUndefined ShiftType = iota
	// a day in the range mon - fri
	ShiftTypeWeekday
	// a day in the range sat - sun
	ShiftTypeWeekend
	// a combination of weekdays and weekend days
	ShiftTypeMixed
)

type OncallShift struct {
	OncallPerson *config.OncallPerson
	Name         string
	StartTime    time.Time
	EndTime      time.Time
	// private fields
	shiftID   string
	shiftType ShiftType
}

// ID generates a SHA256 hash for this shift based on the start and end time.
// It should be unique unless uninitialized struct, shifts with the same start
// and end time, and SHA256 collisions.
func (s *OncallShift) ID() string {
	str := fmt.Sprintf("%s - %s",
		s.StartTime.Format(time.RFC822),
		s.EndTime.Format(time.RFC822),
	)
	h := sha256.New()
	h.Write([]byte(str))
	s.shiftID = string(h.Sum(nil))
	return s.shiftID
}

func isWeekend(t time.Time) bool {
	switch t.Weekday() {
	case time.Saturday, time.Sunday:
		return true
	default:
		return false
	}
}

// Type returns the oncall shift type
func (s *OncallShift) Type() ShiftType {
	var shiftType ShiftType
	cur := s.StartTime
loop:
	for {
		if cur.After(s.EndTime) {
			break
		}
		if isWeekend(cur) {
			// it's a weekend
			switch shiftType {
			case ShiftTypeWeekday, ShiftTypeMixed:
				shiftType = ShiftTypeMixed
				// if it is mixed we don't need to check any further day
				break loop
			default:
				shiftType = ShiftTypeWeekend
			}
		} else {
			// it's a weekday
			switch shiftType {
			case ShiftTypeWeekend, ShiftTypeMixed:
				shiftType = ShiftTypeMixed
				// if it is mixed we don't need to check any further day
				break loop
			default:
				shiftType = ShiftTypeWeekday
			}
		}
		cur = cur.AddDate(0, 0, 1)
	}
	s.shiftType = shiftType
	return shiftType
}

type OncallSchedule []OncallShift

// ShiftCandidate is a struct containing potential candidates for a specific shift
type ShiftCandidate struct {
	Shift           *OncallShift
	AvailablePeople []*config.OncallPerson
}

func GenerateSchedule(cfg *config.OncallGeneratorConfig) (*OncallShift, error) {
	logrus.Debugf("Desired shifts:\n")
	shiftCandidates := make(map[string][]ShiftCandidate, len(cfg.Shifts))
	for _, shift := range cfg.Shifts {
		logrus.Debugf("  %s\n    days: %v\n    start : %s\n    end   : %s\n", shift.Name, shift.Days, shift.StartTime, shift.EndTime)

		// validate weekday strings
		var weekdays []time.Weekday
		for _, day := range shift.Days {
			switch day {
			case time.Sunday.String():
				weekdays = append(weekdays, time.Sunday)
			case time.Monday.String():
				weekdays = append(weekdays, time.Monday)
			case time.Tuesday.String():
				weekdays = append(weekdays, time.Tuesday)
			case time.Wednesday.String():
				weekdays = append(weekdays, time.Wednesday)
			case time.Thursday.String():
				weekdays = append(weekdays, time.Thursday)
			case time.Friday.String():
				weekdays = append(weekdays, time.Friday)
			case time.Saturday.String():
				weekdays = append(weekdays, time.Saturday)
			default:
				return nil, fmt.Errorf("invalid weekday %q", day)
			}
		}

		// parse start and end time
		startTime, err := time.Parse("15:04 MST", shift.StartTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time %q: %w", shift.StartTime, err)
		}
		endTime, err := time.Parse("15:04 MST", shift.EndTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time %q: %w", shift.EndTime, err)
		}

		// get all the shift days in an array of time.Time
		now := time.Now()
		shiftDays := make([]OncallShift, 0, len(weekdays))
		for _, wd := range weekdays {
			daysUntilStart := (wd + 7 - now.Weekday()) % 7
			shiftDay := now.Add(time.Duration(daysUntilStart) * time.Hour * 24)
			shiftStart := time.Date(
				shiftDay.Year(),
				shiftDay.Month(),
				shiftDay.Day(),
				startTime.Hour(),
				startTime.Minute(),
				0,
				0,
				startTime.Location(),
			)
			shiftEnd := time.Date(
				shiftDay.Year(),
				shiftDay.Month(),
				shiftDay.Day(),
				endTime.Hour(),
				endTime.Minute(),
				0,
				0,
				endTime.Location(),
			)
			// if start > end it's because parsing into the location means that the shift overflows to the day after in that TZ,
			// so we add 1 day to get the accurate end time.
			if shiftStart.After(shiftEnd) {
				shiftEnd.AddDate(0, 0, 1)
			}

			shiftDays = append(shiftDays, OncallShift{
				Name:      shift.Name,
				StartTime: shiftStart,
				EndTime:   shiftEnd,
			})
		}
		logrus.Debugf("Shift days: %+v\n", shiftDays)

		// turn shiftDays into a map shift ID -> shift for easier lookup
		shiftsByID := make(map[string]OncallShift, len(shiftDays))
		for _, d := range shiftDays {
			if _, ok := shiftsByID[d.ID()]; ok {
				logrus.Panicf("Found duplicate shift with ID %q, this is probably a program bug", d.ID())
			}
			shiftsByID[d.ID()] = d
		}
		if _, ok := shiftCandidates[shift.Name]; !ok {
			shiftCandidates[shift.Name] = make([]ShiftCandidate, 0)
		}

		// TODO generalize constraints (hard and soft)
		// Create a mapping between the shift and the available people. The key is the output of OncallShift.ID()
		availability := make(map[string][]*config.OncallPerson, 0)

		// check the first hard constraint: earliest/latest hours in the user's timezone
		// FIXME: this algorithm is inefficient
		for _, shiftDay := range shiftDays {
			availablePeople := make([]*config.OncallPerson, 0)
			for _, member := range cfg.Members {
				logrus.Debugf("Checking if %s can cover the shift %s\n", member.Name, shift.Name)
				tz, err := time.LoadLocation(member.Constraints.Timezone)
				if err != nil {
					return nil, fmt.Errorf("invalid timezone %q for person %q: %w", member.Constraints.Timezone, member.Name, err)
				}
				// get shift day in the person's timezone
				start := shiftDay.StartTime.In(tz)
				end := shiftDay.EndTime.In(tz)
				if member.Name == "Christopher Wade" && shift.Name == "weekend_noram" {
					logrus.Infof("start: %s", start)
					logrus.Infof("end  : %s", end)
				}
				if start.Hour() < end.Hour() {
					// the shift is entirely contained in one day
					if start.Hour() < int(member.Constraints.EarliestOncallHour) {
						logrus.Debugf("Skipping person %q for shift %q: shift starts too early in tz %s, want at least %d, got %d", member.Name, shift.Name, tz, member.Constraints.EarliestOncallHour, start.Hour())
						continue
					}
					if end.Hour() > int(member.Constraints.LatestOncallHour) {
						logrus.Debugf("Skipping person %q for shift %q: shift ends too late in tz %s, want no later than %d, got %d", member.Name, shift.Name, tz, member.Constraints.LatestOncallHour, end.Hour())
						continue
					}
				} else {
					// the shift rolls over to the next day
					if start.Hour() < int(member.Constraints.EarliestOncallHour) {
						logrus.Debugf("Skipping person %q for shift %q: shift starts too early in tz %s, want at least %d, got %d", member.Name, shift.Name, tz, member.Constraints.EarliestOncallHour, start.Hour())
						continue
					}
					if end.Hour()+24 > int(member.Constraints.LatestOncallHour) {
						logrus.Debugf("Skipping person %q for shift %q: shift ends too late in tz %s, want no later than %d, got %d on the next day", member.Name, shift.Name, tz, member.Constraints.LatestOncallHour, end.Hour())
						continue
					}
				}
				availablePeople = append(availablePeople, &member)
				logrus.Debugf("Added %s to shift %s", member.Name, shiftDay.Name)
			}
			availability[shiftDay.ID()] = availablePeople
			shiftCandidate := ShiftCandidate{
				Shift:           &shiftDay,
				AvailablePeople: availablePeople,
			}
			shiftCandidates[shift.Name] = append(shiftCandidates[shift.Name], shiftCandidate)
		}

		// check the second constraint: public holidays. This is a soft constraint, there are cases where we simply don't
		// have anyone available (e.g. Christmas in NORAM), so one person will get the lucky day anyway.
		// TODO implement this
		// TODO implement the remaining constraints

	}
	// Print availability for the shifts
	shiftsWithoutOncall := 0
	for name, candidates := range shiftCandidates {
		fmt.Printf("Shift availability for shift %s:\n", name)
		if len(candidates) == 0 {
			fmt.Printf("    WARNING: no available people for this shift\n")
			shiftsWithoutOncall++
			continue
		}
		for _, candidate := range candidates {
			start, end := candidate.Shift.StartTime, candidate.Shift.EndTime
			fmt.Printf("  - Time: %s through %s\n", start.In(time.UTC).Format(time.RFC822), end.In(time.UTC).Format(time.RFC822))
			fmt.Printf("    Available people:\n")
			if len(candidate.AvailablePeople) == 0 {
				fmt.Printf("        WARNING: no available people for this shift day\n")
				shiftsWithoutOncall++
				continue
			}
			for _, p := range candidate.AvailablePeople {
				tz, err := time.LoadLocation(p.Constraints.Timezone)
				if err != nil {
					return nil, fmt.Errorf("invalid timezone %q for person %q: %w", p.Constraints.Timezone, p.Name, err)
				}
				fmt.Printf("        - %s <%s> from %s to %s\n", p.Name, p.Email, start.In(tz).Format(time.RFC822), end.In(tz).Format(time.RFC822))
			}
		}
	}
	if shiftsWithoutOncall > 0 {
		fmt.Printf("WARNING: found %d shift(s) without available oncalls\n", shiftsWithoutOncall)
	}
	return nil, fmt.Errorf("schedule generation not implemented yet")
}
