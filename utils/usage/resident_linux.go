package usage

import (
	"os"
	"strconv"
	"strings"
)

// residentBytes is the memory the process has in RAM, as the OS counts it, with what SQLite allocates outside Go's
// heap: its second field of /proc/self/statm, in pages
func residentBytes() (uint64, bool) {
	raw, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0, false
	}
	fields := strings.Fields(string(raw))
	if len(fields) < 2 {
		return 0, false
	}
	pages, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return pages * uint64(os.Getpagesize()), true
}
