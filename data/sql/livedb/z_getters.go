
package livedb

import (
	"nighttroop.t1d.nz/utils/alarms"
)


// GetSiteID returns the SiteID value.
func (m Alarm) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated Alarm.
func (m *Alarm) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m Alarm) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated Alarm.
func (m *Alarm) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetType returns the Type value.
func (m Alarm) GetType() alarms.Type {
	return m.Type
}

// SetType sets the Type value and returns the updated Alarm.
func (m *Alarm) SetType(v alarms.Type) {
	m.Type = v
}

// GetActiveSince returns the ActiveSince value.
func (m Alarm) GetActiveSince() int64 {
	return m.ActiveSince
}

// SetActiveSince sets the ActiveSince value and returns the updated Alarm.
func (m *Alarm) SetActiveSince(v int64) {
	m.ActiveSince = v
}

// GetNotifiedAt returns the NotifiedAt value.
func (m Alarm) GetNotifiedAt() int64 {
	return m.NotifiedAt
}

// SetNotifiedAt sets the NotifiedAt value and returns the updated Alarm.
func (m *Alarm) SetNotifiedAt(v int64) {
	m.NotifiedAt = v
}

// GetSnoozedUntil returns the SnoozedUntil value.
func (m Alarm) GetSnoozedUntil() int64 {
	return m.SnoozedUntil
}

// SetSnoozedUntil sets the SnoozedUntil value and returns the updated Alarm.
func (m *Alarm) SetSnoozedUntil(v int64) {
	m.SnoozedUntil = v
}

// GetAlerting returns the Alerting value.
func (m Alarm) GetAlerting() bool {
	return m.Alerting
}

// SetAlerting sets the Alerting value and returns the updated Alarm.
func (m *Alarm) SetAlerting(v bool) {
	m.Alerting = v
}

// GetSiteID returns the SiteID value.
func (m NightscoutCursor) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated NightscoutCursor.
func (m *NightscoutCursor) SetSiteID(v string) {
	m.SiteID = v
}

// GetCollection returns the Collection value.
func (m NightscoutCursor) GetCollection() string {
	return m.Collection
}

// SetCollection sets the Collection value and returns the updated NightscoutCursor.
func (m *NightscoutCursor) SetCollection(v string) {
	m.Collection = v
}

// GetSource returns the Source value.
func (m NightscoutCursor) GetSource() string {
	return m.Source
}

// SetSource sets the Source value and returns the updated NightscoutCursor.
func (m *NightscoutCursor) SetSource(v string) {
	m.Source = v
}

// GetLastModified returns the LastModified value.
func (m NightscoutCursor) GetLastModified() int64 {
	return m.LastModified
}

// SetLastModified sets the LastModified value and returns the updated NightscoutCursor.
func (m *NightscoutCursor) SetLastModified(v int64) {
	m.LastModified = v
}

