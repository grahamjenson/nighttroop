// Package mail sends an email through Resend, when it is set up
package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// sendTimeout is the longest an email waits for Resend
const sendTimeout = 10 * time.Second

// ErrNotSetUp is that no email can be sent, as the server has no Resend API key or no address to send from
var ErrNotSetUp = errors.New("email is not set up")

// Mailer sends emails through Resend with its API key, from one address
type Mailer struct {
	apiKey, from, url string
	client            *http.Client
}

// New is a mailer of the Resend account with apiKey, sending from an address of a domain verified in it, to url
func New(apiKey, from, url string) *Mailer {
	return &Mailer{apiKey: apiKey, from: from, url: url, client: &http.Client{Timeout: sendTimeout}}
}

// message is an email as Resend's API takes it
type message struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

// Send sends a plain text email to one address, and is why Resend did not take it, if it did not. The address and the
// text are not logged.
func (m *Mailer) Send(ctx context.Context, to, subject, text string) error {
	if m.apiKey == "" || m.from == "" {
		return ErrNotSetUp
	}
	body, err := json.Marshal(message{From: m.from, To: []string{to}, Subject: subject, Text: text})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Resend answered %s", res.Status)
	}
	return nil
}

// Sent is an email Resend was given, as Resend's API has it: who it was to, when it was made, and the last that
// happened to it, such as delivered or bounced
type Sent struct {
	ID        string   `json:"id"`
	To        []string `json:"to"`
	Subject   string   `json:"subject"`
	CreatedAt string   `json:"created_at"`
	LastEvent string   `json:"last_event"`
}

// sentID is what an email's ID in Resend looks like, so one from a request is safe in a URL
var sentID = regexp.MustCompile(`^[0-9A-Za-z-]{1,64}$`)

// List is the emails Resend was last given, the newest first, at most limit of them
func (m *Mailer) List(ctx context.Context, limit int) ([]Sent, error) {
	var list struct {
		Data []Sent `json:"data"`
	}
	return list.Data, m.get(ctx, fmt.Sprintf("%s?limit=%d", m.url, limit), &list)
}

// Get is one email Resend was given, by its ID
func (m *Mailer) Get(ctx context.Context, id string) (*Sent, error) {
	if !sentID.MatchString(id) {
		return nil, fmt.Errorf("%q is not the ID of an email", id)
	}
	var sent Sent
	return &sent, m.get(ctx, m.url+"/"+id, &sent)
}

// get reads what Resend answers at url into into
func (m *Mailer) get(ctx context.Context, url string, into any) error {
	if m.apiKey == "" || m.from == "" {
		return ErrNotSetUp
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Resend answered %s", res.Status)
	}
	return json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(into)
}
