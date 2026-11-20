package site_settings_templ

import (
	"context"
	"strings"

	"nighttroop.t1d.nz/utils/i18n"
)

// adviceFor is what to do about a failed sync, by the words of its error that say why it failed. The first that
// matches is used, so the more exact are first.
var adviceFor = []struct {
	words  []string
	advice func(ctx context.Context) string
}{
	{[]string{"follows no one", "needs you to"}, func(ctx context.Context) string { return i18n.T(ctx, "site.settings.connector_advice.app") }},
	{[]string{"private or local network"}, func(ctx context.Context) string { return i18n.T(ctx, "site.settings.connector_advice.private") }},
	{[]string{"refused the username or password", "invalid email or password", "HTTP 401", "HTTP 403", "status=401", "status=403"},
		func(ctx context.Context) string { return i18n.T(ctx, "site.settings.connector_advice.credentials") }},
	{[]string{"host unavailable", "no such host", "connection refused", "HTTP 404", "status=404"},
		func(ctx context.Context) string { return i18n.T(ctx, "site.settings.connector_advice.address") }},
	{[]string{"timed out", "timeout", "deadline exceeded", "HTTP 5", "status=5", "HTTP 429", "status=429"},
		func(ctx context.Context) string { return i18n.T(ctx, "site.settings.connector_advice.wait") }},
}

// connectorAdvice says in plain words what to do about a connector's sync errors, or "" when they do not say why
func connectorAdvice(ctx context.Context, syncErrors []string) string {
	all := strings.Join(syncErrors, "\n")
	for _, a := range adviceFor {
		for _, words := range a.words {
			if strings.Contains(all, words) {
				return a.advice(ctx)
			}
		}
	}
	return ""
}
