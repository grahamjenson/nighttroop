package collections

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"

	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/models/nightscout"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

type SettingsCollection struct {
	converter func(c *Setting, raw bool) nsutils.Record
}

func NewSettingsCollection() SettingsCollection {
	return SettingsCollection{
		converter: nsutils.DBModelToRecord(nsutils.Settings, ConvertFromNSToSetting),
	}
}

// ///
// Create Replace and Patch are pretty much the same things
// You have to pass an identifier along to match with the settings
// ///

func (c SettingsCollection) Create(ctx context.Context, siteContext *httputils.SiteContext, data map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	var nsModel nightscout.Setting
	err := utils.FromMap(data, &nsModel)
	if err != nil {
		slog.Error("failed to map data to model", "error", err, "collection", "settings")
		return nil, nsutils.BadDataBodyError
	}

	if nsModel.Identifier.String() == "" {
		slog.Error("setting identifier not available", "data", data, "collection", "settings")
		return nil, nsutils.ErrNotFound
	}

	identitySettingsMap := GetIdentitySettingsMap(siteContext.IdentityModel)
	setting, ok := identitySettingsMap[nsModel.Identifier.String()]
	if !ok {
		slog.Error("setting not found", "identifier", nsModel.Identifier.String())
		return nil, nsutils.ErrNotFound
	}
	setting.Value = data["value"]
	params, httpErr := IdentitySettingsMapToIdentitySettingsParams(siteContext.IdentityModel.ID, siteContext.IdentityModel.Subdomain, identitySettingsMap)
	if httpErr != nil {
		slog.Error("failed to convert settings map to params", "error", httpErr, "collection", "settings")
		return nil, httpErr
	}
	if params == nil {
		slog.Error("failed to convert settings map to params, got nil", "collection", "settings")
		return nil, nsutils.ErrInternal
	}
	err = siteContext.ControlDB.Queries.UpdateIdentitySettings(ctx, *params)
	if err != nil {
		slog.Error("failed to update identity settings", "error", err, "collection", "settings")
		return nil, nsutils.ErrInternal
	}

	return ptr(c.converter(setting, raw)), nil
}

func (c SettingsCollection) Replace(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, canCreate bool, raw bool) (*nsutils.Record, bool, *nsutils.HTTPError) {
	if identifier == "" {
		slog.Error("setting identifier not available", "data", patch, "collection", "settings")
		return nil, false, nsutils.ErrNotFound
	}
	if patch["identifier"] == nil {
		patch["identifier"] = identifier
	}

	var nsModel nightscout.Setting
	err := utils.FromMap(patch, &nsModel)
	if err != nil {
		slog.Error("failed to map data to model", "error", err, "collection", "settings")
		return nil, false, nsutils.BadDataBodyError
	}

	if nsModel.Identifier.String() == "" {
		nsModel.Identifier = ptr(nightscout.IdentifierString(identifier))
	}

	if nsModel.Identifier.String() != identifier {
		slog.Error("setting identifier mismatch", "data", patch, "collection", "settings")
		return nil, false, nsutils.ErrBadRequest
	}

	identitySettingsMap := GetIdentitySettingsMap(siteContext.IdentityModel)
	setting, ok := identitySettingsMap[nsModel.Identifier.String()]
	if !ok {
		slog.Error("setting not found", "identifier", nsModel.Identifier.String())
		return nil, false, nsutils.ErrNotFound
	}
	setting.Value = patch["value"]
	params, httpErr := IdentitySettingsMapToIdentitySettingsParams(siteContext.IdentityModel.ID, siteContext.IdentityModel.Subdomain, identitySettingsMap)
	if httpErr != nil {
		slog.Error("failed to convert settings map to params", "error", httpErr, "collection", "settings")
		return nil, false, httpErr
	}
	if params == nil {
		slog.Error("failed to convert settings map to params, got nil", "collection", "settings")
		return nil, false, nsutils.ErrInternal
	}
	err = siteContext.ControlDB.Queries.UpdateIdentitySettings(ctx, *params)
	if err != nil {
		slog.Error("failed to update identity settings", "error", err, "collection", "settings")
		return nil, false, nsutils.ErrInternal
	}

	return ptr(c.converter(setting, raw)), false, nil
}

