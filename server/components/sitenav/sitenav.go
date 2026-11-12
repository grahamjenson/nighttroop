// Package sitenav is the one list of the site's pages that the header, the account menu, the hubs' tabs and the search
// all show, so each page is listed once
package sitenav

import (
	"context"
	"slices"
	"strings"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/ui/icon"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/i18n"
)

// Hub is an area of the site's pages, each opened from the header or the account menu and showing its pages as tabs,
// by the id of its name's message
type Hub string

const (
	Data       Hub = "site.nav.hubs.data"
	Reports    Hub = "site.nav.hubs.reports"
	Settings   Hub = "site.nav.hubs.settings"
	Connectors Hub = "site.nav.hubs.connectors"
	Alarms     Hub = "site.nav.hubs.alarms"
	Logs       Hub = "site.nav.hubs.logs"
	Usage      Hub = "site.nav.hubs.usage"
)

// HeaderHubs are the hubs the header links to, in its order
var HeaderHubs = []Hub{Data, Reports}

// MenuHubs are the hubs the account menu opens, in its order
var MenuHubs = []Hub{Settings, Connectors, Alarms, Logs, Usage}

// Slug is the last part of the hub's message id, naming it in IDs, e.g. site--nav--menu--logs
func (h Hub) Slug() string {
	return string(h[strings.LastIndex(string(h), ".")+1:])
}

// Page is one of a hub's pages
type Page struct {
	Hub Hub
	// Group is the message id of where the page is in its hub's side tabs, as Data's and Settings' are, "" in a row of
	// tabs, as Reports' and Logs & Events' are
	Group string
	// Label is the message id of the page's name
	Label string
	Path  string
	ID    string
	Icon  func(...icon.Props) templ.Component
	// Store is what the page shows, which the identity needs permission to search to open it
	Store string
	// Within are the paths whose pages, and those under them, count as this page, its Path when none
	Within []string
}

