package aslog

import (
	"context"
	"log/slog"
	"sync"
)

// contextAttrs are the attributes logged with everything done with a context, such as a request's ID, then its site
// and identity once known; they are shared, so ones added deeper in a request show on its outer logs too
type contextAttrs struct {
	mu    sync.Mutex
	attrs []slog.Attr
}

type contextAttrsKey struct{}

// NewContext is ctx holding attrs, logged with everything logged with it, or anything made from it, from now on
func NewContext(ctx context.Context, attrs ...slog.Attr) context.Context {
	return context.WithValue(ctx, contextAttrsKey{}, &contextAttrs{attrs: attrs})
}

// AddAttrs adds attrs to what is logged with ctx, if it was made with NewContext, replacing those of the same key
func AddAttrs(ctx context.Context, attrs ...slog.Attr) {
	held, ok := ctx.Value(contextAttrsKey{}).(*contextAttrs)
	if !ok {
		return
	}
	held.mu.Lock()
	defer held.mu.Unlock()
	for _, attr := range attrs {
		held.attrs = append(deleteKey(held.attrs, attr.Key), attr)
	}
}

func deleteKey(attrs []slog.Attr, key string) []slog.Attr {
	kept := attrs[:0]
	for _, attr := range attrs {
		if attr.Key != key {
			kept = append(kept, attr)
		}
	}
	return kept
}

func attrsOf(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	held, ok := ctx.Value(contextAttrsKey{}).(*contextAttrs)
	if !ok {
		return nil
	}
	held.mu.Lock()
	defer held.mu.Unlock()
	return append([]slog.Attr(nil), held.attrs...)
}

// ContextHandler logs each record with the attributes of its context, so slog.ErrorContext(ctx, ...) in a request
// says which request, site and identity it was for
type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	record.AddAttrs(attrsOf(ctx)...)
	return h.Handler.Handle(ctx, record)
}

func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ContextHandler{h.Handler.WithAttrs(attrs)}
}

func (h ContextHandler) WithGroup(name string) slog.Handler {
	return ContextHandler{h.Handler.WithGroup(name)}
}
