
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

// GetID returns the ID value.
func (m Connection) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Connection.
func (m *Connection) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m Connection) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated Connection.
func (m *Connection) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m Connection) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated Connection.
func (m *Connection) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetKind returns the Kind value.
func (m Connection) GetKind() string {
	return m.Kind
}

// SetKind sets the Kind value and returns the updated Connection.
func (m *Connection) SetKind(v string) {
	m.Kind = v
}

// GetCollection returns the Collection value.
func (m Connection) GetCollection() string {
	return m.Collection
}

// SetCollection sets the Collection value and returns the updated Connection.
func (m *Connection) SetCollection(v string) {
	m.Collection = v
}

// GetOpenedAt returns the OpenedAt value.
func (m Connection) GetOpenedAt() int64 {
	return m.OpenedAt
}

// SetOpenedAt sets the OpenedAt value and returns the updated Connection.
func (m *Connection) SetOpenedAt(v int64) {
	m.OpenedAt = v
}

// GetAt returns the At value.
func (m ResourceReading) GetAt() int64 {
	return m.At
}

// SetAt sets the At value and returns the updated ResourceReading.
func (m *ResourceReading) SetAt(v int64) {
	m.At = v
}

// GetCpuPercent returns the CpuPercent value.
func (m ResourceReading) GetCpuPercent() float64 {
	return m.CpuPercent
}

// SetCpuPercent sets the CpuPercent value and returns the updated ResourceReading.
func (m *ResourceReading) SetCpuPercent(v float64) {
	m.CpuPercent = v
}

// GetCpuPeggedSeconds returns the CpuPeggedSeconds value.
func (m ResourceReading) GetCpuPeggedSeconds() int64 {
	return m.CpuPeggedSeconds
}

// SetCpuPeggedSeconds sets the CpuPeggedSeconds value and returns the updated ResourceReading.
func (m *ResourceReading) SetCpuPeggedSeconds(v int64) {
	m.CpuPeggedSeconds = v
}

// GetMemoryPercent returns the MemoryPercent value.
func (m ResourceReading) GetMemoryPercent() *float64 {
	return m.MemoryPercent
}

// SetMemoryPercent sets the MemoryPercent value and returns the updated ResourceReading.
func (m *ResourceReading) SetMemoryPercent(v *float64) {
	m.MemoryPercent = v
}

// GetDiskPercent returns the DiskPercent value.
func (m ResourceReading) GetDiskPercent() *float64 {
	return m.DiskPercent
}

// SetDiskPercent sets the DiskPercent value and returns the updated ResourceReading.
func (m *ResourceReading) SetDiskPercent(v *float64) {
	m.DiskPercent = v
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

// GetIncludeSnoozed returns the IncludeSnoozed value.
func (m IdentityAlarmsParams) GetIncludeSnoozed() bool {
	return m.IncludeSnoozed
}

// SetIncludeSnoozed sets the IncludeSnoozed value and returns the updated IdentityAlarmsParams.
func (m *IdentityAlarmsParams) SetIncludeSnoozed(v bool) {
	m.IncludeSnoozed = v
}

// GetNow returns the Now value.
func (m IdentityAlarmsParams) GetNow() int64 {
	return m.Now
}

// SetNow sets the Now value and returns the updated IdentityAlarmsParams.
func (m *IdentityAlarmsParams) SetNow(v int64) {
	m.Now = v
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

// GetID returns the ID value.
func (m CloseConnectionParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated CloseConnectionParams.
func (m *CloseConnectionParams) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m OpenConnectionParams) GetSiteID() string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated OpenConnectionParams.
func (m *OpenConnectionParams) SetSiteID(v string) {
	m.SiteID = v
}

// GetIdentityID returns the IdentityID value.
func (m OpenConnectionParams) GetIdentityID() int64 {
	return m.IdentityID
}

// SetIdentityID sets the IdentityID value and returns the updated OpenConnectionParams.
func (m *OpenConnectionParams) SetIdentityID(v int64) {
	m.IdentityID = v
}

// GetKind returns the Kind value.
func (m OpenConnectionParams) GetKind() string {
	return m.Kind
}

// SetKind sets the Kind value and returns the updated OpenConnectionParams.
func (m *OpenConnectionParams) SetKind(v string) {
	m.Kind = v
}

// GetCollection returns the Collection value.
func (m OpenConnectionParams) GetCollection() string {
	return m.Collection
}

// SetCollection sets the Collection value and returns the updated OpenConnectionParams.
func (m *OpenConnectionParams) SetCollection(v string) {
	m.Collection = v
}

// GetMax returns the Max value.
func (m OpenConnectionParams) GetMax() int64 {
	return m.Max
}

// SetMax sets the Max value and returns the updated OpenConnectionParams.
func (m *OpenConnectionParams) SetMax(v int64) {
	m.Max = v
}

// GetAt returns the At value.
func (m CreateResourceReadingParams) GetAt() int64 {
	return m.At
}

// SetAt sets the At value and returns the updated CreateResourceReadingParams.
func (m *CreateResourceReadingParams) SetAt(v int64) {
	m.At = v
}

// GetCPUPercent returns the CPUPercent value.
func (m CreateResourceReadingParams) GetCPUPercent() float64 {
	return m.CPUPercent
}

// SetCPUPercent sets the CPUPercent value and returns the updated CreateResourceReadingParams.
func (m *CreateResourceReadingParams) SetCPUPercent(v float64) {
	m.CPUPercent = v
}

// GetCPUPeggedSeconds returns the CPUPeggedSeconds value.
func (m CreateResourceReadingParams) GetCPUPeggedSeconds() int64 {
	return m.CPUPeggedSeconds
}

// SetCPUPeggedSeconds sets the CPUPeggedSeconds value and returns the updated CreateResourceReadingParams.
func (m *CreateResourceReadingParams) SetCPUPeggedSeconds(v int64) {
	m.CPUPeggedSeconds = v
}

// GetMemoryPercent returns the MemoryPercent value.
func (m CreateResourceReadingParams) GetMemoryPercent() *float64 {
	return m.MemoryPercent
}

// SetMemoryPercent sets the MemoryPercent value and returns the updated CreateResourceReadingParams.
func (m *CreateResourceReadingParams) SetMemoryPercent(v *float64) {
	m.MemoryPercent = v
}

// GetDiskPercent returns the DiskPercent value.
func (m CreateResourceReadingParams) GetDiskPercent() *float64 {
	return m.DiskPercent
}

// SetDiskPercent sets the DiskPercent value and returns the updated CreateResourceReadingParams.
func (m *CreateResourceReadingParams) SetDiskPercent(v *float64) {
	m.DiskPercent = v
}

// GetBefore returns the Before value.
func (m DeleteResourceReadingsBeforeParams) GetBefore() int64 {
	return m.Before
}

// SetBefore sets the Before value and returns the updated DeleteResourceReadingsBeforeParams.
func (m *DeleteResourceReadingsBeforeParams) SetBefore(v int64) {
	m.Before = v
}

// GetSince returns the Since value.
func (m ResourceReadingsSinceParams) GetSince() int64 {
	return m.Since
}

// SetSince sets the Since value and returns the updated ResourceReadingsSinceParams.
func (m *ResourceReadingsSinceParams) SetSince(v int64) {
	m.Since = v
}

