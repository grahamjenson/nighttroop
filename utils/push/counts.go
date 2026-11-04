package push

import (
	"sync"
	"time"
)

// countDays is how many days of push counts are kept, enough for a week from Monday
const countDays = 8

// Count is how many pushes were sent to a push service, and how many of them failed
type Count struct {
	Sent   int64 `json:"sent"`
	Failed int64 `json:"failed"`
}

// Counts are the pushes since midnight and since Monday, in UTC, counted in memory since the server started
type Counts struct {
	Today Count `json:"today"`
	Week  Count `json:"week"`
}

// counts are the pushes each UTC day, by its number since 1970, for the last countDays
var counts = struct {
	sync.Mutex
	days map[int64]Count
}{days: map[int64]Count{}}

// utcDay is the day's number since 1970 in UTC
func utcDay(t time.Time) int64 {
	return t.UTC().Unix() / (24 * 60 * 60)
}

// countPush adds a push to now's day, as failed when it had an error, dropping days older than countDays
func countPush(now time.Time, err error) {
	counts.Lock()
	defer counts.Unlock()
	today := utcDay(now)
	count := counts.days[today]
	count.Sent++
	if err != nil {
		count.Failed++
	}
	counts.days[today] = count
	for day := range counts.days {
		if day <= today-countDays {
			delete(counts.days, day)
		}
	}
}

// CountsAt are the pushes since now's midnight and since its Monday, in UTC
func CountsAt(now time.Time) Counts {
	counts.Lock()
	defer counts.Unlock()
	today := utcDay(now)
	monday := today - int64(now.UTC().Weekday()+6)%7
	var c Counts
	for day, count := range counts.days {
		if day == today {
			c.Today = count
		}
		if day >= monday && day <= today {
			c.Week.Sent += count.Sent
			c.Week.Failed += count.Failed
		}
	}
	return c
}