// Pages are every page of the hubs, in the order their tabs show them
var Pages = []Page{
	{Hub: Data, Group: "site.nav.groups.readings_treatments", Label: "site.nav.pages.glucose_readings", Path: "/data/glucose", ID: "site--data--glucose-button", Icon: icon.Droplet, Store: "glucose"},
	{Hub: Data, Group: "site.nav.groups.readings_treatments", Label: "site.nav.pages.insulin_doses", Path: "/data/insulin", ID: "site--data--insulin-button", Icon: icon.Syringe, Store: "insulin_doses"},
	{Hub: Data, Group: "site.nav.groups.readings_treatments", Label: "site.nav.pages.food_intakes", Path: "/data/food", ID: "site--data--food-button", Icon: icon.Utensils, Store: "food"},
	{Hub: Data, Group: "site.nav.groups.readings_treatments", Label: "site.nav.pages.temp_targets", Path: "/data/temp-targets", ID: "site--data--temp-targets-button", Icon: icon.Target, Store: "temp_targets"},
	{Hub: Data, Group: "site.nav.groups.readings_treatments", Label: "site.nav.pages.temp_basals", Path: "/data/temp-basals", ID: "site--data--temp-basals-button", Icon: icon.Gauge, Store: "temp_basals"},
	{Hub: Data, Group: "site.nav.groups.readings_treatments", Label: "site.nav.pages.profile_switches", Path: "/data/profile-switches", ID: "site--data--profile-switches-button", Icon: icon.Repeat, Store: "profile_switches"},
	{Hub: Data, Group: "site.nav.groups.events", Label: "site.nav.pages.activities", Path: "/data/activities", ID: "site--data--activities-button", Icon: icon.Activity, Store: "activities"},
	{Hub: Data, Group: "site.nav.groups.events", Label: "site.nav.pages.notifications", Path: "/data/notifications", ID: "site--data--notifications-button", Icon: icon.Bell, Store: "notification_events"},
	{Hub: Data, Group: "site.nav.groups.events", Label: "site.nav.pages.event_groups", Path: "/data/event-groups", ID: "site--data--event-groups-button", Icon: icon.Layers, Store: "event_groups"},
	{Hub: Data, Group: "site.nav.groups.devices", Label: "site.nav.pages.device_status", Path: "/data/device-status", ID: "site--data--device-status-button", Icon: icon.Smartphone, Store: "device_statuses"},
	{Hub: Data, Group: "site.nav.groups.presets", Label: "site.nav.pages.food_presets", Path: "/presets/food", ID: "site--presets--food-button", Icon: icon.Utensils, Store: "food_presets"},
	{Hub: Data, Group: "site.nav.groups.presets", Label: "site.nav.pages.profiles", Path: "/presets/profiles", ID: "site--presets--profiles-button", Icon: icon.CalendarClock, Store: "profile_presets"},
	{Hub: Data, Group: "site.nav.groups.presets", Label: "site.nav.pages.loop_settings", Path: "/presets/loop-settings", ID: "site--presets--loop-settings-button", Icon: icon.Settings2, Store: "loop_settings"},

	// Summaries are first, so the Reports hub opens them
	{Hub: Reports, Label: "site.nav.pages.summaries", Path: "/reports", ID: "site--reports--summaries-button", Icon: icon.Table, Store: "summary_periods",
		Within: []string{"/reports", "/reports/summaries"}},
	{Hub: Reports, Label: "site.nav.pages.overview", Path: "/reports/overview", ID: "site--reports--overview-button", Icon: icon.Gauge, Store: "summary_periods"},
	{Hub: Reports, Label: "site.nav.pages.agp", Path: "/reports/agp", ID: "site--reports--agp-button", Icon: icon.ChartLine, Store: "summary_periods"},
	{Hub: Reports, Label: "site.nav.pages.tdd", Path: "/reports/tdd", ID: "site--reports--tdd-menu-button", Icon: icon.ChartColumn, Store: "summary_daily"},

	{Hub: Settings, Group: "site.nav.groups.this_site", Label: "site.nav.pages.general", Path: "/settings", ID: "site--settings--settings-button", Icon: icon.SlidersHorizontal,
		Store: "identity_settings", Within: []string{"/settings/site", "/settings/identity"}},
	{Hub: Settings, Group: "site.nav.groups.people_devices", Label: "site.nav.pages.users_devices", Path: "/settings/identities", ID: "site--settings--identities-button", Icon: icon.Users,
		Store: "identities", Within: []string{"/settings/identities", "/settings/identity-settings"}},
	{Hub: Settings, Group: "site.nav.groups.push", Label: "site.nav.pages.push_devices", Path: "/settings/push-devices", ID: "site--settings--push-devices-button", Icon: icon.Send,
		Store: "push_registrations", Within: []string{"/settings/push-devices", "/settings/push-registrations"}},
	{Hub: Settings, Group: "site.nav.groups.push", Label: "site.nav.pages.push_certificates", Path: "/settings/push-credentials", ID: "site--settings--push-button", Icon: icon.KeyRound, Store: "push_credentials"},

	// Connectors is a hub of its own page, where the site's data comes from, for its admins
	{Hub: Connectors, Label: "site.nav.pages.connectors", Path: "/settings/connectors", ID: "site--settings--connectors-button", Icon: icon.Plug, Store: "connector_settings"},

	// Alarms is a hub of its own page, so it has no tabs
	{Hub: Alarms, Label: "site.nav.pages.alarms", Path: "/settings/alarms", ID: "site--settings--alarms-button", Icon: icon.BellRing, Store: "alarm_settings"},

	{Hub: Logs, Label: "site.nav.pages.events", Path: "/logs/events", ID: "site--logs--events-button", Icon: icon.Activity, Store: "site_events"},
	{Hub: Logs, Label: "site.nav.pages.errors", Path: "/logs/errors", ID: "site--logs--errors-button", Icon: icon.TriangleAlert, Store: "site_errors"},
	{Hub: Logs, Label: "site.nav.pages.logs", Path: "/logs/logs", ID: "site--logs--logs-button", Icon: icon.ScrollText, Store: "site_logs"},

	// Usage is a hub of its own page, the site's DB size and a download of it, for its admins
	{Hub: Usage, Label: "site.nav.pages.usage", Path: "/usage", ID: "site--usage-button", Icon: icon.HardDrive, Store: "site_settings"},
}

