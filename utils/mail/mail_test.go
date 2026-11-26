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
