package utils

import (
	"fmt"
	"time"
)

func TrackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %v\n", name, elapsed)
}
