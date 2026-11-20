package logview

import (
	"context"
	"encoding/json"
	"strings"

	"nighttroop.t1d.nz/utils/i18n"
)

// errorAbout is what an error's attributes say about it: its kind, the status the request was refused with, who sent
// it and from which app, what it asked for, and the connector whose sync failed
type errorAbout struct {
	Kind          string `json:"kind"`
	Status        int    `json:"status"`
	Method        string `json:"method"`
	Route         string `json:"route"`
	Identity      string `json:"identity"`
	Client        string `json:"client"`
	ConnectorType string `json:"connector_type"`
}

func aboutError(raw json.RawMessage) errorAbout {
	var about errorAbout
	_ = json.Unmarshal(raw, &about)
	return about
}

// ErrorFrom is where an error came from, to tell errors apart in a list: the identity, what it asked for and its app,
// which is last as a browser's is long, or the connector, and "" when its attributes do not say
func ErrorFrom(ctx context.Context, raw json.RawMessage) string {
	about := aboutError(raw)
	if about.ConnectorType != "" {
		return i18n.T(ctx, "common.log_view.from_connector", i18n.Data{"Type": about.ConnectorType})
	}
	var parts []string
	for _, part := range []string{about.Identity, strings.TrimSpace(about.Method + " " + about.Route), about.Client} {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, " · ")
}

// ErrorMeaning says in plain words what an error means and what to do about it, by its kind and status, and "" when
// neither says
func ErrorMeaning(ctx context.Context, raw json.RawMessage) string {
	about := aboutError(raw)
	switch {
	case about.Kind == "connector":
		return i18n.T(ctx, "common.log_view.meaning.connector")
	case about.Kind == "site_backup":
		return i18n.T(ctx, "common.log_view.meaning.backup")
	case about.Status == 401:
		return i18n.T(ctx, "common.log_view.meaning.unauthorized")
	case about.Status == 403:
		return i18n.T(ctx, "common.log_view.meaning.forbidden")
	case about.Status == 404:
		return i18n.T(ctx, "common.log_view.meaning.not_found")
	case about.Status == 429:
		return i18n.T(ctx, "common.log_view.meaning.too_many")
	case about.Status >= 500:
		return i18n.T(ctx, "common.log_view.meaning.server")
	case about.Status >= 400:
		return i18n.T(ctx, "common.log_view.meaning.refused")
	}
	return ""
}
