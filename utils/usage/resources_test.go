package usage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// The longest the CPU stays pegged runs across samples, starts again once it drops, and is taken once
func TestCPUWatch(t *testing.T) {
	start := time.Unix(0, 0)
	at := func(seconds int) time.Time { return start.Add(time.Duration(seconds) * time.Second) }
	tracker := New()
	w := &tracker.watch
	for i, percent := range []float64{50, 99, 100, 97, 40, 100, 100} {
		w.record(percent, 1, at(i*5), at(i*5+5))
	}
	mean, pegged := tracker.TakeCPU()
	require.InDelta(t, 83.71, mean, 0.01)
	require.Equal(t, 15*time.Second, pegged, "three pegged samples of 5 seconds")
	mean, pegged = tracker.TakeCPU()
	require.Zero(t, mean)
	require.Zero(t, pegged)

	// Still pegged after it was taken, so the run carries on from before
	w.record(100, 1, at(35), at(40))
	_, pegged = tracker.TakeCPU()
	require.Equal(t, 15*time.Second, pegged)

	// As htop shows it, a percent of one core: on 4 cores, 200% is two of them busy, and only nearly all four is pegged
	four := New()
	for i, percent := range []float64{200, 390, 395} {
		four.watch.record(percent, 4, at(i*5), at(i*5+5))
	}
	require.Equal(t, 395.0, four.Server(t.TempDir(), at(15)).CPUPercent, "the latest sample")
	mean, pegged = four.TakeCPU()
	require.InDelta(t, 328.33, mean, 0.01)
	require.Equal(t, 10*time.Second, pegged, "200% of 4 cores is not pegged")

	var none *Tracker
	none.SampleCPU(at(0), 4)
	mean, pegged = none.TakeCPU()
	require.Zero(t, mean+pegged.Seconds())
}

func TestMemory(t *testing.T) {
	write := func(dir, name, content string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	}
	meminfo := filepath.Join(t.TempDir(), "meminfo")
	write(filepath.Dir(meminfo), "meminfo", "MemTotal:       1000 kB\nMemFree:  100 kB\nMemAvailable:    250 kB\n")

	// A container's limit
	limited := t.TempDir()
	write(limited, "memory.max", "2000\n")
	write(limited, "memory.current", "1500\n")
	used, total, ok := memory(limited, meminfo)
	require.True(t, ok)
	require.Equal(t, []float64{1500, 2000}, []float64{used, total})

	// No limit, so the machine's
	unlimited := t.TempDir()
	write(unlimited, "memory.max", "max\n")
	write(unlimited, "memory.current", "1500\n")
	used, total, ok = memory(unlimited, meminfo)
	require.True(t, ok)
	require.Equal(t, []float64{750 * 1024, 1000 * 1024}, []float64{used, total}, "meminfo is in kB")

	// Neither, as on a Mac
	_, _, ok = memory(t.TempDir(), filepath.Join(t.TempDir(), "missing"))
	require.False(t, ok)
}

func TestDisk(t *testing.T) {
	used, total, ok := Disk(t.TempDir())
	require.True(t, ok)
	require.True(t, used > 0 && used <= total, used, total)
	_, _, ok = Disk(filepath.Join(t.TempDir(), "missing"))
	require.False(t, ok)
}
