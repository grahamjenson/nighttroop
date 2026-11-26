// Package mail sends an email through Resend, when it is set up
package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
