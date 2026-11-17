
package livedb

import (
	"nighttroop.t1d.nz/utils/alarms"
)


// GetIdentityID returns the IdentityID value.
func (m Alarm) GetIdentityID() int64 {
	return m.IdentityID
}

// GetType returns the Type value.
func (m Alarm) GetType() alarms.Type {
	return m.Type
}

// GetID returns the ID value.
func (m Connection) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Connection.
func (m *Connection) SetID(v int64) {
	m.ID = v
}

// GetIdentityID returns the IdentityID value.
func (m Connection) GetIdentityID() int64 {
	return m.IdentityID
}

// GetIdentityID returns the IdentityID value.
func (m DeleteAlarmParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetType returns the Type value.
func (m DeleteAlarmParams) GetType() alarms.Type {
	return m.Type
}

// GetIdentityID returns the IdentityID value.
func (m GetIdentityAlarmParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetType returns the Type value.
func (m GetIdentityAlarmParams) GetType() alarms.Type {
	return m.Type
}

// GetIdentityID returns the IdentityID value.
func (m IdentityAlarmsParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetIdentityID returns the IdentityID value.
func (m SnoozeAlarmParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetType returns the Type value.
func (m SnoozeAlarmParams) GetType() alarms.Type {
	return m.Type
}

// GetIdentityID returns the IdentityID value.
func (m UpsertAlarmParams) GetIdentityID() int64 {
	return m.IdentityID
}

// GetType returns the Type value.
func (m UpsertAlarmParams) GetType() alarms.Type {
	return m.Type
}

// GetID returns the ID value.
func (m CloseConnectionParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated CloseConnectionParams.
func (m *CloseConnectionParams) SetID(v int64) {
	m.ID = v
}

// GetIdentityID returns the IdentityID value.
func (m OpenConnectionParams) GetIdentityID() int64 {
	return m.IdentityID
}

