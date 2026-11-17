// Package page tells the server's admin of a problem that needs them now, such as sites whose DBs cannot be restored
package page

import (
	"context"
	"log/slog"
)

// Func pages the admin with a message and what it is about, as slog attributes
type Func func(ctx context.Context, message string, args ...any)

// Admin pages the server's admin. Paging is not set up yet, so it logs the page as an error, which the control site's
// errors and Sentry show.
func Admin(ctx context.Context, message string, args ...any) {
	slog.ErrorContext(ctx, "Page: "+message, args...)
}
