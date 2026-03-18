package utils

import (
	"fmt"
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
		fmt.Printf("%s took %v (%s)\n", name, elapsed, *info)
	} else {
		fmt.Printf("%s took %v\n", name, elapsed)
	}

}
