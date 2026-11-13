package types

import (
	"cmp"
	"math"
)

// Duration is how long something lasts: in the whole minutes the site keeps, and to the millisecond when that is
// known. Clients send it as duration, in minutes, possibly with a fraction, as Loop sends a bolus's
// 1.6666666666666667, or as durationInMilliseconds, as AAPS sends it beside its whole minutes and reads it first, or
// both, and read either, so it is read from and written as both (SentDuration, DurationFields).
type Duration struct {
	Minutes int64
	// Ms is to the millisecond, nil when it is its minutes
	Ms *int64
	// wholeMinutes writes its duration in whole minutes, as API v3 does
	wholeMinutes bool
}

// SentDuration is a duration as a client sends it, for a model to read it from
type SentDuration struct {
	DurationMinutes        *FlexibleFloat64 `json:"duration"`
	DurationInMilliseconds *FlexibleInt64   `json:"durationInMilliseconds"`
}

// Duration is the one sent, nil when none is. Minutes sent are kept in whole minutes, cut, and to the millisecond
// when they have a fraction and no milliseconds are sent; milliseconds sent are kept as sent, and in the nearest
// minutes when no minutes are sent.
func (s SentDuration) Duration() *Duration {
	if s.DurationMinutes != nil {
		minutes := s.DurationMinutes.Float64()
		d := Duration{Minutes: int64(minutes), Ms: s.DurationInMilliseconds.Int64p()}
		if d.Ms == nil && minutes != math.Trunc(minutes) {
			d.Ms = new(int64(math.Round(minutes * 60_000)))
		}
		return &d
	}
	return DurationOfMs(s.DurationInMilliseconds.Int64p())
}

// DurationOfMinutes is a duration of whole minutes, nil when there are none
func DurationOfMinutes(minutes *int64) *Duration {
	if minutes == nil {
		return nil
	}
	return &Duration{Minutes: *minutes}
}

// DurationOfMs is a duration of milliseconds, in the nearest minutes, nil when there are none
func DurationOfMs(ms *int64) *Duration {
	if ms == nil {
		return nil
	}
	return &Duration{Minutes: (*ms + 30_000) / 60_000, Ms: ms}
}

// WithMs is the duration to the millisecond, the milliseconds kept with it, or of those milliseconds when it has no
// minutes
func (d *Duration) WithMs(ms *int64) *Duration {
	if d == nil || ms == nil {
		return cmp.Or(d, DurationOfMs(ms))
	}
	return &Duration{Minutes: d.Minutes, Ms: ms}
}

// MinutesP are its whole minutes, nil when there is none
func (d *Duration) MinutesP() *int64 {
	if d == nil {
		return nil
	}
	return &d.Minutes
}

// MinutesOrZero are its whole minutes, 0 when there is none
func (d *Duration) MinutesOrZero() int64 {
	if d == nil {
		return 0
	}
	return d.Minutes
}

// MsP are its milliseconds as kept, nil when they are its minutes
func (d *Duration) MsP() *int64 {
	if d == nil {
		return nil
	}
	return d.Ms
}

// Milliseconds are the milliseconds kept while they are within a minute of its minutes, else its minutes', as its
// minutes may be changed on the site
func (d *Duration) Milliseconds() int64 {
	if d.Ms != nil && max(*d.Ms-d.Minutes*60_000, d.Minutes*60_000-*d.Ms) < 60_000 {
		return *d.Ms
	}
	return d.Minutes * 60_000
}

// InWholeMinutes is it written in whole minutes, as API v3 writes it
func (d *Duration) InWholeMinutes() *Duration {
	if d == nil {
		return nil
	}
	whole := *d
	whole.wholeMinutes = true
	return &whole
}

// DurationFields are a duration as it is written, for a model to embed
type DurationFields struct {
	DurationMinutes        any    `json:"duration,omitempty"`
	DurationInMilliseconds *int64 `json:"durationInMilliseconds,omitempty"`
}

// Fields are it as it is written: in both minutes and milliseconds, as a client may read either. Its minutes have the
// fraction of its milliseconds, as Loop's NightscoutKit and LoopFollow read them, unless it is written in whole
// minutes, as AAPS reads them as a whole number and drops a treatment with a fraction.
func (d *Duration) Fields() DurationFields {
	if d == nil {
		return DurationFields{}
	}
	ms := d.Milliseconds()
	fields := DurationFields{DurationMinutes: d.Minutes, DurationInMilliseconds: &ms}
	if !d.wholeMinutes && ms%60_000 != 0 && ms/60_000 == d.Minutes {
		fields.DurationMinutes = float64(ms) / 60_000
	}
	return fields
}
