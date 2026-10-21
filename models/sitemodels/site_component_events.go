package sitemodels

import "nighttroop.t1d.nz/models/nightscoutmodels"

// ComponentEventTypes are the Nightscout device events that change a sensor, site or pump part
var ComponentEventTypes = []string{
	nightscoutmodels.SensorStart.String(),
	nightscoutmodels.SensorChange.String(),
	nightscoutmodels.SiteChange.String(),
	nightscoutmodels.CannulaFill.String(),
	nightscoutmodels.Prime.String(),
	nightscoutmodels.InsulinChange.String(),
	nightscoutmodels.ReservoirChange.String(),
	nightscoutmodels.PumpBatteryChange.String(),
}

type SiteComponentEvent struct {
	ID string `json:"id"`
	// EventType is the change, e.g. Sensor Change or Site Change
	EventType             string  `json:"event_type"`
	SensorCode            *string `json:"sensor_code"`
	TransmitterID         *string `json:"transmitter_id"`
	PumpID                *string `json:"pump_id"`
	PumpSerial            *string `json:"pump_serial"`
	PumpType              *string `json:"pump_type"`
	Notes                 *string `json:"notes"`
	Reason                *string `json:"reason"`
	RawJsonData           *string `json:"raw_json_data"`
	UtcOffsetMinutes      *int64  `json:"utc_offset_minutes"`
	OccurredAt            int64   `json:"occurred_at"`
	CreatedAt             int64   `json:"created_at"`
	UpdatedAt             int64   `json:"updated_at"`
	CreatedByAppName      *string `json:"created_by_app_name"`
	CreatedByDeviceName   *string `json:"created_by_device_name"`
	CreatedByIdentityName string  `json:"created_by_identity_name"`
	UpdatedByIdentityName *string `json:"updated_by_identity_name"`
}
