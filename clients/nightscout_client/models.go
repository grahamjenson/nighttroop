package nightscout_client

import (
	"encoding/json"
	"strings"
	"time"
)

const MmolConversionFactor = 0.0555

//go:generate go tool enumer -type Unit -linecomment -text

type Unit uint8

const (
	UnitMgdl Unit = iota // mg/dL
	UnitMmol             // mmol/L
)

type Mgdl int

func (m Mgdl) Mgdl() int { return int(m) }

func (m Mgdl) Mmol() float64 { return float64(m) * MmolConversionFactor }

type Mills struct {
	time.Time
}

func (m *Mills) UnmarshalJSON(bytes []byte) error {
	var mills int64
	if err := json.Unmarshal(bytes, &mills); err != nil {
		return err
	}
	m.Time = time.UnixMilli(mills)
	return nil
}

func (m *Mills) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.UnixMilli())
}

func (m *Mills) Relative(round bool) string {
	if m.Unix() == 0 {
		return ""
	}

	// Drop resolution to minutes
	duration := time.Since(m.Time)
	if round {
		duration = duration.Round(time.Minute)
	} else {
		duration = duration.Truncate(time.Minute)
	}

	str := duration.String()
	str = strings.TrimSuffix(str, "0s")
	if str == "" {
		str = "0m"
	}
	return str
}

type SGV struct {
	ID         string      `json:"_id"`
	Device     string      `json:"device"`
	Direction  string      `json:"direction"`
	Filtered   json.Number `json:"filtered"`
	Mgdl       Mgdl        `json:"mgdl"`
	Mills      Mills       `json:"mills"`
	Noise      json.Number `json:"noise"`
	Rssi       json.Number `json:"rssi"`
	Scaled     json.Number `json:"scaled"`
	Type       string      `json:"type"`
	Unfiltered json.Number `json:"unfiltered"`
}
