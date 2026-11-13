package i18n

import (
	"context"
	"testing"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// An error people read is in English as an error, for the JSON API and the logs, and in the context's language on a
// page, with its placeholders filled in each
func TestErrorIsInTheContextsLanguage(t *testing.T) {
	if err := bundle.AddMessages(language.Spanish, &goi18n.Message{ID: "errors.settings.device_twice", Other: "{{.Name}} está dos veces"}); err != nil {
		t.Fatal(err)
	}
	err := NewError("errors.settings.device_twice", Data{"Name": "Dexcom G7"})
	if got := err.Error(); got != "Dexcom G7 is listed twice" {
		t.Errorf("Error() = %q, want the English", got)
	}
	if got := err.In(WithLanguages(context.Background(), "es-NZ")); got != "Dexcom G7 está dos veces" {
		t.Errorf("In(es) = %q, want the Spanish", got)
	}
	if got := err.In(WithLanguages(context.Background(), "fr")); got != "Dexcom G7 is listed twice" {
		t.Errorf("In(fr) = %q, want the English it falls back to", got)
	}
}
