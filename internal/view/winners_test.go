package view

import (
	"testing"
)

func TestWinningOptions(t *testing.T) {
	tests := []struct {
		name   string
		totals map[string]int
		want   map[string]bool
	}{
		{
			name:   "nil totals",
			totals: nil,
			want:   nil,
		},
		{
			name:   "all zeros",
			totals: map[string]int{"a": 0, "b": 0},
			want:   nil,
		},
		{
			name:   "single winner",
			totals: map[string]int{"a": 3, "b": 1, "c": 2},
			want:   map[string]bool{"a": true},
		},
		{
			name:   "tie",
			totals: map[string]int{"a": 2, "b": 2, "c": 1},
			want:   map[string]bool{"a": true, "b": true},
		},
		{
			name:   "all tied",
			totals: map[string]int{"a": 1, "b": 1, "c": 1},
			want:   map[string]bool{"a": true, "b": true, "c": true},
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
