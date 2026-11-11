package db

import (
	"context"
	"encoding/json"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
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

// LoadRateLimits uses the rate limits a control admin set, kept in the server settings, in place of the env's
func (db *ControlDB) LoadRateLimits(ctx context.Context) error {
	settings, err := db.Queries.GetServerSettings(ctx)
	if err != nil {
		return err
	}
	rates := map[string]config.Rate{}
	if err := json.Unmarshal([]byte(settings.RateLimitsJson), &rates); err != nil {
		return err
	}
	db.ServerConfig.RateLimiters.SetRates(rates)
	return nil
}

// SetRateLimits uses the rate limits a control admin set, by name, in place of the env's, at once and from each start;
// one left out, not valid or the env's own uses the env's
func (db *ControlDB) SetRateLimits(ctx context.Context, rates map[string]config.Rate, by string) error {
	set, err := json.Marshal(db.ServerConfig.RateLimiters.SetRates(rates))
	if err != nil {
		return err
	}
	return db.Queries.UpdateRateLimits(ctx, controldb.UpdateRateLimitsParams{RateLimitsJson: string(set), UpdatedByIdentityName: &by})
}