func (c SettingsCollection) Patch(ctx context.Context, siteContext *httputils.SiteContext, identifier string, patch map[string]any, raw bool) (*nsutils.Record, *nsutils.HTTPError) {
	if identifier == "" {
		slog.Error("setting identifier not available", "data", patch, "collection", "settings")
		return nil, nsutils.ErrNotFound
	}
	if patch["identifier"] == nil {
		patch["identifier"] = identifier
	}

	record, _, err := c.Replace(ctx, siteContext, identifier, patch, false, raw)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (c SettingsCollection) Delete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	identitySettingsMap := GetIdentitySettingsMap(siteContext.IdentityModel)
	setting, ok := identitySettingsMap[identifier]
	if !ok {
		slog.Error("setting not found", "identifier", identifier)
		return nsutils.ErrNotFound
	}
	setting.Value = nil
	params, httpErr := IdentitySettingsMapToIdentitySettingsParams(siteContext.IdentityModel.ID, siteContext.IdentityModel.Subdomain, identitySettingsMap)
	if httpErr != nil {
		slog.Error("failed to convert settings map to params", "error", httpErr, "collection", "settings")
		return httpErr
	}
	if params == nil {
		slog.Error("failed to convert settings map to params, got nil", "collection", "settings")
		return nsutils.ErrInternal
	}
	err := siteContext.ControlDB.Queries.UpdateIdentitySettings(ctx, *params)
	if err != nil {
		slog.Error("failed to update identity settings", "error", err, "collection", "settings")
		return nsutils.ErrInternal
	}

	return nil
}

func (c SettingsCollection) Get(ctx context.Context, siteContext *httputils.SiteContext, identifier string, query nsquery.Query) (*nsutils.Record, *nsutils.HTTPError) {
	settingsMap := GetSettingsMap(siteContext.SiteModel, siteContext.IdentityModel)
	setting, ok := settingsMap[identifier]
	if !ok {
		return nil, nsutils.ErrNotFound
	}

	return ptr(c.converter(setting, query.Raw)), nil
}

func (c SettingsCollection) Search(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]nsutils.Record, *nsutils.HTTPError) {
	settingsMap := GetSettingsMap(siteContext.SiteModel, siteContext.IdentityModel)
	settingsList := make([]*Setting, 0, len(settingsMap))
	for _, setting := range settingsMap {
		settingsList = append(settingsList, setting)
	}
	return nsutils.CreateRecordSlice(settingsList, c.converter, query.Raw), nil
}

func (c SettingsCollection) DeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {
	return 0, nsutils.ErrNotImplemented
}

///////
// SETTINGS TYPE
///////

func ConvertFromNSToSetting(s *Setting) nightscout.Setting {
	return nightscout.Setting{
		Identifier: ptr(nightscout.IdentifierString(s.Identifier)),
		Value:      s.Value,
		Type:       ptr(nightscout.ShortString(s.Type)),
	}
}

type Setting struct {
	Identifier string `json:"identifier"`
	Value      any    `json:"value"`
	Type       string `json:"type"`

	isDefault bool              `json:"-"`
	site      *control.Site     `json:"-"`
	identity  *control.Identity `json:"-"`
}

func (s Setting) GetID() string {
	return s.Identifier
}

func (s Setting) GetDeletedAt() *int64 {
	return nil
}

func (s Setting) GetCreatedAt() int64 {
	if s.isDefault {
		if s.site != nil {
			return s.site.CreatedAt
		}
	}
	if s.identity != nil {
		return s.identity.CreatedAt
	}
	return 0
}

func (s Setting) GetUpdatedAt() int64 {
	if s.isDefault {
		if s.site != nil {
			return s.site.CreatedAt
		}
	}
	if s.identity != nil {
		return s.identity.CreatedAt
	}
	return 0
}

func (s Setting) GetCreatedByIdentityName() string {
	return ""
}

func (s Setting) GetUtcOffsetMinutes() *int64 {
	return nil
}

func (s Setting) GetOccurredAt() int64 {
	if s.isDefault {
		if s.site != nil {
			return s.site.CreatedAt
		}
	}
	if s.identity != nil {
		return s.identity.CreatedAt
	}
	return 0
}

func (s Setting) GetCreatedByAppName() *string {
	return ptr("")
}

func (s Setting) GetCreatedByDeviceName() *string {
	return ptr("")
}

func (s Setting) GetRawJsonData() *string {
	return nil
}