// GetSiteID returns the SiteID value.
func (m DeleteAlarmParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated DeleteAlarmParams.
func (m *DeleteAlarmParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m DeleteAlarmParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated DeleteAlarmParams.
func (m *DeleteAlarmParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetType returns the Type value.
func (m DeleteAlarmParams) GetType() alarms.Type {
	return m.Type
}

// SetType sets the Type value and returns the updated DeleteAlarmParams.
func (m *DeleteAlarmParams) SetType(v alarms.Type) {
	m.Type = v
}

// GetSnoozedUntil returns the SnoozedUntil value.
func (m DeleteAlarmParams) GetSnoozedUntil() int64 {
	return m.SnoozedUntil
}

// SetSnoozedUntil sets the SnoozedUntil value and returns the updated DeleteAlarmParams.
func (m *DeleteAlarmParams) SetSnoozedUntil(v int64) {
	m.SnoozedUntil = v
}

// GetSiteID returns the SiteID value.
func (m GetIdentityAlarmParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetIdentityAlarmParams.
func (m *GetIdentityAlarmParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m GetIdentityAlarmParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated GetIdentityAlarmParams.
func (m *GetIdentityAlarmParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetType returns the Type value.
func (m GetIdentityAlarmParams) GetType() alarms.Type {
	return m.Type
}

// SetType sets the Type value and returns the updated GetIdentityAlarmParams.
func (m *GetIdentityAlarmParams) SetType(v alarms.Type) {
	m.Type = v
}

// GetSiteID returns the SiteID value.
func (m IdentityAlarmsParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated IdentityAlarmsParams.
func (m *IdentityAlarmsParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m IdentityAlarmsParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated IdentityAlarmsParams.
func (m *IdentityAlarmsParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetSiteID returns the SiteID value.
func (m SiteAlarmsParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SiteAlarmsParams.
func (m *SiteAlarmsParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetSiteID returns the SiteID value.
func (m SnoozeAlarmParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SnoozeAlarmParams.
func (m *SnoozeAlarmParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m SnoozeAlarmParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated SnoozeAlarmParams.
func (m *SnoozeAlarmParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetType returns the Type value.
func (m SnoozeAlarmParams) GetType() alarms.Type {
	return m.Type
}

// SetType sets the Type value and returns the updated SnoozeAlarmParams.
func (m *SnoozeAlarmParams) SetType(v alarms.Type) {
	m.Type = v
}

// GetUntil returns the Until value.
func (m SnoozeAlarmParams) GetUntil() int64 {
	return m.Until
}

// SetUntil sets the Until value and returns the updated SnoozeAlarmParams.
func (m *SnoozeAlarmParams) SetUntil(v int64) {
	m.Until = v
}

// GetNow returns the Now value.
func (m SnoozedAlertingSiteIDsParams) GetNow() int64 {
	return m.Now
}

// SetNow sets the Now value and returns the updated SnoozedAlertingSiteIDsParams.
func (m *SnoozedAlertingSiteIDsParams) SetNow(v int64) {
	m.Now = v
}

// GetSiteID returns the SiteID value.
func (m UpsertAlarmParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated UpsertAlarmParams.
func (m *UpsertAlarmParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m UpsertAlarmParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated UpsertAlarmParams.
func (m *UpsertAlarmParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetType returns the Type value.
func (m UpsertAlarmParams) GetType() alarms.Type {
	return m.Type
}

// SetType sets the Type value and returns the updated UpsertAlarmParams.
func (m *UpsertAlarmParams) SetType(v alarms.Type) {
	m.Type = v
}

// GetActiveSince returns the ActiveSince value.
func (m UpsertAlarmParams) GetActiveSince() int64 {
	return m.ActiveSince
}

// SetActiveSince sets the ActiveSince value and returns the updated UpsertAlarmParams.
func (m *UpsertAlarmParams) SetActiveSince(v int64) {
	m.ActiveSince = v
}

// GetNotifiedAt returns the NotifiedAt value.
func (m UpsertAlarmParams) GetNotifiedAt() int64 {
	return m.NotifiedAt
}

// SetNotifiedAt sets the NotifiedAt value and returns the updated UpsertAlarmParams.
func (m *UpsertAlarmParams) SetNotifiedAt(v int64) {
	m.NotifiedAt = v
}

// GetSnoozedUntil returns the SnoozedUntil value.
func (m UpsertAlarmParams) GetSnoozedUntil() int64 {
	return m.SnoozedUntil
}

// SetSnoozedUntil sets the SnoozedUntil value and returns the updated UpsertAlarmParams.
func (m *UpsertAlarmParams) SetSnoozedUntil(v int64) {
	m.SnoozedUntil = v
}

// GetAlerting returns the Alerting value.
func (m UpsertAlarmParams) GetAlerting() bool {
	return m.Alerting
}

// SetAlerting sets the Alerting value and returns the updated UpsertAlarmParams.
func (m *UpsertAlarmParams) SetAlerting(v bool) {
	m.Alerting = v
}

// GetSiteID returns the SiteID value.
func (m NightscoutCursorsParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated NightscoutCursorsParams.
func (m *NightscoutCursorsParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetSource returns the Source value.
func (m NightscoutCursorsParams) GetSource() string {
	return m.Source
}

// SetSource sets the Source value and returns the updated NightscoutCursorsParams.
func (m *NightscoutCursorsParams) SetSource(v string) {
	m.Source = v
}

// GetSiteID returns the SiteID value.
func (m SetNightscoutCursorParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SetNightscoutCursorParams.
func (m *SetNightscoutCursorParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetCollection returns the Collection value.
func (m SetNightscoutCursorParams) GetCollection() string {
	return m.Collection
}

// SetCollection sets the Collection value and returns the updated SetNightscoutCursorParams.
func (m *SetNightscoutCursorParams) SetCollection(v string) {
	m.Collection = v
}

// GetSource returns the Source value.
func (m SetNightscoutCursorParams) GetSource() string {
	return m.Source
}

// SetSource sets the Source value and returns the updated SetNightscoutCursorParams.
func (m *SetNightscoutCursorParams) SetSource(v string) {
	m.Source = v
}

// GetLastModified returns the LastModified value.
func (m SetNightscoutCursorParams) GetLastModified() int64 {
	return m.LastModified
}

// SetLastModified sets the LastModified value and returns the updated SetNightscoutCursorParams.
func (m *SetNightscoutCursorParams) SetLastModified(v int64) {
	m.LastModified = v
}

