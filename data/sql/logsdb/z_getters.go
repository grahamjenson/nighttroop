
package logsdb


// GetID returns the ID value.
func (m Error) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Error.
func (m *Error) SetID(v int64) {
	m.ID = v
}

// GetCreatedAt returns the CreatedAt value.
func (m Error) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated Error.
func (m *Error) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetID returns the ID value.
func (m Event) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Event.
func (m *Event) SetID(v int64) {
	m.ID = v
}

// GetName returns the Name value.
func (m Event) GetName() string {
	return m.Name
}

// GetID returns the ID value.
func (m Log) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Log.
func (m *Log) SetID(v int64) {
	m.ID = v
}

// SetLevel sets the Level value and returns the updated Log.
func (m *Log) SetLevel(v string) {
	m.Level = v
}

// GetCreatedAt returns the CreatedAt value.
func (m Log) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated Log.
func (m *Log) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetID returns the ID value.
func (m GetErrorParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetErrorParams.
func (m *GetErrorParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetEventParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetEventParams.
func (m *GetEventParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetLogParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetLogParams.
func (m *GetLogParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetSiteErrorParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetSiteErrorParams.
func (m *GetSiteErrorParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetSiteEventParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetSiteEventParams.
func (m *GetSiteEventParams) SetID(v int64) {
	m.ID = v
}

// GetID returns the ID value.
func (m GetSiteLogParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetSiteLogParams.
func (m *GetSiteLogParams) SetID(v int64) {
	m.ID = v
}

// GetName returns the Name value.
func (m InsertEventParams) GetName() string {
	return m.Name
}

// SetLevel sets the Level value and returns the updated InsertLogParams.
func (m *InsertLogParams) SetLevel(v string) {
	m.Level = v
}

// GetCreatedAt returns the CreatedAt value.
func (m InsertLogParams) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated InsertLogParams.
func (m *InsertLogParams) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// SetLimit sets the Limit value and returns the updated RecentConnectorErrorsParams.
func (m *RecentConnectorErrorsParams) SetLimit(v int64) {
	m.Limit = v
}

// SetLimit sets the Limit value and returns the updated RecentConnectorEventsParams.
func (m *RecentConnectorEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetType returns the Type value.
func (m SearchErrorsParams) GetType() *string {
	return m.Type
}

// GetSearch returns the Search value.
func (m SearchErrorsParams) GetSearch() *string {
	return m.Search
}

// SetLimit sets the Limit value and returns the updated SearchErrorsParams.
func (m *SearchErrorsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetType returns the Type value.
func (m SearchEventsParams) GetType() *string {
	return m.Type
}

// GetSearch returns the Search value.
func (m SearchEventsParams) GetSearch() *string {
	return m.Search
}

// SetLimit sets the Limit value and returns the updated SearchEventsParams.
func (m *SearchEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSearch returns the Search value.
func (m SearchLogsParams) GetSearch() *string {
	return m.Search
}

// SetLimit sets the Limit value and returns the updated SearchLogsParams.
func (m *SearchLogsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSearch returns the Search value.
func (m SearchSiteErrorsParams) GetSearch() *string {
	return m.Search
}

// GetType returns the Type value.
func (m SearchSiteErrorsParams) GetType() *string {
	return m.Type
}

// SetLimit sets the Limit value and returns the updated SearchSiteErrorsParams.
func (m *SearchSiteErrorsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSearch returns the Search value.
func (m SearchSiteEventsParams) GetSearch() *string {
	return m.Search
}

// GetType returns the Type value.
func (m SearchSiteEventsParams) GetType() *string {
	return m.Type
}

// SetLimit sets the Limit value and returns the updated SearchSiteEventsParams.
func (m *SearchSiteEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSearch returns the Search value.
func (m SearchSiteLogsParams) GetSearch() *string {
	return m.Search
}

// SetLimit sets the Limit value and returns the updated SearchSiteLogsParams.
func (m *SearchSiteLogsParams) SetLimit(v int64) {
	m.Limit = v
}

