package sitemodels

// SiteDeviceStatus is one status report from an uploader: its battery, the pump and the loop
type SiteDeviceStatus struct {
	ID                    string   `json:"id"`
	DeviceBatteryPercent  *float64 `json:"device_battery_percent"`
	IsDeviceCharging      *bool    `json:"is_device_charging"`
	PumpState             *string  `json:"pump_state"`
	PumpReservoir         *float64 `json:"pump_reservoir"`
	PumpBatteryPercent    *float64 `json:"pump_battery_percent"`
	PumpChargingState     *string  `json:"pump_charging_state"`
	LoopName              *string  `json:"loop_name"`
	LoopIob               *float64 `json:"loop_iob"`
	LoopIobBasal          *float64 `json:"loop_iob_basal"`
	LoopIobTreatment      *float64 `json:"loop_iob_treatment"`
	LoopCob               *float64 `json:"loop_cob"`
	RawJsonData           *string  `json:"raw_json_data"`
	UtcOffsetMinutes      *int64   `json:"utc_offset_minutes"`
	OccurredAt            int64    `json:"occurred_at"`
	CreatedAt             int64    `json:"created_at"`
	UpdatedAt             int64    `json:"updated_at"`
	CreatedByAppName      *string  `json:"created_by_app_name"`
	CreatedByDeviceName   *string  `json:"created_by_device_name"`
	CreatedByIdentityName string   `json:"created_by_identity_name"`
	UpdatedByIdentityName *string  `json:"updated_by_identity_name"`
}
