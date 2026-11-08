// Package sitenav is the one list of the site's pages that the header, the hubs' tabs and the search all show, so
// each page is listed once
package sitenav

import (
	"slices"
	"strings"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/components/ui/icon"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
)

// Hub is an area of the site's pages, each opened from the header and showing its pages as tabs
type Hub string

const (
	Data     Hub = "Data"
	Reports  Hub = "Reports"
	Settings Hub = "Settings"
)

// Hubs are in the order the header shows them
var Hubs = []Hub{Data, Reports, Settings}

// Page is one of a hub's pages
type Page struct {
	Hub Hub
	// Group is where the page is in its hub's side tabs, as Data's and Settings' are, "" in Reports' row of tabs
	Group string
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
	{Hub: Data, Group: "Readings & treatments", Label: "Glucose Readings", Path: "/data/glucose", ID: "site--data--glucose-button", Icon: icon.Droplet, Store: "glucose"},
	{Hub: Data, Group: "Readings & treatments", Label: "Insulin Doses", Path: "/data/insulin", ID: "site--data--insulin-button", Icon: icon.Syringe, Store: "insulin_doses"},
	{Hub: Data, Group: "Readings & treatments", Label: "Food Intakes", Path: "/data/food", ID: "site--data--food-button", Icon: icon.Utensils, Store: "food"},
	{Hub: Data, Group: "Readings & treatments", Label: "Temp Targets", Path: "/data/temp-targets", ID: "site--data--temp-targets-button", Icon: icon.Target, Store: "temp_targets"},
	{Hub: Data, Group: "Readings & treatments", Label: "Temp Basals", Path: "/data/temp-basals", ID: "site--data--temp-basals-button", Icon: icon.Gauge, Store: "temp_basals"},
	{Hub: Data, Group: "Readings & treatments", Label: "Profile Switches", Path: "/data/profile-switches", ID: "site--data--profile-switches-button", Icon: icon.Repeat, Store: "profile_switches"},
	{Hub: Data, Group: "Events", Label: "Activities", Path: "/data/activities", ID: "site--data--activities-button", Icon: icon.Activity, Store: "activities"},
	{Hub: Data, Group: "Events", Label: "Site & Sensor Changes", Path: "/data/component-events", ID: "site--data--component-events-button", Icon: icon.Wrench, Store: "component_events"},
	{Hub: Data, Group: "Events", Label: "Notifications", Path: "/data/notifications", ID: "site--data--notifications-button", Icon: icon.Bell, Store: "notification_events"},
	{Hub: Data, Group: "Events", Label: "Event Groups", Path: "/data/event-groups", ID: "site--data--event-groups-button", Icon: icon.Layers, Store: "event_groups"},
	{Hub: Data, Group: "Devices", Label: "Device Status", Path: "/data/device-status", ID: "site--data--device-status-button", Icon: icon.Smartphone, Store: "device_statuses"},
	{Hub: Data, Group: "Presets", Label: "Food Presets", Path: "/presets/food", ID: "site--presets--food-button", Icon: icon.Utensils, Store: "food_presets"},
	{Hub: Data, Group: "Presets", Label: "Profiles", Path: "/presets/profiles", ID: "site--presets--profiles-button", Icon: icon.CalendarClock, Store: "profile_presets"},

	{Hub: Reports, Label: "AGP", Path: "/reports", ID: "site--reports--agp-button", Icon: icon.ChartLine, Store: "summary_periods"},
	{Hub: Reports, Label: "TDD", Path: "/reports/tdd", ID: "site--reports--tdd-menu-button", Icon: icon.ChartColumn, Store: "summary_daily"},
	{Hub: Reports, Label: "Summaries", Path: "/reports/summaries", ID: "site--reports--summaries-button", Icon: icon.Table, Store: "summary_periods"},

	{Hub: Settings, Group: "This site", Label: "General", Path: "/settings", ID: "site--settings--settings-button", Icon: icon.SlidersHorizontal,
		Store: "identity_settings", Within: []string{"/settings/site", "/settings/identity"}},
	{Hub: Settings, Group: "This site", Label: "Alarms", Path: "/settings/alarms", ID: "site--settings--alarms-button", Icon: icon.BellRing, Store: "alarm_settings"},
	{Hub: Settings, Group: "People & devices", Label: "Users & Devices", Path: "/settings/identities", ID: "site--settings--identities-button", Icon: icon.Users,
		Store: "identities", Within: []string{"/settings/identities", "/settings/identity-settings"}},
	{Hub: Settings, Group: "People & devices", Label: "Push Devices", Path: "/settings/push-devices", ID: "site--settings--push-devices-button", Icon: icon.Send,
		Store: "push_registrations", Within: []string{"/settings/push-devices", "/settings/push-registrations"}},
	{Hub: Settings, Group: "Connections", Label: "Connectors", Path: "/settings/connectors", ID: "site--settings--connectors-button", Icon: icon.Plug, Store: "connector_settings"},
	{Hub: Settings, Group: "Connections", Label: "Push Notifications", Path: "/settings/push-credentials", ID: "site--settings--push-button", Icon: icon.KeyRound, Store: "push_credentials"},
	{Hub: Settings, Group: "Troubleshooting", Label: "Events", Path: "/logs/events", ID: "site--logs--events-button", Icon: icon.Activity, Store: "site_events"},
	{Hub: Settings, Group: "Troubleshooting", Label: "Errors", Path: "/logs/errors", ID: "site--logs--errors-button", Icon: icon.TriangleAlert, Store: "site_errors"},
	{Hub: Settings, Group: "Troubleshooting", Label: "Logs", Path: "/logs/logs", ID: "site--logs--logs-button", Icon: icon.ScrollText, Store: "site_logs"},
}

