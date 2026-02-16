package view

import (
	"testing"

	"meetkat/internal/poll"
)

func TestWinningOptions(t *testing.T) {
	tests := []struct {
		name   string
		totals map[string]poll.OptionTotal
		want   map[string]bool
	}{
		{
			name:   "nil totals",
			totals: nil,
			want:   nil,
		},
		{
			name:   "all zeros",
			totals: map[string]poll.OptionTotal{"a": {Yes: 0}, "b": {Yes: 0}},
			want:   nil,
		},
		{
			name:   "single winner",
			totals: map[string]poll.OptionTotal{"a": {Yes: 3}, "b": {Yes: 1}, "c": {Yes: 2}},
			want:   map[string]bool{"a": true},
		},
		{
			name:   "tie",
			totals: map[string]poll.OptionTotal{"a": {Yes: 2}, "b": {Yes: 2}, "c": {Yes: 1}},
			want:   map[string]bool{"a": true, "b": true},
		},
		{
			name:   "all tied",
			totals: map[string]poll.OptionTotal{"a": {Yes: 1}, "b": {Yes: 1}, "c": {Yes: 1}},
			want:   map[string]bool{"a": true, "b": true, "c": true},
		},
		{
			name:   "maybe does not count for winner",
			totals: map[string]poll.OptionTotal{"a": {Yes: 1, Maybe: 5}, "b": {Yes: 2, Maybe: 0}},
			want:   map[string]bool{"b": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WinningOptions(tt.totals)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for k := range tt.want {
				if !got[k] {
					t.Errorf("expected %q in winners, got %v", k, got)
				}
			}
		})
	}
}
