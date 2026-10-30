package glucose

import "testing"

func TestPercentile(t *testing.T) {
	tests := []struct {
		name   string
		sorted []int64
		p      float64
		want   float64
	}{
		{name: "no values is 0", sorted: nil, p: 50, want: 0},
		{name: "one value is every percentile", sorted: []int64{120}, p: 5, want: 120},
		{name: "the 0th is the lowest", sorted: []int64{80, 100, 200}, p: 0, want: 80},
		{name: "the 100th is the highest", sorted: []int64{80, 100, 200}, p: 100, want: 200},
		{name: "the median of an odd count is the middle", sorted: []int64{80, 100, 200}, p: 50, want: 100},
		{name: "the median of an even count is between the middle two", sorted: []int64{80, 100, 120, 200}, p: 50, want: 110},
		{name: "a percentile between values is linear between them", sorted: []int64{100, 200}, p: 25, want: 125},
		{name: "the 95th of 0 to 100 is 95", sorted: func() []int64 {
			values := make([]int64, 101)
			for i := range values {
				values[i] = int64(i)
			}
			return values
		}(), p: 95, want: 95},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Percentile(tt.sorted, tt.p); got != tt.want {
				t.Errorf("percentile %v of %v is %v, want %v", tt.p, tt.sorted, got, tt.want)
			}
		})
	}
}
