package sitemodels

import (
	"encoding/json"

	"nighttroop.t1d.nz/data/sql/logsdb"
)

// SiteLog is a message logged about the site
type SiteLog struct {
	ID         int64           `json:"id"`
	Level      string          `json:"level"`
	Message    string          `json:"message"`
	Attributes json.RawMessage `json:"attributes"`
	CreatedAt  int64           `json:"created_at"`
	// RequestID and JobID are the request or background job run it was for, to find it by
	RequestID *string `json:"request_id"`
	JobID     *string `json:"job_id"`
}

// SiteEvent is something that happened to the site, from when it started until it finished when it took time
type SiteEvent struct {
	ID         int64           `json:"id"`
	Name       string          `json:"name"`
	Attributes json.RawMessage `json:"attributes"`
	StartedAt  int64           `json:"started_at"`
	FinishedAt *int64          `json:"finished_at"`
	// RequestID and JobID are the request or background job run it was for, to find it by
	RequestID *string `json:"request_id"`
	JobID     *string `json:"job_id"`
}

// SiteError is something that failed for the site
type SiteError struct {
	ID         int64           `json:"id"`
	Message    string          `json:"message"`
	Attributes json.RawMessage `json:"attributes"`
	CreatedAt  int64           `json:"created_at"`
	// RequestID and JobID are the request or background job run it was for, to find it by
	RequestID *string `json:"request_id"`
	JobID     *string `json:"job_id"`
}

func SiteLogFromDB(row *logsdb.Log) *SiteLog {
	return &SiteLog{ID: row.ID, Level: row.Level, Message: row.Message, Attributes: attributesJSON(row.Attributes), CreatedAt: row.CreatedAt,
		RequestID: row.RequestID, JobID: row.JobID}
}

func SiteEventFromDB(row *logsdb.Event) *SiteEvent {
	return &SiteEvent{ID: row.ID, Name: row.Name, Attributes: attributesJSON(row.Attributes), StartedAt: row.StartedAt, FinishedAt: row.FinishedAt,
		RequestID: row.RequestID, JobID: row.JobID}
}

func SiteErrorFromDB(row *logsdb.Error) *SiteError {
	return &SiteError{ID: row.ID, Message: row.Message, Attributes: attributesJSON(row.Attributes), CreatedAt: row.CreatedAt,
		RequestID: row.RequestID, JobID: row.JobID}
}

// attributesJSON is a row's attributes as JSON, or null when it has none
func attributesJSON(attributes *string) json.RawMessage {
	if attributes == nil {
		return nil
	}
	return json.RawMessage(*attributes)
}
