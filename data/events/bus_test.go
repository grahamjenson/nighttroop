package events

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBus(t *testing.T) {
	bus := NewBus[string, int]()
	var a, b []int
	unsubscribeA := bus.Subscribe("site-a", func(event int) { a = append(a, event) })
	bus.Subscribe("site-b", func(event int) { b = append(b, event) })

	// Each key's handlers get only its events
	bus.Publish("site-a", 1)
	bus.Publish("site-b", 2)
	bus.Publish("site-c", 3)
	require.Equal(t, []int{1}, a)
	require.Equal(t, []int{2}, b)

	// After unsubscribing a handler gets nothing more
	unsubscribeA()
	bus.Publish("site-a", 4)
	require.Equal(t, []int{1}, a)
	require.Empty(t, bus.handlers["site-a"], "a key with no handlers is dropped")
}

func TestBusHandlerPanicDoesNotStopTheOthers(t *testing.T) {
	bus := NewBus[string, int]()
	var got []int
	bus.Subscribe("site", func(int) { panic("broken handler") })
	bus.Subscribe("site", func(event int) { got = append(got, event) })
	require.NotPanics(t, func() { bus.Publish("site", 1) })
	require.Equal(t, []int{1}, got)
}

func TestBusIsSafeToUseConcurrently(t *testing.T) {
	bus := NewBus[string, int]()
	var mu sync.Mutex
	count := 0
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			unsubscribe := bus.Subscribe("site", func(int) {
				mu.Lock()
				count++
				mu.Unlock()
			})
			bus.Publish("site", 1)
			unsubscribe()
		})
	}
	wg.Wait()
	require.Positive(t, count)
}
