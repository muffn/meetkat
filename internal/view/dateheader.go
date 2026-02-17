package view

import (
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

// HeaderColumn represents a single column in the vote table header.
type HeaderColumn struct {
	Raw   string // original option string (used for form keys)
	Label string // display: "Mon 10" for dates, raw string for non-dates
}

// HeaderGroup represents a group of consecutive columns sharing the same month.
type HeaderGroup struct {
	Label   string // "Jun 2025" for dates, "" for non-dates
	Colspan int
	Columns []HeaderColumn
}

// BuildDateHeaders groups poll options into header groups for display.
// Options that parse as YYYY-MM-DD are grouped by consecutive same-month runs.
// Non-date options become standalone groups with an empty Label.
func BuildDateHeaders(options []string, tFunc func(string, ...any) string) []HeaderGroup {
	if len(options) == 0 {
		return nil
	}

	var groups []HeaderGroup

	for _, opt := range options {
		t, err := time.Parse(dateLayout, opt)
		if err != nil {
			// Non-date option: standalone group
			groups = append(groups, HeaderGroup{
				Label:   "",
				Colspan: 1,
				Columns: []HeaderColumn{{Raw: opt, Label: opt}},
			})
			continue
		}

		monthKey := fmt.Sprintf("month.%d", int(t.Month()))
		weekdayKey := fmt.Sprintf("weekday.%d", int(t.Weekday()))
		monthLabel := fmt.Sprintf("%s %d", tFunc(monthKey), t.Year())
		dayLabel := fmt.Sprintf("%s %d", tFunc(weekdayKey), t.Day())

		col := HeaderColumn{Raw: opt, Label: dayLabel}

		// Try to merge with the last group if it has the same month label
		if len(groups) > 0 {
			last := &groups[len(groups)-1]
			if last.Label == monthLabel {
				last.Columns = append(last.Columns, col)
				last.Colspan++
				continue
			}
		}

		groups = append(groups, HeaderGroup{
			Label:   monthLabel,
			Colspan: 1,
			Columns: []HeaderColumn{col},
		})
	}

	return groups
}
