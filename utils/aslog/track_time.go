package aslog

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// TrackTime logs how long since start, with ctx's attributes, unless log is false
func TrackTime(ctx context.Context, start time.Time, name string, log *bool, info *string) {

	elapsed := time.Since(start)
	if log != nil && !*log {
		return
	}
	// log is true or nil here

	if info != nil {
		slog.InfoContext(ctx, fmt.Sprintf("Timer: %s", name), "elapsed", elapsed.Milliseconds(), "info", *info)
	} else {
		slog.InfoContext(ctx, fmt.Sprintf("Timer: %s", name), "elapsed", elapsed.Milliseconds())
	}

}
