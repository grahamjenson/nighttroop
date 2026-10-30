package glucose

// Percentile is the p-th percentile, 0 to 100, of sorted values, linearly between the two values either side of it
func Percentile(sorted []int64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	rank := p / 100 * float64(len(sorted)-1)
	low := int(rank)
	if low+1 >= len(sorted) {
		return float64(sorted[len(sorted)-1])
	}
	return float64(sorted[low]) + (rank-float64(low))*float64(sorted[low+1]-sorted[low])
}
