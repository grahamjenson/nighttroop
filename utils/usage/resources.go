package usage

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// peggedPercent is the CPU a sample counts as pegged at, as one rarely reads exactly 100%
const peggedPercent = 95

// cpuWatch is the process's CPU sampled every few seconds between checks of the server's resources
type cpuWatch struct {
	last        cpuSample
	peggedSince time.Time
	longest     time.Duration
	sum         float64
	samples     int
}

// processCPU is the CPU time the process has used
func processCPU() (time.Duration, bool) {
	var rusage syscall.Rusage
	if syscall.Getrusage(syscall.RUSAGE_SELF, &rusage) != nil {
		return 0, false
	}
	return time.Duration(rusage.Utime.Nano() + rusage.Stime.Nano()), true
}

// SampleCPU samples the process's CPU since its last sample, as a percent of every core, keeping how long it has been
// pegged
func (t *Tracker) SampleCPU(now time.Time, cpus int) {
	used, ok := processCPU()
	if t == nil || !ok || cpus < 1 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	w := &t.watch
	if !w.last.at.IsZero() && now.After(w.last.at) {
		w.record(100*float64(used-w.last.used)/float64(now.Sub(w.last.at))/float64(cpus), w.last.at, now)
	}
	w.last = cpuSample{at: now, used: used}
}

// record is a sample of percent from one time to the next
func (w *cpuWatch) record(percent float64, from time.Time, to time.Time) {
	w.sum += percent
	w.samples++
	switch {
	case percent < peggedPercent:
		w.peggedSince = time.Time{}
	case w.peggedSince.IsZero():
		w.peggedSince = from
	}
	if !w.peggedSince.IsZero() {
		w.longest = max(w.longest, to.Sub(w.peggedSince))
	}
}

// TakeCPU is the mean CPU percent sampled and the longest it was pegged since it was last taken
func (t *Tracker) TakeCPU() (meanPercent float64, longestPegged time.Duration) {
	if t == nil {
		return 0, 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	w := &t.watch
	if w.samples > 0 {
		meanPercent = w.sum / float64(w.samples)
	}
	longestPegged = w.longest
	w.sum, w.samples, w.longest = 0, 0, 0
	return meanPercent, longestPegged
}

// Memory is the bytes of memory used, of the container's limit when it has one, or else of the machine's
func Memory() (used float64, total float64, ok bool) {
	return memory("/sys/fs/cgroup", "/proc/meminfo")
}

func memory(cgroupDir string, meminfoPath string) (float64, float64, bool) {
	limit, limitErr := readUint(cgroupDir + "/memory.max")
	current, currentErr := readUint(cgroupDir + "/memory.current")
	if limitErr == nil && currentErr == nil && limit > 0 {
		return float64(current), float64(limit), true
	}
	// No limit ("max"), or not a cgroup v2 container: the machine's memory, less what it can still give
	raw, err := os.ReadFile(meminfoPath)
	if err != nil {
		return 0, 0, false
	}
	var total, available float64
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, _ := strconv.ParseFloat(fields[1], 64)
		switch fields[0] {
		case "MemTotal:":
			total = value
		case "MemAvailable:":
			available = value
		}
	}
	if total == 0 {
		return 0, 0, false
	}
	return 1024 * (total - available), 1024 * total, true
}

func readUint(path string) (uint64, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)
}

// Disk is the bytes used on dir's disk, of its total, as df counts them: the space not kept for root
func Disk(dir string) (used float64, total float64, ok bool) {
	var disk syscall.Statfs_t
	if syscall.Statfs(dir, &disk) != nil {
		return 0, 0, false
	}
	used = float64(disk.Blocks-disk.Bfree) * float64(disk.Bsize)
	if total = used + float64(disk.Bavail)*float64(disk.Bsize); total > 0 {
		return used, total, true
	}
	return 0, 0, false
}
