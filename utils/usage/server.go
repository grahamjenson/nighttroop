package usage

import (
	"runtime"
	"syscall"
	"time"
)

type cpuSample struct {
	at   time.Time
	used time.Duration
}

// Server is the server process's resources now: CPU since the last time they were read, as a percent of every core,
// memory in bytes, and the disk the data is on
type Server struct {
	CPUPercent float64 `json:"cpu_percent"`
	CPUs       int     `json:"cpus"`
	Goroutines int     `json:"goroutines"`
	HeapBytes  uint64  `json:"heap_bytes"`
	SysBytes   uint64  `json:"sys_bytes"`
	GCs        uint32  `json:"gcs"`
	UptimeMs   int64   `json:"uptime_ms"`
	DiskUsed   uint64  `json:"disk_used"`
	DiskFree   uint64  `json:"disk_free"`
}

// Server reads the process's resources, and the disk dataDir is on
func (t *Tracker) Server(dataDir string, now time.Time) Server {
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	server := Server{CPUs: runtime.NumCPU(), Goroutines: runtime.NumGoroutine(), HeapBytes: memory.HeapAlloc, SysBytes: memory.Sys, GCs: memory.NumGC}
	if t == nil {
		return server
	}
	server.UptimeMs = now.Sub(t.started).Milliseconds()

	var rusage syscall.Rusage
	if syscall.Getrusage(syscall.RUSAGE_SELF, &rusage) == nil {
		used := time.Duration(rusage.Utime.Nano() + rusage.Stime.Nano())
		t.mu.Lock()
		last := t.cpu
		if last.at.IsZero() {
			last = cpuSample{at: t.started}
		}
		if wall := now.Sub(last.at); wall > 0 {
			server.CPUPercent = 100 * float64(used-last.used) / float64(wall) / float64(server.CPUs)
		}
		t.cpu = cpuSample{at: now, used: used}
		t.mu.Unlock()
	}

	// The data folder's disk, or the working folder's before the data folder is made
	var disk syscall.Statfs_t
	if syscall.Statfs(dataDir, &disk) == nil || syscall.Statfs(".", &disk) == nil {
		blockSize := uint64(disk.Bsize)
		server.DiskFree = disk.Bavail * blockSize
		server.DiskUsed = (disk.Blocks - disk.Bfree) * blockSize
	}
	return server
}
