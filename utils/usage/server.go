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

// Server is the server process's resources now: CPU in its latest sample, of a few seconds, as htop shows it, a percent
// of one core, memory in bytes, the size of the DBs and the space left on their disk. ResidentBytes is all the memory the OS counts
// for the process, which HeapBytes and SysBytes, Go's own, leave out SQLite's from.
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

	// ResidentBytes is 0 when the OS cannot say
	ResidentBytes uint64 `json:"resident_bytes"`
}

// Server reads the process's resources, the size of the files in dbDir, and the space left on its disk
func (t *Tracker) Server(dbDir string, now time.Time) Server {
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	server := Server{CPUs: runtime.NumCPU(), Goroutines: runtime.NumGoroutine(), HeapBytes: memory.HeapAlloc, SysBytes: memory.Sys, GCs: memory.NumGC}
	server.ResidentBytes, _ = residentBytes()
	if t == nil {
		return server
	}
	server.UptimeMs = now.Sub(t.started).Milliseconds()

	t.mu.Lock()
	server.CPUPercent = t.watch.latest
	t.mu.Unlock()

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
