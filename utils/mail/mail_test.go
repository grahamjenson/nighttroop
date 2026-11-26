package mail

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	var sent []message
	status := http.StatusOK
	resend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer re_key", r.Header.Get("Authorization"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var m message
		require.NoError(t, json.NewDecoder(r.Body).Decode(&m))
		sent = append(sent, m)
		w.WriteHeader(status)
	}))
	t.Cleanup(resend.Close)

	mailer := New("re_key", "Nighttroop <hello@nt.example.com>", resend.URL)
	require.NoError(t, mailer.Send(t.Context(), "sam@example.com", "Hello", "A message"))
	require.Equal(t, []message{{From: "Nighttroop <hello@nt.example.com>", To: []string{"sam@example.com"}, Subject: "Hello", Text: "A message"}}, sent)

	// An email Resend does not take is an error, with its answer
	status = http.StatusUnprocessableEntity
	require.ErrorContains(t, mailer.Send(t.Context(), "sam@example.com", "Hello", "A message"), "422")

	// Without a key, or an address to send from, nothing is sent
	for _, notSetUp := range []*Mailer{New("", "hello@nt.example.com", resend.URL), New("re_key", "", resend.URL)} {
		require.ErrorIs(t, notSetUp.Send(t.Context(), "sam@example.com", "Hello", "A message"), ErrNotSetUp)
	}
	require.Len(t, sent, 2)
}

// The emails Resend was given are read with the key, as many as asked for, and one by its ID
func TestListAndGet(t *testing.T) {
	var asked []string
	resend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer re_key", r.Header.Get("Authorization"))
		asked = append(asked, r.Method+" "+r.URL.String())
		switch r.URL.Path {
		case "/emails":
			_, _ = w.Write([]byte(`{"object": "list", "has_more": false, "data": [{"id": "e1", "to": ["sam@example.com"], "subject": "Hello", "created_at": "2026-11-25 22:13:42.674981+00", "last_event": "delivered"}]}`))
		case "/emails/e1":
			_, _ = w.Write([]byte(`{"object": "email", "id": "e1", "to": ["sam@example.com"], "subject": "Hello", "last_event": "delivered"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(resend.Close)
	mailer := New("re_key", "hello@nt.example.com", resend.URL+"/emails")

	list, err := mailer.List(t.Context(), 50)
	require.NoError(t, err)
	require.Equal(t, []Sent{{ID: "e1", To: []string{"sam@example.com"}, Subject: "Hello", CreatedAt: "2026-11-25 22:13:42.674981+00", LastEvent: "delivered"}}, list)
	one, err := mailer.Get(t.Context(), "e1")
	require.NoError(t, err)
	require.Equal(t, []string{"sam@example.com"}, one.To)
	require.Equal(t, []string{"GET /emails?limit=50", "GET /emails/e1"}, asked)

	// An ID that is not one asks Resend nothing, one Resend does not have is an error, and so is no key
	_, err = mailer.Get(t.Context(), "../domains")
	require.Error(t, err)
	_, err = mailer.Get(t.Context(), "e2")
	require.ErrorContains(t, err, "404")
	_, err = New("", "hello@nt.example.com", resend.URL).List(t.Context(), 50)
	require.ErrorIs(t, err, ErrNotSetUp)
	require.Len(t, asked, 3)
}
