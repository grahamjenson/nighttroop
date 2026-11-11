package settings

// slot is where a SettingMap keeps one setting, so the setting can be read or written by its identifier
type slot interface {
	identifier() string
	anyType() *SettingAnyType
	parseValidate(value any, createdAt *int64, subject *string) error
}

type settingSlot[T any] struct {
	definition *SettingDefinition[T]
	setting    **Setting[T]
}

func slotOf[T any](definition *SettingDefinition[T], setting **Setting[T]) slot {
	return settingSlot[T]{definition: definition, setting: setting}
}

func (s settingSlot[T]) identifier() string {
	return s.definition.Identifier
}

func (s settingSlot[T]) anyType() *SettingAnyType {
	return (*s.setting).ToSettingAnyType()
}

// parseValidate keeps the value parsed and validated, or no setting if the value is nil or not valid
func (s settingSlot[T]) parseValidate(value any, createdAt *int64, subject *string) error {
	var err error
	*s.setting, err = NewSettingAndParseValidate(s.definition, value, createdAt, subject)
	return err
}

// slots are where m keeps each setting, in the order the settings are listed
func (m *SettingMap) slots() []slot {
	return []slot{
		slotOf(Units, &m.Units),
		slotOf(UnitsMgdlLabel, &m.UnitsMgdlLabel),
		slotOf(UnitsMmolLabel, &m.UnitsMmolLabel),
		slotOf(Timezone, &m.Timezone),
		slotOf(Language, &m.Language),
		slotOf(DateFormat, &m.DateFormat),
		slotOf(TimeFormat, &m.TimeFormat),
		slotOf(WeekStartDay, &m.WeekStartDay),
		slotOf(BolusDIAHours, &m.BolusDiaHours),
		slotOf(BolusOnsetMinutes, &m.BolusOnsetMinutes),
		slotOf(BolusPeakMinutes, &m.BolusPeakMinutes),
		slotOf(BasalDIAHours, &m.BasalDiaHours),
		slotOf(BasalOnsetMinutes, &m.BasalOnsetMinutes),
		slotOf(BasalPeakMinutes, &m.BasalPeakMinutes),
		slotOf(DefaultInsulinSensitivity, &m.DefaultInsulinSensitivity),
		slotOf(DefaultCarbRatio, &m.DefaultCarbRatio),
		slotOf(DefaultBasalRate, &m.DefaultBasalRate),
		slotOf(CarbMaxAbsorptionHours, &m.CarbMaxAbsorptionHours),
		slotOf(CarbMin5mImpact, &m.CarbMin5mImpact),
		slotOf(AlarmUrgentLowEnabled, &m.AlarmUrgentLowEnabled),
		slotOf(RangeUrgentLowMgdl, &m.RangeUrgentLowMgdl),
		slotOf(AlarmLowEnabled, &m.AlarmLowEnabled),
		slotOf(RangeLowMgdl, &m.RangeLowMgdl),
		slotOf(AlarmLowMinutes, &m.AlarmLowMinutes),
		slotOf(AlarmLowPredictionEnabled, &m.AlarmLowPredictionEnabled),
		slotOf(AlarmLowPredictionMinutes, &m.AlarmLowPredictionMinutes),
		slotOf(AlarmHighEnabled, &m.AlarmHighEnabled),
		slotOf(RangeHighMgdl, &m.RangeHighMgdl),
		slotOf(AlarmHighMinutes, &m.AlarmHighMinutes),
		slotOf(AlarmUrgentHighEnabled, &m.AlarmUrgentHighEnabled),
		slotOf(RangeUrgentHighMgdl, &m.RangeUrgentHighMgdl),
		slotOf(AlarmFastRiseEnabled, &m.AlarmFastRiseEnabled),
		slotOf(AlarmFastRiseMgdl, &m.AlarmFastRiseMgdl),
		slotOf(AlarmFastRiseMinutes, &m.AlarmFastRiseMinutes),
		slotOf(AlarmFastDropEnabled, &m.AlarmFastDropEnabled),
		slotOf(AlarmFastDropMgdl, &m.AlarmFastDropMgdl),
		slotOf(AlarmFastDropMinutes, &m.AlarmFastDropMinutes),
		slotOf(AlarmNoDataEnabled, &m.AlarmNoDataEnabled),
		slotOf(AlarmNoDataMinutes, &m.AlarmNoDataMinutes),
		slotOf(AlarmDeviceBatteryLowEnabled, &m.AlarmDeviceBatteryLowEnabled),
		slotOf(AlarmDeviceBatteryLowPercent, &m.AlarmDeviceBatteryLowPercent),
		slotOf(AlarmDeviceBatteryLowMinutes, &m.AlarmDeviceBatteryLowMinutes),
		slotOf(AlarmDeviceBatteryLowDevicesJson, &m.AlarmDeviceBatteryLowDevicesJson),
		slotOf(AlarmLowReservoirEnabled, &m.AlarmLowReservoirEnabled),
		slotOf(AlarmLowReservoirUnits, &m.AlarmLowReservoirUnits),
		slotOf(NotificationsEnabled, &m.NotificationsEnabled),
	}
}

// slotFor is where m keeps the setting with the identifier, or nil if there is no such setting
func (m *SettingMap) slotFor(identifier string) slot {
	for _, s := range m.slots() {
		if s.identifier() == identifier {
			return s
		}
	}
	return nil
}

// allSettingNames are the identifiers of every setting, in the order they are listed
var allSettingNames = func() []string {
	var names []string
	for _, s := range (&SettingMap{}).slots() {
		names = append(names, s.identifier())
	}
	return names
}()
