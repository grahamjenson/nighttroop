package retention

import (
	"encoding/json"
	"errors"
	"math"
)

// gridVersion is the "v" of a packed grid, so the packing can change and old days still be read
const gridVersion = 1

// grid is a day of one series as its cold day holds it, as JSON: D has one entry a step of the grid, null where there
// is no value, and else the value's difference from the value before it, as a whole number
type grid struct {
	V int      `json:"v"`
	D []*int64 `json:"d"`
}

// PackGrid packs a day of one series as JSON: values has one entry a step of the grid, NaN where there is none. Each
// value is kept as a whole number, value*scale, and written as its difference from the one before: a CGM's glucose
// changes by a few mg/dL in 5 minutes, so most entries are one or two characters. It is JSON, not bytes, so the
// column is checked with json_valid, as every JSON column is, and a person can read it.
func PackGrid(values []float64, scale float64) string {
	packed, last := grid{V: gridVersion, D: make([]*int64, len(values))}, int64(0)
	for i, value := range values {
		if math.IsNaN(value) {
			continue
		}
		whole := int64(math.Round(value * scale))
		packed.D[i], last = new(whole-last), whole
	}
	body, _ := json.Marshal(packed)
	return string(body)
}

// maxSteps is the most steps a day's grid has: one a minute
const maxSteps = 24 * 60

// UnpackGrid is the values PackGrid packed, NaN where there is none
func UnpackGrid(packed string, scale float64) ([]float64, error) {
	var unpacked grid
	if err := json.Unmarshal([]byte(packed), &unpacked); err != nil {
		return nil, err
	}
	if unpacked.V != gridVersion || len(unpacked.D) > maxSteps {
		return nil, errors.New("retention: not a packed grid of a known version")
	}
	values, last := make([]float64, len(unpacked.D)), int64(0)
	for i, delta := range unpacked.D {
		values[i] = math.NaN()
		if delta != nil {
			last += *delta
			values[i] = float64(last) / scale
		}
	}
	return values, nil
}

// Stats are what is known of a set of values that can be added to another set's: the stats of a week are its days'
// merged, and a year's its weeks', with no row read again. Bins counts the values in each BinWidth from 0, so the time
// in any range, and any percentile, can be read from it later, for a range the person had not yet chosen.
type Stats struct {
	Count      int64
	Sum, SumSq float64
	Min, Max   float64
	Bins       []int64
}

// BinWidth is the width of each of Stats' bins: 5 mg/dL for glucose
const BinWidth = 5

// Add counts one value
func (s *Stats) Add(value float64) {
	if s.Count == 0 || value < s.Min {
		s.Min = value
	}
	if s.Count == 0 || value > s.Max {
		s.Max = value
	}
	s.Count, s.Sum, s.SumSq = s.Count+1, s.Sum+value, s.SumSq+value*value
	bin := int(math.Max(value, 0) / BinWidth)
	for len(s.Bins) <= bin {
		s.Bins = append(s.Bins, 0)
	}
	s.Bins[bin]++
}

// Merge adds another set's stats to these
func (s *Stats) Merge(other Stats) {
	if other.Count == 0 {
		return
	}
	if s.Count == 0 || other.Min < s.Min {
		s.Min = other.Min
	}
	if s.Count == 0 || other.Max > s.Max {
		s.Max = other.Max
	}
	s.Count, s.Sum, s.SumSq = s.Count+other.Count, s.Sum+other.Sum, s.SumSq+other.SumSq
	for len(s.Bins) < len(other.Bins) {
		s.Bins = append(s.Bins, 0)
	}
	for i, count := range other.Bins {
		s.Bins[i] += count
	}
}

// Mean is the mean of the values, and SD their standard deviation
func (s Stats) Mean() float64 { return s.Sum / float64(s.Count) }

func (s Stats) SD() float64 {
	return math.Sqrt(math.Max(s.SumSq/float64(s.Count)-s.Mean()*s.Mean(), 0))
}

// PercentBetween is the percent of the values from low up to, not including, high, to the width of a bin
func (s Stats) PercentBetween(low, high float64) float64 {
	var count int64
	for bin := int(low / BinWidth); bin < len(s.Bins) && float64(bin*BinWidth) < high; bin++ {
		count += s.Bins[bin]
	}
	return 100 * float64(count) / float64(s.Count)
}
