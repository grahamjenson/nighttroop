package retention

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"slices"
	"time"
)

// Version is the version of a cold day's bytes this code writes, a day's first byte. A day says its own, so the
// form can change and the days packaged years ago still be read: add a case to UnpackGrid, never change what an old
// version means.
const Version = 1

// shapeBytes are the shapes a cold day can say it is, as its second byte
var shapeBytes = map[Shape]byte{SeriesShape: 1, EventsShape: 2}

// The third byte of a series says how its values were packed, as each day is packed the way that makes it smallest
const (
	// secondOrder is that each value is written as the change of its change, which is smaller for a smooth series
	secondOrder = 1 << iota
	// deflated is that the values are DEFLATE compressed
	deflated
)

// day is how long a cold day is
const day = 24 * time.Hour

// maxSteps is the most steps a day's grid has: one a minute
const maxSteps = int(day / time.Minute)

// Empty is a day of the series with no values: NaN at every step
func (s Series) Empty() []float64 {
	values := make([]float64, day/s.Step)
	for i := range values {
		values[i] = math.NaN()
	}
	return values
}

// PackGrid packs a day of a series as bytes: values has one entry a step, NaN where there is none. The bytes are
// the version, the shape, how the values were packed, then the series' step in seconds, its scale and how many
// steps, so a day is read as it was written whatever its series' policy has since become. Then the values: a bit a
// step for whether it has one, and each value, as a whole number value*Scale, as its difference from the one
// before, or the difference of that difference, in as few bytes as it needs, and DEFLATE compressed. Each way is
// tried and the smallest kept: a CGM's glucose changes by a few mg/dL in 5 minutes, and a day of it is about 170
// bytes.
func PackGrid(series Series, values []float64) []byte {
	present := make([]byte, (len(values)+7)/8)
	orders := [2][]byte{}
	var last, lastDelta int64
	for i, value := range values {
		if math.IsNaN(value) {
			continue
		}
		present[i/8] |= 1 << (i % 8)
		whole := int64(math.Round(value * series.Scale))
		delta := whole - last
		orders[0], orders[1] = binary.AppendVarint(orders[0], delta), binary.AppendVarint(orders[1], delta-lastDelta)
		last, lastDelta = whole, delta
	}
	var smallest []byte
	var how byte
	for order, deltas := range orders {
		plain := append(slices.Clone(present), deltas...)
		var compressed bytes.Buffer
		writer, _ := flate.NewWriter(&compressed, flate.BestCompression)
		_, _ = writer.Write(plain)
		_ = writer.Close()
		for flag, packed := range map[byte][]byte{0: plain, deflated: compressed.Bytes()} {
			if smallest == nil || len(packed) < len(smallest) {
				smallest, how = packed, flag|byte(order)*secondOrder
			}
		}
	}
	head := []byte{Version, shapeBytes[SeriesShape], how}
	for _, number := range []uint64{uint64(series.Step / time.Second), uint64(series.Scale), uint64(len(values))} {
		head = binary.AppendUvarint(head, number)
	}
	return append(head, smallest...)
}

// errNotASeries is that bytes are not a series of a version this code reads
var errNotASeries = errors.New("retention: not a series of a known version")

// Header is what a cold day's first bytes say it is: its version and shape, and for a series how its values were
// packed, its step and scale, and how many steps it has
type Header struct {
	Version     int
	Shape       Shape
	SecondOrder bool
	Deflated    bool
	Step        time.Duration
	Scale       float64
	Steps       int
}

// UnpackGrid is the values PackGrid packed, NaN where there is none, by the version, scale and packing the day says
func UnpackGrid(packed []byte) ([]float64, error) {
	_, values, err := ReadGrid(packed)
	return values, err
}

// ReadGrid is a packed series with what its first bytes say it is, to look at a cold day as it is kept
func ReadGrid(packed []byte) (Header, []float64, error) {
	if len(packed) < 3 || packed[0] != Version || packed[1] != shapeBytes[SeriesShape] || packed[2]&^(secondOrder|deflated) != 0 {
		return Header{}, nil, errNotASeries
	}
	how, rest := packed[2], packed[3:]
	var head [3]uint64
	for i := range head {
		number, read := binary.Uvarint(rest)
		if read <= 0 {
			return Header{}, nil, errNotASeries
		}
		head[i], rest = number, rest[read:]
	}
	header := Header{Version: Version, Shape: SeriesShape, SecondOrder: how&secondOrder != 0, Deflated: how&deflated != 0,
		Step: time.Duration(min(head[0], uint64(day/time.Second))) * time.Second, Scale: float64(head[1]), Steps: int(min(head[2], uint64(maxSteps)+1))}
	if header.Scale == 0 || header.Steps > maxSteps {
		return Header{}, nil, errNotASeries
	}
	if header.Deflated {
		// A day is at most a bit and 10 bytes a step, so no more is read of bytes that say they are more
		plain, err := io.ReadAll(io.LimitReader(flate.NewReader(bytes.NewReader(rest)), int64(maxSteps*11)))
		if err != nil {
			return Header{}, nil, errNotASeries
		}
		rest = plain
	}
	bitmap := (header.Steps + 7) / 8
	if len(rest) < bitmap {
		return Header{}, nil, errNotASeries
	}
	present, deltas := rest[:bitmap], rest[bitmap:]
	values := make([]float64, header.Steps)
	var last, lastDelta int64
	for i := range values {
		values[i] = math.NaN()
		if present[i/8]&(1<<(i%8)) == 0 {
			continue
		}
		delta, read := binary.Varint(deltas)
		if read <= 0 {
			return Header{}, nil, errNotASeries
		}
		if header.SecondOrder {
			delta += lastDelta
		}
		deltas, last, lastDelta = deltas[read:], last+delta, delta
		values[i] = float64(last) / header.Scale
	}
	return header, values, nil
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
