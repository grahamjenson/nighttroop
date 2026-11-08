
package logsdb


// GetID returns the ID value.
func (m Error) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Error.
func (m *Error) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m Error) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated Error.
func (m *Error) SetSiteID(v *string) {
	m.SiteID = v
}

// GetMessage returns the Message value.
func (m Error) GetMessage() string {
	return m.Message
}

// SetMessage sets the Message value and returns the updated Error.
func (m *Error) SetMessage(v string) {
	m.Message = v
}

// GetAttributes returns the Attributes value.
func (m Error) GetAttributes() *string {
	return m.Attributes
}

// SetAttributes sets the Attributes value and returns the updated Error.
func (m *Error) SetAttributes(v *string) {
	m.Attributes = v
}

// GetRequestID returns the RequestID value.
func (m Error) GetRequestID() *string {
	return m.RequestID
}

// SetRequestID sets the RequestID value and returns the updated Error.
func (m *Error) SetRequestID(v *string) {
	m.RequestID = v
}

// GetJobID returns the JobID value.
func (m Error) GetJobID() *string {
	return m.JobID
}

// SetJobID sets the JobID value and returns the updated Error.
func (m *Error) SetJobID(v *string) {
	m.JobID = v
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

// GetSiteID returns the SiteID value.
func (m Event) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated Event.
func (m *Event) SetSiteID(v *string) {
	m.SiteID = v
}

// GetBatch returns the Batch value.
func (m Event) GetBatch() *string {
	return m.Batch
}

// SetBatch sets the Batch value and returns the updated Event.
func (m *Event) SetBatch(v *string) {
	m.Batch = v
}

// GetName returns the Name value.
func (m Event) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated Event.
func (m *Event) SetName(v string) {
	m.Name = v
}

// GetAttributes returns the Attributes value.
func (m Event) GetAttributes() *string {
	return m.Attributes
}

// SetAttributes sets the Attributes value and returns the updated Event.
func (m *Event) SetAttributes(v *string) {
	m.Attributes = v
}

// GetStartedAt returns the StartedAt value.
func (m Event) GetStartedAt() int64 {
	return m.StartedAt
}

// SetStartedAt sets the StartedAt value and returns the updated Event.
func (m *Event) SetStartedAt(v int64) {
	m.StartedAt = v
}

// GetFinishedAt returns the FinishedAt value.
func (m Event) GetFinishedAt() *int64 {
	return m.FinishedAt
}

// SetFinishedAt sets the FinishedAt value and returns the updated Event.
func (m *Event) SetFinishedAt(v *int64) {
	m.FinishedAt = v
}

// GetRequestID returns the RequestID value.
func (m Event) GetRequestID() *string {
	return m.RequestID
}

// SetRequestID sets the RequestID value and returns the updated Event.
func (m *Event) SetRequestID(v *string) {
	m.RequestID = v
}

// GetJobID returns the JobID value.
func (m Event) GetJobID() *string {
	return m.JobID
}

// SetJobID sets the JobID value and returns the updated Event.
func (m *Event) SetJobID(v *string) {
	m.JobID = v
}

// GetID returns the ID value.
func (m Log) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated Log.
func (m *Log) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m Log) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated Log.
func (m *Log) SetSiteID(v *string) {
	m.SiteID = v
}

// GetLevel returns the Level value.
func (m Log) GetLevel() string {
	return m.Level
}

// SetLevel sets the Level value and returns the updated Log.
func (m *Log) SetLevel(v string) {
	m.Level = v
}

// GetMessage returns the Message value.
func (m Log) GetMessage() string {
	return m.Message
}

// SetMessage sets the Message value and returns the updated Log.
func (m *Log) SetMessage(v string) {
	m.Message = v
}

// GetAttributes returns the Attributes value.
func (m Log) GetAttributes() *string {
	return m.Attributes
}

// SetAttributes sets the Attributes value and returns the updated Log.
func (m *Log) SetAttributes(v *string) {
	m.Attributes = v
}

