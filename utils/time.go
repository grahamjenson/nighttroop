package utils

import (
	"fmt"
	"log/slog"
	"time"
)

// If Log is nil, ignore it
// If log is false, the dont log
func TrackTime(start time.Time, name string, log *bool, info *string) {

	elapsed := time.Since(start)
	if log != nil && !*log {
		return
	}
	// log is true or nil here

	if info != nil {
		slog.Info(fmt.Sprintf("Timer: %s", name), "elapsed", elapsed.Milliseconds(), "info", *info)
	} else {
		slog.Info(fmt.Sprintf("Timer: %s", name), "elapsed", elapsed.Milliseconds())
	}

}
