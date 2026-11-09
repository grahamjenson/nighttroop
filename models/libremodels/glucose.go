package libremodels

import (
	"encoding/json"
	"fmt"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
)

// timestampLayout is how LibreLinkUp writes times: FactoryTimestamp in UTC, Timestamp in the sensor's local time
const timestampLayout = "1/2/2006 3:04:05 PM"

// trendDirections are the Nightscout directions of LibreLinkUp's TrendArrow 1 to 5
var trendDirections = map[int]string{1: "SingleDown", 2: "FortyFiveDown", 3: "Flat", 4: "FortyFiveUp", 5: "SingleUp"}

// Reading is a LibreLinkUp glucose measurement
type Reading struct {
	Mgdl        float64
	FactoryTime time.Time
	// UTCOffsetMinutes is how far the sensor's local time is ahead of UTC
	UTCOffsetMinutes int64
	// TrendArrow is only given on the latest measurement
	TrendArrow *int
}

type readingResponse struct {
	FactoryTimestamp string  `json:"FactoryTimestamp"`
	Timestamp        string  `json:"Timestamp"`
	ValueInMgPerDl   float64 `json:"ValueInMgPerDl"`
	TrendArrow       *int    `json:"TrendArrow"`
}

func (r *Reading) UnmarshalJSON(data []byte) error {
	var resp readingResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	if resp.ValueInMgPerDl <= 0 || resp.ValueInMgPerDl > consts.MaxGlucoseMgdl {
		return fmt.Errorf("libre reading %v mg/dL is out of range", resp.ValueInMgPerDl)
	}
	factory, err := time.Parse(timestampLayout, resp.FactoryTimestamp)
	if err != nil {
		return fmt.Errorf("invalid libre FactoryTimestamp: %w", err)
	}
	local, err := time.Parse(timestampLayout, resp.Timestamp)
	if err != nil {
		return fmt.Errorf("invalid libre Timestamp: %w", err)
	}

	r.Mgdl = resp.ValueInMgPerDl
	r.FactoryTime = factory
	r.UTCOffsetMinutes = int64(local.Sub(factory).Round(time.Minute) / time.Minute)
	r.TrendArrow = resp.TrendArrow
	return nil
}

func ReadingToCreate(r Reading, createdByIdentityName string) sitedb.CreateGlucoseReadingParams {
	params := sitedb.CreateGlucoseReadingParams{
		// Derived from the reading time so readings fetched again by later syncs are duplicates
		ID:                    fmt.Sprintf("libre-%d", r.FactoryTime.Unix()),
		OccurredAt:            r.FactoryTime.UnixMilli(),
		Mgdl:                  r.Mgdl,
		ReadingType:           "sgv",
		CreatedByIdentityName: createdByIdentityName,
		UtcOffsetMinutes:      &r.UTCOffsetMinutes,
	}
	if r.TrendArrow != nil {
		if direction, ok := trendDirections[*r.TrendArrow]; ok {
			params.Direction = &direction
		}
	}
	return params
}
