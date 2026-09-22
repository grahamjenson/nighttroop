package nighttroop

import "nighttroop.t1d.nz/utils/consts"

type GlucoseReading struct {
	ValueStr string
	TimeAgo  string
	Diff     string
}

type Alarm struct {
	Name string
	Icon string
}

type Notification struct {
	Value   string
	Icon    string
	Tooltip string
}

type GlucoseReadingType int64

const (
	GlucoseReadingTypeSGV GlucoseReadingType = iota
	GlucoseReadingTypeMBG
)

type InsulinType int64

const (
	InsulinTypeBolus InsulinType = iota
	InsulinTypeSMB
)

type TimeInfo struct {
	Timestamp int64
	TimeAgo   string

	GlucoseReading     *int64
	GlucoseReadingType *GlucoseReadingType

	Insulin     *int64
	InsulinType *InsulinType
	Carbs       *int64

	BasalRate  *int64
	HighTarget *int64
	LowTarget  *int64
}

type Dashboard struct {
	Units   *consts.Units
	Glucose *GlucoseReading
	Alarm   *Alarm

	Alerts        []Notification
	Notifications []Notification

	// For the graph
	Last24Hours []TimeInfo

	IOB            *int64
	COB            *int64
	ProfilePercent *int64
}
