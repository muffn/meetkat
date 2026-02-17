package view

import "meetkat/internal/poll"

// WinningOptions returns the set of option keys that share the highest yes-vote total.
// Returns nil if there are no votes (maxYes is 0).
func WinningOptions(totals map[string]poll.OptionTotal) map[string]bool {
	maxYes := 0
	for _, t := range totals {
		if t.Yes > maxYes {
			maxYes = t.Yes
		}
	}
	if maxYes == 0 {
		return nil
	}
	winners := make(map[string]bool)
	for opt, t := range totals {
		if t.Yes == maxYes {
			winners[opt] = true
		}
	}
	return winners
}
