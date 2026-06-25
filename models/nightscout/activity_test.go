package nightscout

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	fixtures_tests "nighttroop.t1d.nz/models/nightscout/fixtures"
)

func Test_JSON_Activity(t *testing.T) {
	var activity Activity
	err := json.Unmarshal(fixtures_tests.ActivityFixture, &activity)
	assert.NoError(t, err)

	assert.Equal(t, "walking", activity.ActivityType.String())
	assert.Equal(t, int64(3600), activity.Duration.Int64())
	assert.Equal(t, int64(120), activity.HeartRate.Int64())
	assert.Equal(t, int64(1000), activity.Steps.Int64())
	assert.Equal(t, "activity walking", activity.Notes.String())
	assert.Equal(t, int64(720), activity.UtcOffsetMinutes.Int64())

	activityJson, err := json.MarshalIndent(activity, "", "  ")
	fmt.Printf("Activity JSON: %s %v\n", activityJson, err)
}

func Test_ActivityToCreateParams(t *testing.T) {
	var activity Activity
	err := json.Unmarshal(fixtures_tests.ActivityFixture, &activity)
	assert.NoError(t, err)

	createParams := ActivityToCreate(activity, CreateContext{"prefix", 123, "identityName"})

	assert.NotNil(t, createParams)
	assert.Equal(t, "walking", *createParams.Type)
	assert.Equal(t, int64(3600), *createParams.DurationMinutes)
	assert.Equal(t, int64(120), *createParams.HeartRate)
	assert.Equal(t, int64(1000), *createParams.Steps)
	assert.Equal(t, "activity walking", *createParams.Notes)
	assert.Equal(t, int64(720), *createParams.UtcOffsetMinutes)

	paramsJson, err := json.MarshalIndent(createParams, "", "  ")
	fmt.Printf("CreateActivityParams JSON: %s %v\n", paramsJson, err)
	fmt.Println(" ")
	activityJson, err := json.MarshalIndent(activity, "", "  ")
	fmt.Printf("Activity JSON: %s %v\n", activityJson, err)
}
