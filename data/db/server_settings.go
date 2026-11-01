package db

import (
	"context"
	"log/slog"
)

// PublicSiteCreation is whether the public can create sites from the home page, which a control admin can turn off.
// It is off when the setting cannot be read, so no site is made that should not be.
func (db *ControlDB) PublicSiteCreation(ctx context.Context) bool {
	settings, err := db.Queries.GetServerSettings(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to read the server settings", "error", err)
		return false
	}
	return settings.PublicSiteCreation
}