func GetIdentitySettingsMap(identity *control.Identity) map[string]*Setting {
	settings := make(map[string]*Setting)
	for _, setting := range control.ALL_SETTINGS {
		settings[setting] = NewSetting(
			setting,
			identity.GetSetting(setting),
			nil,
			nil,
			control.ALL_SETTINGS_TYPES[setting],
			nil,
			identity,
		)
	}

	return settings

}

func GetSettingsMap(site *control.Site, identity *control.Identity) map[string]*Setting {
	settings := make(map[string]*Setting)
	for _, setting := range control.ALL_SETTINGS {
		defaultValue := control.ALL_SETTINGS_DEFAULTS[setting]
		settings[setting] = NewSetting(
			setting,
			identity.GetSetting(setting),
			site.GetSetting(setting),
			defaultValue,
			control.ALL_SETTINGS_TYPES[setting],
			site,
			identity,
		)
	}

	return settings
}

func NewSetting(
	identifier string,
	identityValue any,
	siteValue any,
	defaultValue any,
	typ string,
	site *control.Site,
	identity *control.Identity,
) *Setting {
	return &Setting{
		Identifier: identifier,
		Value:      first(identityValue, siteValue, defaultValue),
		Type:       typ,
		isDefault:  identityValue == nil,
		site:       site,
		identity:   identity,
	}
}

func first(values ...any) any {
	for _, value := range values {
		if value == nil {
			continue
		}

		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
			if v.IsNil() {
				continue
			}
		}

		return value
	}

	return nil
}

