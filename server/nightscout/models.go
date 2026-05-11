package nightscout

import "time"

type NSEntry struct {
	ID        string     `json:"_id"`
	Date      *time.Time `json:"date"`
	SGV       int        `json:"sgv"`
	Type      string     `json:"type"`
	Direction string     `json:"direction,omitempty"`
}

type NSProfile struct {
	ID    string `json:"_id"`
	Units string `json:"units"`
}
