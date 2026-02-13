package view

import (
	"fmt"
	"testing"
)

// stubT returns a simple translation function that maps keys like
// "month.6" → "Jun", "weekday.1" → "Mon", etc.
func stubT() func(string, ...any) string {
	m := map[string]string{
		"month.1": "Jan", "month.2": "Feb", "month.3": "Mar",
		"month.4": "Apr", "month.5": "May", "month.6": "Jun",
		"month.7": "Jul", "month.8": "Aug", "month.9": "Sep",
		"month.10": "Oct", "month.11": "Nov", "month.12": "Dec",
		"weekday.0": "Sun", "weekday.1": "Mon", "weekday.2": "Tue",
		"weekday.3": "Wed", "weekday.4": "Thu", "weekday.5": "Fri",
		"weekday.6": "Sat",
	}
	return func(key string, args ...any) string {
		if v, ok := m[key]; ok {
			return v
		}
		return key
	}
}

func TestBuildDateHeaders(t *testing.T) {
	tFunc := stubT()

	tests := []struct {
		name    string
		options []string
		want    []HeaderGroup
	}{
		{
			name:    "empty list",
			options: nil,
			want:    nil,
		},
		{
			name:    "single date",
			options: []string{"2025-06-10"},
			want: []HeaderGroup{
				{Label: "Jun 2025", Colspan: 1, Columns: []HeaderColumn{
					{Raw: "2025-06-10", Label: "Tue 10"},
				}},
			},
		},
		{
			name:    "all dates same month",
			options: []string{"2025-06-10", "2025-06-11", "2025-06-12"},
			want: []HeaderGroup{
				{Label: "Jun 2025", Colspan: 3, Columns: []HeaderColumn{
					{Raw: "2025-06-10", Label: "Tue 10"},
					{Raw: "2025-06-11", Label: "Wed 11"},
					{Raw: "2025-06-12", Label: "Thu 12"},
				}},
			},
		},
		{
			name:    "multiple months",
			options: []string{"2025-06-10", "2025-06-11", "2025-07-01"},
			want: []HeaderGroup{
				{Label: "Jun 2025", Colspan: 2, Columns: []HeaderColumn{
					{Raw: "2025-06-10", Label: "Tue 10"},
					{Raw: "2025-06-11", Label: "Wed 11"},
				}},
				{Label: "Jul 2025", Colspan: 1, Columns: []HeaderColumn{
					{Raw: "2025-07-01", Label: "Tue 1"},
				}},
			},
		},
		{
			name:    "non-date options",
			options: []string{"Option A", "Option B"},
			want: []HeaderGroup{
				{Label: "", Colspan: 1, Columns: []HeaderColumn{{Raw: "Option A", Label: "Option A"}}},
				{Label: "", Colspan: 1, Columns: []HeaderColumn{{Raw: "Option B", Label: "Option B"}}},
			},
		},
		{
			name:    "mixed dates and non-dates no merging across gap",
			options: []string{"2025-06-10", "Option X", "2025-06-15"},
			want: []HeaderGroup{
				{Label: "Jun 2025", Colspan: 1, Columns: []HeaderColumn{
					{Raw: "2025-06-10", Label: "Tue 10"},
				}},
				{Label: "", Colspan: 1, Columns: []HeaderColumn{{Raw: "Option X", Label: "Option X"}}},
				{Label: "Jun 2025", Colspan: 1, Columns: []HeaderColumn{
					{Raw: "2025-06-15", Label: "Sun 15"},
				}},
			},
		},
		{
			name:    "non-consecutive same month stays separate",
			options: []string{"2025-06-10", "2025-07-01", "2025-06-20"},
			want: []HeaderGroup{
				{Label: "Jun 2025", Colspan: 1, Columns: []HeaderColumn{
					{Raw: "2025-06-10", Label: "Tue 10"},
				}},
				{Label: "Jul 2025", Colspan: 1, Columns: []HeaderColumn{
					{Raw: "2025-07-01", Label: "Tue 1"},
				}},
				{Label: "Jun 2025", Colspan: 1, Columns: []HeaderColumn{
					{Raw: "2025-06-20", Label: "Fri 20"},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildDateHeaders(tt.options, tFunc)

			if len(got) != len(tt.want) {
				t.Fatalf("got %d groups, want %d\n  got:  %s\n  want: %s",
					len(got), len(tt.want), fmtGroups(got), fmtGroups(tt.want))
			}

			for i := range tt.want {
				g, w := got[i], tt.want[i]
				if g.Label != w.Label || g.Colspan != w.Colspan || len(g.Columns) != len(w.Columns) {
					t.Errorf("group[%d] mismatch:\n  got:  %s\n  want: %s", i, fmtGroup(g), fmtGroup(w))
					continue
				}
				for j := range w.Columns {
					if g.Columns[j] != w.Columns[j] {
						t.Errorf("group[%d].Columns[%d] = %+v, want %+v", i, j, g.Columns[j], w.Columns[j])
					}
				}
			}
		})
	}
}

func fmtGroup(g HeaderGroup) string {
	return fmt.Sprintf("{Label:%q Colspan:%d Columns:%+v}", g.Label, g.Colspan, g.Columns)
}

func fmtGroups(gs []HeaderGroup) string {
	var s string
	for i, g := range gs {
		if i > 0 {
			s += ", "
		}
		s += fmtGroup(g)
	}
	return "[" + s + "]"
}
