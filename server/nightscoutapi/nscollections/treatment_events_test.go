package nscollections

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/models/sitemodels"
)

// Every type of treatment keeps the kinds of event the site lets its group have, each a kind the API can write
func TestTreatmentTypeEvents(t *testing.T) {
	for _, eventType := range nightscoutmodels.TreatmentEventTypes {
		kinds, ok := sitemodels.EventGroupTypeKinds[eventType]
		require.True(t, ok, eventType)
		require.Len(t, treatmentTypeEvents[eventType], len(kinds), eventType)
		for _, event := range treatmentTypeEvents[eventType] {
			require.NotNil(t, event, eventType)
		}
	}
}
