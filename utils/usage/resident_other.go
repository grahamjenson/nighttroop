//go:build !linux

package usage

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// residentBytes is the memory the process has in RAM, as the OS counts it, with what SQLite allocates outside Go's
// heap. Without /proc, as on a Mac, ps says it, in KB.
func residentBytes() (uint64, bool) {
	out, err := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		return 0, false
	}
	kb, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, false
	}
	return kb * 1024, true
}
