package controlmodels

import "encoding/json"

// Log is a message logged from around the server, optionally about a site
type Log struct {
	ID         int64           `json:"id"`
	SiteID     *string         `json:"site_id"`
	Level      string          `json:"level"`
	Message    string          `json:"message"`
	Attributes json.RawMessage `json:"attributes"`
	CreatedAt  int64           `json:"created_at"`
}

// Event is something that happened, optionally to a site, from when it started until it finished when it took time
type Event struct {
	ID         int64           `json:"id"`
	SiteID     *string         `json:"site_id"`
	Name       string          `json:"name"`
	Attributes json.RawMessage `json:"attributes"`
	StartedAt  int64           `json:"started_at"`
	FinishedAt *int64          `json:"finished_at"`
}

// Error is something that failed, optionally for a site
type Error struct {
	ID         int64           `json:"id"`
	SiteID     *string         `json:"site_id"`
	Message    string          `json:"message"`
	Attributes json.RawMessage `json:"attributes"`
	CreatedAt  int64           `json:"created_at"`
}

// AttributesJSON is a row's attributes as JSON, or null when it has none
func AttributesJSON(attributes *string) json.RawMessage {
	if attributes == nil {
		return nil
	}
	return json.RawMessage(*attributes)
}