func IdentitySettingsMapToIdentitySettingsParams(identityID int64, subdomain string, settingsMap map[string]*Setting) (*control.UpdateIdentitySettingsParams, *nsutils.HTTPError) {
	params := &control.UpdateIdentitySettingsParams{ID: identityID, Subdomain: subdomain}
	var err *nsutils.HTTPError

	params.Units, err = SettingToString(settingsMap, "units")
	if err != nil {
		return nil, err
	}

	params.Timezone, err = SettingToString(settingsMap, "timezone")
	if err != nil {
		return nil, err
	}
	params.Language, err = SettingToString(settingsMap, "language")
	if err != nil {
		return nil, err
	}
	params.DateFormat, err = SettingToString(settingsMap, "date_format")
	if err != nil {
		return nil, err
	}
	params.TimeFormat, err = SettingToString(settingsMap, "time_format")
	if err != nil {
		return nil, err
	}
	params.WeekStartDay, err = SettingToString(settingsMap, "week_start_day")
	if err != nil {
		return nil, err
	}
	params.BolusDiaHours, err = SettingToFloat64(settingsMap, "bolus_dia_hours")
	if err != nil {
		return nil, err
	}
	params.BolusOnsetMinutes, err = SettingToFloat64(settingsMap, "bolus_onset_minutes")
	if err != nil {
		return nil, err
	}
	params.BolusPeakMinutes, err = SettingToFloat64(settingsMap, "bolus_peak_minutes")
	if err != nil {
		return nil, err
	}
	params.BasalDiaHours, err = SettingToFloat64(settingsMap, "basal_dia_hours")
	if err != nil {
		return nil, err
	}
	params.BasalOnsetMinutes, err = SettingToFloat64(settingsMap, "basal_onset_minutes")
	if err != nil {
		return nil, err
	}
	params.BasalPeakMinutes, err = SettingToFloat64(settingsMap, "basal_peak_minutes")
	if err != nil {
		return nil, err
	}
	params.DefaultInsulinSensitivity, err = SettingToFloat64(settingsMap, "default_insulin_sensitivity")
	if err != nil {
		return nil, err
	}
	params.DefaultCarbRatio, err = SettingToFloat64(settingsMap, "default_carb_ratio")
	if err != nil {
		return nil, err
	}
	params.DefaultBasalRate, err = SettingToFloat64(settingsMap, "default_basal_rate")
	if err != nil {
		return nil, err
	}
	params.CarbMaxAbsorptionHours, err = SettingToFloat64(settingsMap, "carb_max_absorption_hours")
	if err != nil {
		return nil, err
	}
	params.CarbMin5mImpact, err = SettingToFloat64(settingsMap, "carb_min_5m_impact")
	if err != nil {
		return nil, err
	}
	params.AlarmUrgentLowEnabled, err = SettingToBool(settingsMap, "alarm_urgent_low_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmUrgentLow, err = SettingToInt64(settingsMap, "alarm_urgent_low")
	if err != nil {
		return nil, err
	}
	params.AlarmLowEnabled, err = SettingToBool(settingsMap, "alarm_low_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmLow, err = SettingToInt64(settingsMap, "alarm_low")
	if err != nil {
		return nil, err
	}
	params.AlarmLowMinutes, err = SettingToInt64(settingsMap, "alarm_low_minutes")
	if err != nil {
		return nil, err
	}
	params.AlarmLowPredictionEnabled, err = SettingToBool(settingsMap, "alarm_low_prediction_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmHighEnabled, err = SettingToBool(settingsMap, "alarm_high_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmHigh, err = SettingToInt64(settingsMap, "alarm_high")
	if err != nil {
		return nil, err
	}
	params.AlarmHighMinutes, err = SettingToInt64(settingsMap, "alarm_high_minutes")
	if err != nil {
		return nil, err
	}
	params.AlarmUrgentHighEnabled, err = SettingToBool(settingsMap, "alarm_urgent_high_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmUrgentHigh, err = SettingToInt64(settingsMap, "alarm_urgent_high")
	if err != nil {
		return nil, err
	}
	params.AlarmFastRiseEnabled, err = SettingToBool(settingsMap, "alarm_fast_rise_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmFastRise, err = SettingToInt64(settingsMap, "alarm_fast_rise")
	if err != nil {
		return nil, err
	}
	params.AlarmFastRiseMinutes, err = SettingToInt64(settingsMap, "alarm_fast_rise_minutes")
	if err != nil {
		return nil, err
	}
	params.AlarmFastDropEnabled, err = SettingToBool(settingsMap, "alarm_fast_drop_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmFastDrop, err = SettingToInt64(settingsMap, "alarm_fast_drop")
	if err != nil {
		return nil, err
	}
	params.AlarmFastDropMinutes, err = SettingToInt64(settingsMap, "alarm_fast_drop_minutes")
	if err != nil {
		return nil, err
	}
	params.AlarmNoDataEnabled, err = SettingToBool(settingsMap, "alarm_no_data_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmNoDataMinutes, err = SettingToInt64(settingsMap, "alarm_no_data_minutes")
	if err != nil {
		return nil, err
	}
	params.AlarmBatteryLowEnabled, err = SettingToBool(settingsMap, "alarm_battery_low_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmBatteryLowPercent, err = SettingToInt64(settingsMap, "alarm_battery_low_percent")
	if err != nil {
		return nil, err
	}
	params.AlarmBatteryLowPwdEnabled, err = SettingToBool(settingsMap, "alarm_battery_low_pwd_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmBatteryLowPwdPercent, err = SettingToInt64(settingsMap, "alarm_battery_low_pwd_percent")
	if err != nil {
		return nil, err
	}
	params.AlarmLowReservoirEnabled, err = SettingToBool(settingsMap, "alarm_low_reservoir_enabled")
	if err != nil {
		return nil, err
	}
	params.AlarmLowReservoirUnits, err = SettingToInt64(settingsMap, "alarm_low_reservoir_units")
	if err != nil {
		return nil, err
	}
	params.NotificationsEnabled, err = SettingToBool(settingsMap, "notifications_enabled")
	if err != nil {
		return nil, err
	}

	// Convert the settings map to the appropriate params here.
	// If an error occurs during conversion, return it.

	return params, nil
}

func SettingToString(settingsMap map[string]*Setting, identifier string) (*string, *nsutils.HTTPError) {
	s, ok := settingsMap[identifier]
	if !ok || s == nil {
		return nil, nsutils.BadStringFieldError(identifier)
	}

	if s.Value == nil {
		return nil, nil
	}
	switch s.Value.(type) {
	case string:
		str := s.Value.(string)
		return &str, nil
	case *string:
		strPtr := s.Value.(*string)
		return strPtr, nil
	default:
		return nil, nsutils.BadStringFieldError(fmt.Sprintf("%s not a string: %v type %T", s.Identifier, s.Value, s.Value))
	}
}

func SettingToFloat64(settingsMap map[string]*Setting, identifier string) (*float64, *nsutils.HTTPError) {
	s, ok := settingsMap[identifier]
	if !ok || s == nil {
		slog.Error("setting not found", "identifier", identifier)
		return nil, nsutils.BadNumericFieldError(identifier)
	}

	if s.Value == nil {
		return nil, nil
	}
	switch s.Value.(type) {
	case float64:
		f := s.Value.(float64)
		return &f, nil
	case *float64:
		fPtr := s.Value.(*float64)
		return fPtr, nil
	case int64:
		i := s.Value.(int64)
		f := float64(i)
		return &f, nil
	case *int64:
		iPtr := s.Value.(*int64)
		if iPtr == nil {
			return nil, nil
		}
		f := float64(*iPtr)
		return &f, nil
	case json.Number:
		num := s.Value.(json.Number)
		f, err := num.Float64()
		if err != nil {
			return nil, nsutils.BadNumericFieldError(fmt.Sprintf("%s not a number: %v type %T", s.Identifier, s.Value, s.Value))
		}
		return &f, nil
	case *json.Number:
		numPtr := s.Value.(*json.Number)
		if numPtr == nil {
			return nil, nil
		}
		f, err := (*numPtr).Float64()
		if err != nil {
			return nil, nsutils.BadNumericFieldError(fmt.Sprintf("%s not a number: %v type %T", s.Identifier, s.Value, s.Value))
		}
		return &f, nil
	default:
		return nil, nsutils.BadNumericFieldError(fmt.Sprintf("%s not a number: %v type %T", s.Identifier, s.Value, s.Value))
	}
}

func SettingToInt64(settingsMap map[string]*Setting, identifier string) (*int64, *nsutils.HTTPError) {
	s, ok := settingsMap[identifier]
	if !ok || s == nil {
		slog.Error("setting not found", "identifier", identifier)
		return nil, nsutils.BadNumericFieldError(identifier)
	}

	if s.Value == nil {
		return nil, nil
	}

	switch s.Value.(type) {
	case int64:
		i := s.Value.(int64)
		return &i, nil
	case *int64:
		iPtr := s.Value.(*int64)
		return iPtr, nil
	case float64:
		f := s.Value.(float64)
		i := int64(f)
		return &i, nil
	case *float64:
		fPtr := s.Value.(*float64)
		if fPtr == nil {
			return nil, nil
		}
		i := int64(*fPtr)
		return &i, nil
	case json.Number:
		num := s.Value.(json.Number)
		i, err := num.Int64()
		if err != nil {
			return nil, nsutils.BadNumericFieldError(fmt.Sprintf("%s not a number: %v type %T", s.Identifier, s.Value, s.Value))
		}
		return &i, nil
	case *json.Number:
		numPtr := s.Value.(*json.Number)
		if numPtr == nil {
			return nil, nil
		}
		i, err := (*numPtr).Int64()
		if err != nil {
			return nil, nsutils.BadNumericFieldError(fmt.Sprintf("%s not a number: %v type %T", s.Identifier, s.Value, s.Value))
		}
		return &i, nil
	default:
		return nil, nsutils.BadNumericFieldError(fmt.Sprintf("%s not an int64: %v type %T", s.Identifier, s.Value, s.Value))
	}
}

func SettingToBool(settingsMap map[string]*Setting, identifier string) (*bool, *nsutils.HTTPError) {
	s, ok := settingsMap[identifier]
	if !ok || s == nil {
		slog.Error("setting not found", "identifier", identifier)
		return nil, nsutils.BadBoolFieldError(identifier + " not found settings")
	}

	if s.Value == nil {
		return nil, nil
	}

	switch s.Value.(type) {
	case bool:
		b := s.Value.(bool)
		return &b, nil
	case *bool:
		bPtr := s.Value.(*bool)
		return bPtr, nil
	case string:
		str := s.Value.(string)
		if str == "true" {
			b := true
			return &b, nil
		} else if str == "false" {
			b := false
			return &b, nil
		} else {
			return nil, nsutils.BadBoolFieldError(fmt.Sprintf("%s not a bool: %v type %T", s.Identifier, s.Value, s.Value))
		}
	case *string:
		strPtr := s.Value.(*string)
		if strPtr == nil {
			return nil, nil
		}
		if *strPtr == "true" {
			b := true
			return &b, nil
		} else if *strPtr == "false" {
			b := false
			return &b, nil
		} else {
			return nil, nsutils.BadBoolFieldError(fmt.Sprintf("%s not a bool: %v type %T", s.Identifier, s.Value, s.Value))
		}
	default:
		return nil, nsutils.BadBoolFieldError(fmt.Sprintf("%s not a bool: %v type %T", s.Identifier, s.Value, s.Value))
	}
}
