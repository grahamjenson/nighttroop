package nightscout

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type NSParams struct {
	Count           int
	Skip            int
	Token           string
	EventType       string
	SGV             int
	Direction       string
	Device          string
	EnteredBy       string
	DateGteDuration time.Duration
}

// parseMSDuration converts Nightscout's Unix millisecond strings into a relative duration
func parseMSDuration(val string, now time.Time) (time.Duration, error) {
	ms, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	t := time.Unix(0, ms*int64(time.Millisecond))
	return now.Sub(t), nil
}

// parseISODuration converts ISO8601 strings into a relative duration
func parseISODuration(val string, now time.Time) (time.Duration, error) {
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return 0, err
	}
	return now.Sub(t), nil
}

func ExtractNSParams(r *http.Request) NSParams {
	nsp := NSParams{}
	query := r.URL.Query()
	now := time.Now()

	for key, values := range query {
		if len(values) == 0 {
			continue
		}
		val := values[0]
		cleanKey := strings.ReplaceAll(key, "\"", "")

		switch {
		case cleanKey == "count":
			if c, err := strconv.Atoi(val); err != nil {
				log.Printf("error parsing count: %v", err)
			} else {
				nsp.Count = c
			}

		case cleanKey == "skip":
			if s, err := strconv.Atoi(val); err != nil {
				log.Printf("error parsing skip: %v", err)
			} else {
				nsp.Skip = s
			}

		case cleanKey == "token":
			nsp.Token = val

		case strings.Contains(cleanKey, "find[eventType]"):
			nsp.EventType = val

		case strings.Contains(cleanKey, "find[sgv]"):
			if s, err := strconv.Atoi(val); err != nil {
				log.Printf("error parsing sgv: %v", err)
			} else {
				nsp.SGV = s
			}

		case strings.Contains(cleanKey, "find[direction]"):
			nsp.Direction = val

		case strings.Contains(cleanKey, "find[device]"):
			nsp.Device = val

		case strings.Contains(cleanKey, "find[enteredBy]"):
			nsp.EnteredBy = val

		case strings.Contains(cleanKey, "find[date][$gte]"):
			d, err := parseMSDuration(val, now)
			if err != nil {
				log.Printf("error parsing date[$gte]: %v", err)
			} else {
				nsp.DateGteDuration = d
			}

		case strings.Contains(cleanKey, "find[dateString][$gte]"):
			d, err := parseISODuration(val, now)
			if err != nil {
				log.Printf("error parsing dateString[$gte]: %v", err)
			} else {
				nsp.DateGteDuration = d
			}

		default:
			log.Printf("unused query parameter: %s=%s", key, val)
		}

	}

	return nsp
}