// keywords are the headings and words on a page, by its path, that the search also finds it by, showing the one found:
// their message ids, or names and abbreviations that are the same in every language
var keywords = map[string][]string{
	"/data/glucose":              {"CGM", "site.nav.keywords.sensor_glucose", "site.nav.keywords.finger_prick", "site.nav.keywords.meter", "site.nav.keywords.trend"},
	"/data/insulin":              {"site.nav.keywords.bolus", "site.nav.keywords.extended_bolus", "site.nav.keywords.correction"},
	"/data/food":                 {"site.nav.keywords.carbs", "site.nav.keywords.meals", "site.nav.keywords.snacks"},
	"/data/temp-targets":         {"site.nav.keywords.target_range"},
	"/data/temp-basals":          {"site.nav.keywords.basal_rate"},
	"/data/profile-switches":     {"site.nav.keywords.profile_change"},
	"/data/activities":           {"site.nav.keywords.exercise", "site.nav.keywords.steps", "site.nav.keywords.heart_rate", "site.nav.keywords.calories"},
	"/data/event-groups":         {"site.nav.keywords.sensor_change", "site.nav.keywords.site_change", "site.nav.keywords.cannula", "site.nav.keywords.transmitter"},
	"/data/notifications":        {"site.nav.keywords.announcements", "site.nav.keywords.severity"},
	"/data/device-status":        {"site.nav.keywords.battery", "site.nav.keywords.reservoir", "IOB", "COB", "Loop", "site.nav.keywords.uploader"},
	"/reports":                   {"site.nav.keywords.daily_carbs", "site.nav.keywords.daily_dose"},
	"/reports/overview":          {"site.nav.keywords.average_glucose", "GMI", "site.nav.keywords.standard_deviation", "site.nav.keywords.coefficient_of_variation", "site.nav.keywords.sensor_usage", "site.nav.keywords.best_day"},
	"/reports/agp":               {"site.nav.keywords.ambulatory_glucose_profile", "site.nav.keywords.time_in_range"},
	"/reports/tdd":               {"site.nav.keywords.total_daily_dose"},
	"/settings":                  {"site.nav.keywords.units", "mmol/L", "mg/dL", "site.nav.keywords.timezone", "site.nav.keywords.language", "site.nav.keywords.date_format", "site.nav.keywords.time_format", "site.nav.keywords.week_starts_on", "site.nav.keywords.password_totp", "site.nav.keywords.delete_site"},
	"/usage":                     {"site.nav.keywords.db_size", "site.nav.keywords.storage", "site.nav.keywords.truncate", "site.nav.keywords.download_data"},
	"/settings/alarms":           {"site.nav.keywords.low_alarm", "site.nav.keywords.urgent_low_alarm", "site.nav.keywords.predicted_low_alarm", "site.nav.keywords.high_alarm", "site.nav.keywords.urgent_high_alarm", "site.nav.keywords.fast_drop_alarm", "site.nav.keywords.fast_rise_alarm", "site.nav.keywords.no_data_alarm"},
	"/settings/identities":       {"site.nav.keywords.admins", "site.nav.keywords.read_only", "site.nav.keywords.pwd_phone"},
	"/settings/push-devices":     {"site.nav.keywords.phones"},
	"/settings/connectors":       {"Nightscout", "Dexcom", "Libre", "LibreLinkUp", "site.nav.keywords.virtual", "site.nav.keywords.backfilling"},
	"/settings/push-credentials": {"APNs", "Apple", "Firebase", "FCM", "site.nav.keywords.bundle_ids"},
	"/presets/food":              {"site.nav.keywords.carbs", "site.nav.keywords.fat", "site.nav.keywords.protein", "site.nav.keywords.fibre", "site.nav.keywords.portions", "site.nav.keywords.pre_bolus"},
	"/presets/profiles":          {"site.nav.keywords.basal_rates", "site.nav.keywords.carb_ratio", "site.nav.keywords.insulin_sensitivity", "site.nav.keywords.targets"},
	"/presets/loop-settings":     {"site.nav.keywords.loop_configuration", "Trio", "AAPS", "site.nav.keywords.max_bolus", "site.nav.keywords.max_basal", "site.nav.keywords.override_presets"},
}

