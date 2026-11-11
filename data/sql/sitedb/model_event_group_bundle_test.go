package sitedb_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
)

// fill gives every field of v, a pointer to a struct, a value of its type, each field's different from the last, so
// a field copied into another is seen
func fill(v any) {
	s := reflect.ValueOf(v).Elem()
	for i := range s.NumField() {
		field := s.Field(i)
		target := field
		if field.Kind() == reflect.Pointer {
			field.Set(reflect.New(field.Type().Elem()))
			target = field.Elem()
		}
		switch {
		case target.Type() == reflect.TypeFor[consts.Units]():
			// Units are read as one of the two the DB's CHECK allows
			target.SetString(string(consts.UnitsMMOLL))
		case target.Kind() == reflect.String:
			target.SetString("text " + s.Type().Field(i).Name)
		case target.Kind() == reflect.Int64:
			target.SetInt(int64(1000 + i))
		case target.Kind() == reflect.Float64:
			target.SetFloat(1.5 + float64(i))
		case target.Kind() == reflect.Bool:
			target.SetBool(true)
		default:
			panic("fill cannot set " + target.Kind().String())
		}
	}
}

// fullRow is an event group row with each of its events' every column set
func fullRow() sitedb.GetEventGroupRow {
	var row sitedb.GetEventGroupRow
	fill(&row.EventGroup)
	fill(&row.GlucoseReadingsVw)
	fill(&row.InsulinDosesVw)
	fill(&row.FoodIntakesVw)
	fill(&row.TempBasalsVw)
	fill(&row.TempTargetsVw)
	fill(&row.ProfileSwitchesVw)
	fill(&row.ActivitiesVw)
	fill(&row.NotificationEventsVw)
	return row
}

// A bundle has each event its row joined, every column as the row has it, and none it did not join
func TestGetEventGroupToEventGroupsBundle(t *testing.T) {
	row := fullRow()
	bundle := sitedb.GetEventGroupToEventGroupsBundle(row)
	require.Equal(t, row.EventGroup, bundle.EventGroup)
	// Each event is its view's row, as the JSON of one read as the other was
	pairs := []struct {
		view, event any
	}{
		{row.GlucoseReadingsVw, bundle.GlucoseReading},
		{row.InsulinDosesVw, bundle.InsulinDose},
		{row.FoodIntakesVw, bundle.FoodIntake},
		{row.TempBasalsVw, bundle.TempBasal},
		{row.TempTargetsVw, bundle.TempTarget},
		{row.ProfileSwitchesVw, bundle.ProfileSwitch},
		{row.ActivitiesVw, bundle.Activity},
		{row.NotificationEventsVw, bundle.NotificationEvent},
	}
	for _, pair := range pairs {
		require.NotNil(t, pair.event)
		want := reflect.New(reflect.TypeOf(pair.event).Elem())
		data, err := json.Marshal(pair.view)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, want.Interface()))
		require.Equal(t, want.Interface(), pair.event, "%T", pair.view)
	}

	// A row that joined no events has none
	empty := sitedb.GetEventGroupToEventGroupsBundle(sitedb.GetEventGroupRow{EventGroup: row.EventGroup})
	require.Empty(t, empty.GetAllSubEvents())
}

// A meal bolus's group, its insulin dose and food intake, as each treatment a search finds is made a bundle
func BenchmarkGetEventGroupToEventGroupsBundle(b *testing.B) {
	full := fullRow()
	row := sitedb.GetEventGroupRow{EventGroup: full.EventGroup, InsulinDosesVw: full.InsulinDosesVw, FoodIntakesVw: full.FoodIntakesVw}
	for b.Loop() {
		sitedb.GetEventGroupToEventGroupsBundle(row)
	}
}