// keywords are the headings and words on a page, by its path, that the search also finds it by, showing the one found
var keywords = map[string][]string{
	"/data/glucose":              {"CGM", "Sensor Glucose", "Finger Prick", "Meter", "Trend"},
	"/data/insulin":              {"Bolus", "Extended Bolus", "Correction"},
	"/data/food":                 {"Carbs", "Meals", "Snacks"},
	"/data/temp-targets":         {"Target Range"},
	"/data/temp-basals":          {"Basal Rate"},
	"/data/profile-switches":     {"Profile Change"},
	"/data/activities":           {"Exercise", "Steps", "Heart Rate", "Calories"},
	"/data/component-events":     {"Cannula", "Pump Site", "Sensor Start", "Transmitter"},
	"/data/notifications":        {"Announcements", "Severity"},
	"/data/device-status":        {"Battery", "Reservoir", "IOB", "COB", "Loop", "Uploader"},
	"/reports":                   {"Ambulatory Glucose Profile", "Time in Range"},
	"/reports/tdd":               {"Total Daily Dose"},
	"/reports/summaries":         {"Daily Carbs", "Daily Dose"},
	"/settings":                  {"Units", "mmol/L", "mg/dL", "Timezone", "Language", "Date Format", "Time Format", "Week Starts On", "Password & TOTP", "Download Data", "Delete Site"},
	"/settings/alarms":           {"Low Alarm", "Urgent Low Alarm", "Predicted Low Alarm", "High Alarm", "Urgent High Alarm", "Fast Drop Alarm", "Fast Rise Alarm", "No Data Alarm"},
	"/settings/identities":       {"Admins", "Read Only", "PWD Phone"},
	"/settings/push-devices":     {"Phones"},
	"/settings/connectors":       {"Nightscout", "Dexcom", "Libre", "LibreLinkUp", "Virtual", "Backfilling"},
	"/settings/push-credentials": {"APNs", "Apple", "Firebase", "FCM", "Bundle IDs"},
	"/presets/food":              {"Carbs", "Fat", "Protein", "Fibre", "Portions", "Pre-bolus"},
	"/presets/profiles":          {"Basal Rates", "Carb Ratio", "Insulin Sensitivity", "Targets"},
}

// synonyms are what people also call a word in a page's name, which the search also finds the page by
var synonyms = map[string][]string{
	"glucose":    {"bgl", "bg", "blood sugar", "sgv", "mbg"},
	"insulin":    {"bolus", "dose", "smb", "injection", "shot"},
	"food":       {"carbs", "carbohydrates", "meal", "snack"},
	"activities": {"exercise", "workout", "sport"},
	"alarms":     {"alerts", "hypo", "hyper"},
	"connectors": {"sync", "integrations", "import", "sources", "cgm"},
	"users":      {"people", "carers", "caregivers", "family"},
	"profiles":   {"basal", "isf", "icr"},
}

// SearchText is what the search finds the page by, in lower case: its name, group, hub, keywords and synonyms
func (p Page) SearchText() string {
	words := append([]string{p.Label, p.Group, string(p.Hub)}, keywords[p.Path]...)
	for _, word := range strings.Fields(strings.ToLower(p.Label)) {
		words = append(words, synonyms[word]...)
	}
	return strings.ToLower(strings.Join(words, " "))
}

// Hints are the page's keywords, split by "|", one of which the search shows when it found the page by it
func (p Page) Hints() string {
	return strings.Join(keywords[p.Path], "|")
}

// SearchID is the ID of the page's result in the header's search, e.g. site--nav--search--settings--push
func (p Page) SearchID() string {
	return "site--nav--search--" + strings.TrimSuffix(strings.TrimPrefix(p.ID, "site--"), "-button")
}

// Visible are the pages the identity may open, of every hub or of one, so no tab or result is refused
func Visible(identity *controldb.SiteIdentity, hubs ...Hub) []Page {
	var pages []Page
	for _, page := range Pages {
		if (len(hubs) == 0 || page.Hub == hubs[0]) && siteapiutils.HasPermission(page.Store, siteapiutils.ActionSearch, identity) {
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