// synonyms are what people also call a word in a page's name, by its path, which the search also finds the page by: the
// id of a message of words
var synonyms = map[string]string{
	"/data/glucose":        "site.nav.synonyms.glucose",
	"/data/insulin":        "site.nav.synonyms.insulin",
	"/data/food":           "site.nav.synonyms.food",
	"/presets/food":        "site.nav.synonyms.food",
	"/data/activities":     "site.nav.synonyms.activities",
	"/settings/alarms":     "site.nav.synonyms.alarms",
	"/settings/connectors": "site.nav.synonyms.connectors",
	"/settings/identities": "site.nav.synonyms.users",
	"/presets/profiles":    "site.nav.synonyms.profiles",
}

// translated is the message with the id in the context's language, or "" for none
func translated(ctx context.Context, id string) string {
	if id == "" {
		return ""
	}
	return i18n.T(ctx, id)
}

// keywordsIn are the page's keywords in the context's language
func (p Page) keywordsIn(ctx context.Context) []string {
	var words []string
	for _, word := range keywords[p.Path] {
		if strings.HasPrefix(word, "site.nav.") {
			word = i18n.T(ctx, word)
		}
		words = append(words, word)
	}
	return words
}

// SearchText is what the search finds the page by, in lower case and the context's language: its name, group, hub,
// keywords and synonyms
func (p Page) SearchText(ctx context.Context) string {
	words := append([]string{i18n.T(ctx, p.Label), translated(ctx, p.Group), i18n.T(ctx, string(p.Hub))}, p.keywordsIn(ctx)...)
	if synonyms := translated(ctx, synonyms[p.Path]); synonyms != "" {
		words = append(words, synonyms)
	}
	return strings.ToLower(strings.Join(words, " "))
}

// Hints are the page's keywords in the context's language, split by "|", one of which the search shows when it found
// the page by it
func (p Page) Hints(ctx context.Context) string {
	return strings.Join(p.keywordsIn(ctx), "|")
}

// SearchID is the ID of the page's result in the header's search, e.g. site--nav--search--settings--push
func (p Page) SearchID() string {
	return "site--nav--search--" + strings.TrimSuffix(strings.TrimPrefix(p.ID, "site--"), "-button")
}

// Visible are the pages the identity may open, of every hub or of those given, so no tab or result is refused
func Visible(identity *controldb.SiteIdentity, hubs ...Hub) []Page {
	var pages []Page
	for _, page := range Pages {
		if (len(hubs) == 0 || slices.Contains(hubs, page.Hub)) && siteapiutils.HasPermission(page.Store, siteapiutils.ActionSearch, identity) {
			pages = append(pages, page)
		}
	}
	return pages
}

// At is the page path is of, the one whose path matches most of it, or nil for a page in no hub, such as the summary
func At(path string) *Page {
	var at *Page
	longest := 0
	for i, page := range Pages {
		prefixes := page.Within
		if prefixes == nil {
			prefixes = []string{page.Path}
		}
		for _, prefix := range append(prefixes, page.Path) {
			matches := path == prefix || (strings.HasPrefix(path, prefix+"/") && slices.Contains(prefixes, prefix))
			if matches && len(prefix) > longest {
				at, longest = &Pages[i], len(prefix)
			}
		}
	}
	return at
}

// Group is a hub's pages in one of its groups
type Group struct {
	// Name is the message id of the group's or hub's name
	Name  string
	Pages []Page
}

// Groups are the pages by their hub, or by their group in it, in the order they are in
func Groups(pages []Page, byHub bool) []Group {
	var groups []Group
	for _, page := range pages {
		name := page.Group
		if byHub {
			name = string(page.Hub)
		}
		if len(groups) == 0 || groups[len(groups)-1].Name != name {
			groups = append(groups, Group{Name: name})
		}
		groups[len(groups)-1].Pages = append(groups[len(groups)-1].Pages, page)
	}
	return groups
}
