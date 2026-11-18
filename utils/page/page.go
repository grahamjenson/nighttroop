// Package page tells the server's admin of a problem that needs them now, such as sites whose DBs cannot be restored,
// through PagerDuty when it is set up
package page

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Func pages the admin at a level with a message and what it is about, as slog attributes
type Func func(ctx context.Context, level Level, message string, args ...any)

// Level is how a page reaches the admin. PagerDuty gets it as the alert's severity; its service's urgency rule and
// each person's notification rules decide whether that is an email, an SMS or a call.
type Level string

const (
	// Email is a low urgency page, for a problem that can wait for the working day
	Email Level = "warning"
	// SMS is a high urgency page, for a problem that needs the admin soon
	SMS Level = "error"
	// Call is a high urgency page, for a problem that needs the admin now, such as sites that are down
	Call Level = "critical"
)

// sendTimeout is the longest a page waits for PagerDuty
const sendTimeout = 10 * time.Second

// Pager pages the admin through PagerDuty with its routing key, or only logs a page without one
type Pager struct {
	// routingKey is the integration key of the PagerDuty service to page, and eventsURL where its events go
	routingKey, eventsURL string
	// source names this server in PagerDuty, as its host name
	source string
	client *http.Client
}

// New is a pager of the PagerDuty service with routingKey, sending to eventsURL, as source. With no routing key it
// only logs each page.
func New(routingKey, eventsURL, source string) *Pager {
	return &Pager{routingKey: routingKey, eventsURL: eventsURL, source: source, client: &http.Client{Timeout: sendTimeout}}
}

// event is a PagerDuty Events API v2 trigger. Its dedup key is the message, so a problem paged again while its
// incident is open adds to it, and does not page again.
type event struct {
	RoutingKey  string  `json:"routing_key"`
	EventAction string  `json:"event_action"`
	DedupKey    string  `json:"dedup_key"`
	Payload     payload `json:"payload"`
}

type payload struct {
	Summary       string         `json:"summary"`
	Source        string         `json:"source"`
	Severity      Level          `json:"severity"`
	CustomDetails map[string]any `json:"custom_details,omitempty"`
}

// Levels are the levels by the name a form sends, as an admin picks one to test
var Levels = map[string]Level{"email": Email, "sms": SMS, "call": Call}

// Admin pages the server's admin at level. It is always logged as an error, which the control site's errors and
// Sentry show, and sent to PagerDuty when it is set up; a page PagerDuty does not take is logged too.
func (p *Pager) Admin(ctx context.Context, level Level, message string, args ...any) {
	if err := p.Page(ctx, level, message, args...); err != nil {
		slog.ErrorContext(ctx, "Cannot send a page to PagerDuty", "error", err)
	}
}

// Page pages the admin as Admin does, and is why PagerDuty did not take it, if it did not
func (p *Pager) Page(ctx context.Context, level Level, message string, args ...any) error {
	slog.ErrorContext(ctx, "Page: "+message, append([]any{"level", level}, args...)...)
	if p.routingKey == "" {
		return nil
	}
	return p.send(ctx, event{RoutingKey: p.routingKey, EventAction: "trigger", DedupKey: message,
		Payload: payload{Summary: message, Source: p.source, Severity: level, CustomDetails: details(args)}})
}

func (p *Pager) send(ctx context.Context, e event) error {
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.WithoutCancel(ctx), http.MethodPost, p.eventsURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("PagerDuty answered %s", res.Status)
	}
	return nil
}

// details are a page's slog attributes as PagerDuty's custom details, each value as text
func details(args []any) map[string]any {
	details := map[string]any{}
	for i := 0; i+1 < len(args); i += 2 {
		details[fmt.Sprint(args[i])] = fmt.Sprint(args[i+1])
	}
	return details
}