// GetRequestID returns the RequestID value.
func (m Log) GetRequestID() *string {
	return m.RequestID
}

// SetRequestID sets the RequestID value and returns the updated Log.
func (m *Log) SetRequestID(v *string) {
	m.RequestID = v
}

// GetJobID returns the JobID value.
func (m Log) GetJobID() *string {
	return m.JobID
}

// SetJobID sets the JobID value and returns the updated Log.
func (m *Log) SetJobID(v *string) {
	m.JobID = v
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

// GetSiteID returns the SiteID value.
func (m GetSiteErrorParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetSiteErrorParams.
func (m *GetSiteErrorParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetID returns the ID value.
func (m GetSiteEventParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetSiteEventParams.
func (m *GetSiteEventParams) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m GetSiteEventParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetSiteEventParams.
func (m *GetSiteEventParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetBatch returns the Batch value.
func (m GetSiteEventParams) GetBatch() *string {
	return m.Batch
}

// SetBatch sets the Batch value and returns the updated GetSiteEventParams.
func (m *GetSiteEventParams) SetBatch(v *string) {
	m.Batch = v
}

// GetID returns the ID value.
func (m GetSiteLogParams) GetID() int64 {
	return m.ID
}

// SetID sets the ID value and returns the updated GetSiteLogParams.
func (m *GetSiteLogParams) SetID(v int64) {
	m.ID = v
}

// GetSiteID returns the SiteID value.
func (m GetSiteLogParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated GetSiteLogParams.
func (m *GetSiteLogParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetSiteID returns the SiteID value.
func (m InsertErrorParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated InsertErrorParams.
func (m *InsertErrorParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetMessage returns the Message value.
func (m InsertErrorParams) GetMessage() string {
	return m.Message
}

// SetMessage sets the Message value and returns the updated InsertErrorParams.
func (m *InsertErrorParams) SetMessage(v string) {
	m.Message = v
}

// GetAttributes returns the Attributes value.
func (m InsertErrorParams) GetAttributes() *string {
	return m.Attributes
}

// SetAttributes sets the Attributes value and returns the updated InsertErrorParams.
func (m *InsertErrorParams) SetAttributes(v *string) {
	m.Attributes = v
}

// GetRequestID returns the RequestID value.
func (m InsertErrorParams) GetRequestID() *string {
	return m.RequestID
}

// SetRequestID sets the RequestID value and returns the updated InsertErrorParams.
func (m *InsertErrorParams) SetRequestID(v *string) {
	m.RequestID = v
}

// GetJobID returns the JobID value.
func (m InsertErrorParams) GetJobID() *string {
	return m.JobID
}

// SetJobID sets the JobID value and returns the updated InsertErrorParams.
func (m *InsertErrorParams) SetJobID(v *string) {
	m.JobID = v
}

// GetSiteID returns the SiteID value.
func (m InsertEventParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated InsertEventParams.
func (m *InsertEventParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetBatch returns the Batch value.
func (m InsertEventParams) GetBatch() *string {
	return m.Batch
}

// SetBatch sets the Batch value and returns the updated InsertEventParams.
func (m *InsertEventParams) SetBatch(v *string) {
	m.Batch = v
}

// GetName returns the Name value.
func (m InsertEventParams) GetName() string {
	return m.Name
}

// SetName sets the Name value and returns the updated InsertEventParams.
func (m *InsertEventParams) SetName(v string) {
	m.Name = v
}

// GetAttributes returns the Attributes value.
func (m InsertEventParams) GetAttributes() *string {
	return m.Attributes
}

// SetAttributes sets the Attributes value and returns the updated InsertEventParams.
func (m *InsertEventParams) SetAttributes(v *string) {
	m.Attributes = v
}

// GetStartedAt returns the StartedAt value.
func (m InsertEventParams) GetStartedAt() int64 {
	return m.StartedAt
}

// SetStartedAt sets the StartedAt value and returns the updated InsertEventParams.
func (m *InsertEventParams) SetStartedAt(v int64) {
	m.StartedAt = v
}

// GetFinishedAt returns the FinishedAt value.
func (m InsertEventParams) GetFinishedAt() *int64 {
	return m.FinishedAt
}

// SetFinishedAt sets the FinishedAt value and returns the updated InsertEventParams.
func (m *InsertEventParams) SetFinishedAt(v *int64) {
	m.FinishedAt = v
}

// GetRequestID returns the RequestID value.
func (m InsertEventParams) GetRequestID() *string {
	return m.RequestID
}

// SetRequestID sets the RequestID value and returns the updated InsertEventParams.
func (m *InsertEventParams) SetRequestID(v *string) {
	m.RequestID = v
}

// GetJobID returns the JobID value.
func (m InsertEventParams) GetJobID() *string {
	return m.JobID
}

// SetJobID sets the JobID value and returns the updated InsertEventParams.
func (m *InsertEventParams) SetJobID(v *string) {
	m.JobID = v
}

// GetSiteID returns the SiteID value.
func (m InsertLogParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated InsertLogParams.
func (m *InsertLogParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetLevel returns the Level value.
func (m InsertLogParams) GetLevel() string {
	return m.Level
}

// SetLevel sets the Level value and returns the updated InsertLogParams.
func (m *InsertLogParams) SetLevel(v string) {
	m.Level = v
}

// GetMessage returns the Message value.
func (m InsertLogParams) GetMessage() string {
	return m.Message
}

// SetMessage sets the Message value and returns the updated InsertLogParams.
func (m *InsertLogParams) SetMessage(v string) {
	m.Message = v
}

// GetAttributes returns the Attributes value.
func (m InsertLogParams) GetAttributes() *string {
	return m.Attributes
}

// SetAttributes sets the Attributes value and returns the updated InsertLogParams.
func (m *InsertLogParams) SetAttributes(v *string) {
	m.Attributes = v
}

// GetRequestID returns the RequestID value.
func (m InsertLogParams) GetRequestID() *string {
	return m.RequestID
}

// SetRequestID sets the RequestID value and returns the updated InsertLogParams.
func (m *InsertLogParams) SetRequestID(v *string) {
	m.RequestID = v
}

// GetJobID returns the JobID value.
func (m InsertLogParams) GetJobID() *string {
	return m.JobID
}

// SetJobID sets the JobID value and returns the updated InsertLogParams.
func (m *InsertLogParams) SetJobID(v *string) {
	m.JobID = v
}

// GetCreatedAt returns the CreatedAt value.
func (m InsertLogParams) GetCreatedAt() int64 {
	return m.CreatedAt
}

// SetCreatedAt sets the CreatedAt value and returns the updated InsertLogParams.
func (m *InsertLogParams) SetCreatedAt(v int64) {
	m.CreatedAt = v
}

// GetSiteID returns the SiteID value.
func (m RecentConnectorErrorsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated RecentConnectorErrorsParams.
func (m *RecentConnectorErrorsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetConnectorType returns the ConnectorType value.
func (m RecentConnectorErrorsParams) GetConnectorType() *string {
	return m.ConnectorType
}

// SetConnectorType sets the ConnectorType value and returns the updated RecentConnectorErrorsParams.
func (m *RecentConnectorErrorsParams) SetConnectorType(v *string) {
	m.ConnectorType = v
}

// GetLimit returns the Limit value.
func (m RecentConnectorErrorsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated RecentConnectorErrorsParams.
func (m *RecentConnectorErrorsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m RecentConnectorEventsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated RecentConnectorEventsParams.
func (m *RecentConnectorEventsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetConnectorType returns the ConnectorType value.
func (m RecentConnectorEventsParams) GetConnectorType() *string {
	return m.ConnectorType
}

// SetConnectorType sets the ConnectorType value and returns the updated RecentConnectorEventsParams.
func (m *RecentConnectorEventsParams) SetConnectorType(v *string) {
	m.ConnectorType = v
}

// GetLimit returns the Limit value.
func (m RecentConnectorEventsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated RecentConnectorEventsParams.
func (m *RecentConnectorEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m SearchErrorsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SearchErrorsParams.
func (m *SearchErrorsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetType returns the Type value.
func (m SearchErrorsParams) GetType() *string {
	return m.Type
}

// SetType sets the Type value and returns the updated SearchErrorsParams.
func (m *SearchErrorsParams) SetType(v *string) {
	m.Type = v
}

// GetSearch returns the Search value.
func (m SearchErrorsParams) GetSearch() *string {
	return m.Search
}

// SetSearch sets the Search value and returns the updated SearchErrorsParams.
func (m *SearchErrorsParams) SetSearch(v *string) {
	m.Search = v
}

// GetRequestID returns the RequestID value.
func (m SearchErrorsParams) GetRequestID() *string {
	return m.RequestID
}

// SetRequestID sets the RequestID value and returns the updated SearchErrorsParams.
func (m *SearchErrorsParams) SetRequestID(v *string) {
	m.RequestID = v
}

// GetJobID returns the JobID value.
func (m SearchErrorsParams) GetJobID() *string {
	return m.JobID
}

// SetJobID sets the JobID value and returns the updated SearchErrorsParams.
func (m *SearchErrorsParams) SetJobID(v *string) {
	m.JobID = v
}

// GetSkip returns the Skip value.
func (m SearchErrorsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchErrorsParams.
func (m *SearchErrorsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchErrorsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchErrorsParams.
func (m *SearchErrorsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m SearchEventsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SearchEventsParams.
func (m *SearchEventsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetType returns the Type value.
func (m SearchEventsParams) GetType() *string {
	return m.Type
}

// SetType sets the Type value and returns the updated SearchEventsParams.
func (m *SearchEventsParams) SetType(v *string) {
	m.Type = v
}

// GetSearch returns the Search value.
func (m SearchEventsParams) GetSearch() *string {
	return m.Search
}

// SetSearch sets the Search value and returns the updated SearchEventsParams.
func (m *SearchEventsParams) SetSearch(v *string) {
	m.Search = v
}

// GetRequestID returns the RequestID value.
func (m SearchEventsParams) GetRequestID() *string {
	return m.RequestID
}

// SetRequestID sets the RequestID value and returns the updated SearchEventsParams.
func (m *SearchEventsParams) SetRequestID(v *string) {
	m.RequestID = v
}

// GetJobID returns the JobID value.
func (m SearchEventsParams) GetJobID() *string {
	return m.JobID
}

// SetJobID sets the JobID value and returns the updated SearchEventsParams.
func (m *SearchEventsParams) SetJobID(v *string) {
	m.JobID = v
}

// GetSkip returns the Skip value.
func (m SearchEventsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchEventsParams.
func (m *SearchEventsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchEventsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchEventsParams.
func (m *SearchEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m SearchLogsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SearchLogsParams.
func (m *SearchLogsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetSearch returns the Search value.
func (m SearchLogsParams) GetSearch() *string {
	return m.Search
}

// SetSearch sets the Search value and returns the updated SearchLogsParams.
func (m *SearchLogsParams) SetSearch(v *string) {
	m.Search = v
}

// GetRequestID returns the RequestID value.
func (m SearchLogsParams) GetRequestID() *string {
	return m.RequestID
}

// SetRequestID sets the RequestID value and returns the updated SearchLogsParams.
func (m *SearchLogsParams) SetRequestID(v *string) {
	m.RequestID = v
}

// GetJobID returns the JobID value.
func (m SearchLogsParams) GetJobID() *string {
	return m.JobID
}

// SetJobID sets the JobID value and returns the updated SearchLogsParams.
func (m *SearchLogsParams) SetJobID(v *string) {
	m.JobID = v
}

// GetSkip returns the Skip value.
func (m SearchLogsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchLogsParams.
func (m *SearchLogsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchLogsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchLogsParams.
func (m *SearchLogsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m SearchSiteErrorsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SearchSiteErrorsParams.
func (m *SearchSiteErrorsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetSearch returns the Search value.
func (m SearchSiteErrorsParams) GetSearch() *string {
	return m.Search
}

// SetSearch sets the Search value and returns the updated SearchSiteErrorsParams.
func (m *SearchSiteErrorsParams) SetSearch(v *string) {
	m.Search = v
}

// GetType returns the Type value.
func (m SearchSiteErrorsParams) GetType() *string {
	return m.Type
}

// SetType sets the Type value and returns the updated SearchSiteErrorsParams.
func (m *SearchSiteErrorsParams) SetType(v *string) {
	m.Type = v
}

// GetSkip returns the Skip value.
func (m SearchSiteErrorsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchSiteErrorsParams.
func (m *SearchSiteErrorsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchSiteErrorsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchSiteErrorsParams.
func (m *SearchSiteErrorsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSkip returns the Skip value.
func (m SearchSiteEventsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchSiteEventsParams.
func (m *SearchSiteEventsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchSiteEventsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchSiteEventsParams.
func (m *SearchSiteEventsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetSiteID returns the SiteID value.
func (m SearchSiteEventsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SearchSiteEventsParams.
func (m *SearchSiteEventsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetSearch returns the Search value.
func (m SearchSiteEventsParams) GetSearch() *string {
	return m.Search
}

// SetSearch sets the Search value and returns the updated SearchSiteEventsParams.
func (m *SearchSiteEventsParams) SetSearch(v *string) {
	m.Search = v
}

// GetType returns the Type value.
func (m SearchSiteEventsParams) GetType() *string {
	return m.Type
}

// SetType sets the Type value and returns the updated SearchSiteEventsParams.
func (m *SearchSiteEventsParams) SetType(v *string) {
	m.Type = v
}

// GetBatch returns the Batch value.
func (m SearchSiteEventsParams) GetBatch() *string {
	return m.Batch
}

// SetBatch sets the Batch value and returns the updated SearchSiteEventsParams.
func (m *SearchSiteEventsParams) SetBatch(v *string) {
	m.Batch = v
}

// GetSiteID returns the SiteID value.
func (m SearchSiteLogsParams) GetSiteID() *string {
	return m.SiteID
}

// SetSiteID sets the SiteID value and returns the updated SearchSiteLogsParams.
func (m *SearchSiteLogsParams) SetSiteID(v *string) {
	m.SiteID = v
}

// GetSearch returns the Search value.
func (m SearchSiteLogsParams) GetSearch() *string {
	return m.Search
}

// SetSearch sets the Search value and returns the updated SearchSiteLogsParams.
func (m *SearchSiteLogsParams) SetSearch(v *string) {
	m.Search = v
}

// GetSkip returns the Skip value.
func (m SearchSiteLogsParams) GetSkip() int64 {
	return m.Skip
}

// SetSkip sets the Skip value and returns the updated SearchSiteLogsParams.
func (m *SearchSiteLogsParams) SetSkip(v int64) {
	m.Skip = v
}

// GetLimit returns the Limit value.
func (m SearchSiteLogsParams) GetLimit() int64 {
	return m.Limit
}

// SetLimit sets the Limit value and returns the updated SearchSiteLogsParams.
func (m *SearchSiteLogsParams) SetLimit(v int64) {
	m.Limit = v
}

// GetKeep returns the Keep value.
func (m TrimErrorsParams) GetKeep() int64 {
	return m.Keep
}

// SetKeep sets the Keep value and returns the updated TrimErrorsParams.
func (m *TrimErrorsParams) SetKeep(v int64) {
	m.Keep = v
}

// GetKeep returns the Keep value.
func (m TrimEventsParams) GetKeep() int64 {
	return m.Keep
}

// SetKeep sets the Keep value and returns the updated TrimEventsParams.
func (m *TrimEventsParams) SetKeep(v int64) {
	m.Keep = v
}

// GetKeep returns the Keep value.
func (m TrimLogsParams) GetKeep() int64 {
	return m.Keep
}

// SetKeep sets the Keep value and returns the updated TrimLogsParams.
func (m *TrimLogsParams) SetKeep(v int64) {
	m.Keep = v
}

