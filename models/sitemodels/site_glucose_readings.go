package sitemodels

import "nighttroop.t1d.nz/utils/consts"

type SiteGlucoseReading struct {
	ID                    string        `json:"id"`
	Mgdl                  float64       `json:"mgdl"`
	ReadingType           string        `json:"reading_type"`
	Units                 *consts.Units `json:"units"`
	MeasurementMethod     *string       `json:"measurement_method"`
	Delta                 *float64      `json:"delta"`
	Direction             *string       `json:"direction"`
	Noise                 *float64      `json:"noise"`
	Filtered              *float64      `json:"filtered"`
	Unfiltered            *float64      `json:"unfiltered"`
	Rssi                  *float64      `json:"rssi"`
	Notes                 *string       `json:"notes"`
	Reason                *string       `json:"reason"`
	RawJsonData           *string       `json:"raw_json_data"`
	UtcOffsetMinutes      *int64        `json:"utc_offset_minutes"`
	OccurredAt            int64         `json:"occurred_at"`
	CreatedAt             int64         `json:"created_at"`
	UpdatedAt             int64         `json:"updated_at"`
	CreatedByAppName      *string       `json:"created_by_app_name"`
	CreatedByDeviceName   *string       `json:"created_by_device_name"`
	CreatedByIdentityName string        `json:"created_by_identity_name"`
	UpdatedByIdentityName *string       `json:"updated_by_identity_name"`
}
