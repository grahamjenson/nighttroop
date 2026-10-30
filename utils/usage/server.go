package usage

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

type cpuSample struct {
	at   time.Time
	used time.Duration
}

// Server is the server process's resources now: CPU since the last time they were read, as a percent of every core,
// memory in bytes, the size of the DBs and the space left on their disk
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

// Server reads the process's resources, the size of the files in dbDir, and the space left on its disk
func (t *Tracker) Server(dbDir string, now time.Time) Server {
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

	// The DB folder's disk, or the working folder's before the DB folder is made
	server.DiskUsed = dirSize(dbDir)
	var disk syscall.Statfs_t
	if syscall.Statfs(dbDir, &disk) == nil || syscall.Statfs(".", &disk) == nil {
		server.DiskFree = disk.Bavail * uint64(disk.Bsize)
	}
	return server
}

// dirSize is the size of the files in dir and its folders, skipping any that cannot be read, such as a DB deleted
// while it is walked
func dirSize(dir string) uint64 {
	var size uint64
	filepath.WalkDir(dir, func(_ string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			if info, err := entry.Info(); err == nil {
				size += uint64(info.Size())
			}
		}
		return nil
	})
	return size
}
