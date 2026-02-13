package view

// WinningOptions returns the set of option keys that share the highest vote total.
// Returns nil if there are no votes (max is 0).
func WinningOptions(totals map[string]int) map[string]bool {
	max := 0
	for _, n := range totals {
		if n > max {
			max = n
		}
	}
	if max == 0 {
		return nil
	}
	winners := make(map[string]bool)
	for opt, n := range totals {
		if n == max {
			winners[opt] = true
		}
	}
	return winners
}
