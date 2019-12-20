package util

// MinInt returns the minimum integer value
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
