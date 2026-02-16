package view

import "meetkat/internal/poll"

// WinningOptions returns the set of option keys that share the highest yes-vote total.
// Returns nil if there are no votes (max is 0).
func WinningOptions(totals map[string]poll.OptionTotal) map[string]bool {
	max := 0
	for _, t := range totals {
		if t.Yes > max {
			max = t.Yes
		}
	}
	if max == 0 {
		return nil
	}
	winners := make(map[string]bool)
	for opt, t := range totals {
		if t.Yes == max {
			winners[opt] = true
		}
	}
	return winners
}
