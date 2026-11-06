package aslog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

// With is ctx with attrs logged too, replacing those of the same key, for it and what is made from it but not the
// context it was made from, so work done at once for many sites each logs its own site
func With(ctx context.Context, attrs ...slog.Attr) context.Context {
	held := attrsOf(ctx)
	for _, attr := range attrs {
		held = append(deleteKey(held, attr.Key), attr)
	}
	return NewContext(ctx, held...)
}

// WithSite is ctx logging the site its work is for
func WithSite(ctx context.Context, siteID string) context.Context {
	return With(ctx, slog.String("site_id", siteID))
}

// NewJob is ctx for one run of a background job: what is logged with it says the job and a job_id new to this run
func NewJob(ctx context.Context, job string, attrs ...slog.Attr) context.Context {
	id := make([]byte, 8)
	_, _ = rand.Read(id)
	return With(ctx, append([]slog.Attr{slog.String("job_id", hex.EncodeToString(id)), slog.String("job", job)}, attrs...)...)
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

// Attrs are what is logged with ctx, such as its request's site and identity once they are known
func Attrs(ctx context.Context) []slog.Attr {
	return attrsOf(ctx)
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
