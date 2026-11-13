// Package i18n is the text people read on the site, the control site and the public pages, in their language. Each
// message is in a locales/<area>.<language>.yaml file, by an id grouped by where it is shown, e.g. site.reports.title,
// with English as every message's fallback.
package i18n

import (
	"context"
	"embed"
	"io/fs"
	"log/slog"
	"maps"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

//go:embed locales/*.yaml
var Locales embed.FS

// Data fills a message's {{.Name}} placeholders
type Data = map[string]any

type localizerKey struct{}

var (
	bundle  = load()
	english = goi18n.NewLocalizer(bundle, language.English.String())
)

func load() *goi18n.Bundle {
	b := goi18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	files, _ := fs.Glob(Locales, "locales/*.yaml")
	for _, file := range files {
		if _, err := b.LoadMessageFileFS(Locales, file); err != nil {
			panic(err)
		}
	}
	return b
}

// WithLanguages is the context with its text in the first of the languages, or Accept-Language headers, there are
// messages for, else English
func WithLanguages(ctx context.Context, languages ...string) context.Context {
	return context.WithValue(ctx, localizerKey{}, goi18n.NewLocalizer(bundle, languages...))
}

// T is the message with the id in the context's language, its placeholders filled from data
func T(ctx context.Context, id string, data ...Data) string {
	config := &goi18n.LocalizeConfig{MessageID: id}
	if len(data) > 0 {
		config.TemplateData = data[0]
	}
	return localize(ctx, config)
}

// N is the plural form of the message with the id for count, in the context's language, with count as {{.Count}}
func N(ctx context.Context, id string, count int, data ...Data) string {
	templateData := Data{"Count": count}
	if len(data) > 0 {
		maps.Copy(templateData, data[0])
	}
	return localize(ctx, &goi18n.LocalizeConfig{MessageID: id, PluralCount: count, TemplateData: templateData})
}

func localize(ctx context.Context, config *goi18n.LocalizeConfig) string {
	localizer, ok := ctx.Value(localizerKey{}).(*goi18n.Localizer)
	if !ok {
		localizer = english
	}
	text, err := localizer.Localize(config)
	if err != nil {
		// A message missing from the locales is a bug, which TestEveryMessageIsInEnglish finds before it is served
		slog.ErrorContext(ctx, "message not localized", "id", config.MessageID, "error", err)
		return config.MessageID
	}
	return text
}

// Error is an error written for people, by its message's id and placeholders, so a page can show it in its language.
// As an error it is in English, as the JSON API and the logs have it.
type Error struct {
	ID   string
	Data Data
}

// NewError is the error the message with the id says, its placeholders filled from data
func NewError(id string, data ...Data) *Error {
	e := &Error{ID: id}
	if len(data) > 0 {
		e.Data = data[0]
	}
	return e
}

func (e *Error) Error() string {
	return T(context.Background(), e.ID, e.Data)
}

// In is the error in the context's language
func (e *Error) In(ctx context.Context) string {
	return T(ctx, e.ID, e.Data)
}
