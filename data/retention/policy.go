// Package retention is a proposal, not yet used by the server: how a site's old data becomes small without the
// person losing what they look back for. See docs/retention.md.
//
// A site's data is in two forms. Hot rows are what the tables hold today: one row a record, indexed, as the apps sent
// it, for the days the apps and pages read row by row. A cold day is everything of one kind for one day in one row: a
// signal packed on its time grid (grid.go), or a day's events as a short list. Sealing a day turns its hot rows into
// cold days, by the policy of each table, and deletes the rows.
package retention

import "time"

// DefaultHotDays is how many days of a site are hot rows unless RETENTION_HOT_DAYS says another number; a site over
// its size limit has fewer, and never fewer than MinHotDays
const (
	DefaultHotDays = 14
	MinHotDays     = 3
)

// Shape is the shape a table's rows take in a cold day, which decides what sealing a day does with them
type Shape string

const (
	// Signal is a number a machine writes on a clock, such as a CGM's glucose: its rows become series on a grid
	Signal Shape = "signal"
	// Event is what a person did or decided, a few a day, such as a meal: its rows become the day's list, each
	// with the columns the policy keeps
	Event Shape = "event"
	// Telemetry is what a device says about itself, such as a loop's status: its numbers become series, and its
	// rows are deleted
	Telemetry Shape = "telemetry"
	// Dictionary is a value many records repeat, such as a profile: kept once, by its hash, while a record refers to it
	Dictionary Shape = "dictionary"
	// Config is what a person manages as a list, such as food presets: never sealed
	Config Shape = "config"
	// Derived is built from the other tables, and can be built again from hot rows and cold days: a cache
	Derived Shape = "derived"
)

// Series is a number through the day on a grid: each step of the grid is one value, or none
type Series struct {
	Name string
	// Step is the grid's step. A cold day is a UTC day, so every day has the same steps, whatever the site's
	// timezone is or becomes; a day in the site's timezone is read from the two UTC days it is in.
	Step time.Duration
	// Scale is what a value is multiplied by to be kept as a whole number: 1 keeps whole mg/dL, 100 keeps 0.01 U
	Scale float64
	// Sum is whether a step holds the total of its rows (insulin delivered); else it holds the newest (glucose)
	Sum bool
}

// Policy is what sealing a day does with one table's rows of that day
type Policy struct {
	Shape Shape
	// Series are the series its rows are folded into
	Series []Series
	// Keep are the columns of each row that the day's list keeps, for an Event. A column not here is lost.
	Keep []string
	// Where narrows the rows the day's list keeps to those a person made; the others are only folded into Series
	Where string
	// Lost says, for the person, what cannot be read back once a day is sealed
	Lost string
}

const (
	fiveMinutes = 5 * time.Minute
	hour        = time.Hour
)

// who, when and why are what every event keeps
var common = []string{"id", "occurred_at", "utc_offset_minutes", "notes", "reason", "created_by_identity_name"}

func keep(columns ...string) []string { return append(append([]string{}, common...), columns...) }

// Policies are every table of a site's DB. A table with no policy fails TestEveryTableHasAPolicy, so a new table
// cannot be added without saying how it stays small.
var Policies = map[string]Policy{
	"glucose_readings": {Shape: Signal,
		Series: []Series{{Name: "glucose", Step: fiveMinutes, Scale: 1}},
		Keep:   keep("mgdl", "reading_type", "measurement_method"), Where: "reading_type = 'mbg' OR notes IS NOT NULL",
		Lost: "A CGM reading's exact second, its raw signal, and which app sent it. With two sources, the readings of the source not used."},
	"insulin_doses": {Shape: Event,
		Series: []Series{{Name: "insulin_automatic", Step: hour, Scale: 1000, Sum: true}, {Name: "insulin_manual", Step: hour, Scale: 1000, Sum: true}},
		Keep:   keep("dose", "dose_type", "extended_dose", "extended_duration_ms", "is_basal_insulin", "insulin_type"), Where: "automatic IS NOT 1",
		Lost: "Each automatic dose's time and size inside its hour."},
	"temp_basals": {Shape: Signal,
		Series: []Series{{Name: "basal_delivered", Step: hour, Scale: 1000, Sum: true}, {Name: "basal_suspended_minutes", Step: hour, Scale: 1, Sum: true}},
		Lost:   "Each temp basal's start, rate and duration. The basal delivered in each hour is kept."},
	"food_intakes":     {Shape: Event, Series: []Series{{Name: "carbs", Step: hour, Scale: 10, Sum: true}}, Keep: keep("carbs", "name", "food_type", "absorption_time_minutes", "fat", "protein", "fibre", "energy")},
	"activities":       {Shape: Event, Keep: keep("name", "activity_type", "duration_minutes", "calories", "steps", "heart_rate")},
	"temp_targets":     {Shape: Event, Keep: keep("low_target_mgdl", "high_target_mgdl", "duration_minutes")},
	"profile_switches": {Shape: Event, Keep: keep("name", "multiplier", "timeshift", "duration_minutes", "profile_id"), Lost: "A switch to the profile already in force. profile_id is new: the hash of the profile, kept once as a dictionary row."},
	"notification_events": {Shape: Event, Keep: keep("notification_type", "title", "severity", "content"), Where: "created_by_identity_id is a person's",
		Lost: "What a device announced."},
	"event_groups": {Shape: Event, Keep: []string{"group_type", "entered_by", "duration_minutes", "sensor_code", "transmitter_id"},
		Lost: "The app's own IDs of a treatment (uuid, sent_hash, pump_id, sync_identifier) and its bolus calculator. A group's kept columns move onto its event; a group with no event (a site change) is an event."},
	"device_reports": {Shape: Telemetry,
		Series: []Series{{Name: "iob", Step: fiveMinutes, Scale: 100}, {Name: "iob_basal", Step: fiveMinutes, Scale: 100}, {Name: "cob", Step: fiveMinutes, Scale: 10},
			{Name: "pump_reservoir", Step: hour, Scale: 10}, {Name: "pump_battery", Step: hour, Scale: 1}, {Name: "phone_battery", Step: hour, Scale: 1}},
		Lost: "The loop's reasons and predictions, and each report."},
	"loop_settings":    {Shape: Dictionary},
	"profile_presets":  {Shape: Config},
	"food_presets":     {Shape: Config},
	"general_settings": {Shape: Config},
	"summary_period":   {Shape: Derived},
}
