// Package page tells the server's admin of what needs them now, such as a site that cannot be restored
package page

import (
	"context"
	"log/slog"
)

// Page tells the admin message, with args as slog's. Paging is not set up yet, so it only logs, as an error, that it
// would page.
func Page(ctx context.Context, message string, args ...any) {
	slog.ErrorContext(ctx, "Paging is not set up yet, so the admin is not paged: "+message, args...)
}
