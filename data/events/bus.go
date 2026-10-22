// Package events is a publish and subscribe bus: handlers subscribe to a key, such as a site's ID,
// and are called with every event published under it
package events

import (
	"log/slog"
	"sync"
)

// Bus delivers each event published under a key to the handlers subscribed to that key
type Bus[K comparable, E any] struct {
	mu       sync.RWMutex
	handlers map[K]map[uint64]func(E)
	next     uint64
}

func NewBus[K comparable, E any]() *Bus[K, E] {
	return &Bus[K, E]{handlers: map[K]map[uint64]func(E){}}
}

// Subscribe calls handler with each event published under key, until unsubscribe is called
func (b *Bus[K, E]) Subscribe(key K, handler func(E)) (unsubscribe func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.next++
	id := b.next
	if b.handlers[key] == nil {
		b.handlers[key] = map[uint64]func(E){}
	}
	b.handlers[key][id] = handler
	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.handlers[key], id)
		if len(b.handlers[key]) == 0 {
			delete(b.handlers, key)
		}
	}
}

// Publish calls each handler subscribed to key in the publisher's goroutine, so a handler must not block.
// A handler that panics is logged, and the rest are still called.
func (b *Bus[K, E]) Publish(key K, event E) {
	b.mu.RLock()
	handlers := make([]func(E), 0, len(b.handlers[key]))
	for _, handler := range b.handlers[key] {
		handlers = append(handlers, handler)
	}
	b.mu.RUnlock()
	for _, handler := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Event handler panicked", "key", key, "panic", r)
				}
			}()
			handler(event)
		}()
	}
}
