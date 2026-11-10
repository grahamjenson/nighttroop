package badge

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

var classAttribute = regexp.MustCompile(`class="([^"]*)"`)

// TestBadge checks each variant's colours, that one without a variant is neutral, and that a Class can hide it
func TestBadge(t *testing.T) {
	const base = "whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-semibold "
	tests := []struct {
		name    string
		props   Props
		classes string
	}{
		{"neutral", Props{}, base + "bg-gray-100 text-gray-700"},
		{"primary", Props{Variant: VariantPrimary}, base + "bg-primary/10 text-primary"},
		{"gold", Props{Variant: VariantGold}, base + "bg-brand-gold text-brand-night"},
		{"warning", Props{Variant: VariantWarning}, base + "bg-amber-100 text-amber-800"},
		{"error", Props{Variant: VariantError}, base + "bg-red-100 text-red-800"},
		{"unknown", Props{Variant: "unknown"}, base + "bg-gray-100 text-gray-700"},
		{"class", Props{ID: "b", Variant: VariantError, Class: "hidden sm:inline"}, base + "bg-red-100 text-red-800 hidden sm:inline"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var html bytes.Buffer
			ctx := templ.WithChildren(context.Background(), templ.Raw("x"))
			require.NoError(t, Badge(tt.props).Render(ctx, &html))
			require.Equal(t, tt.props.ID != "", strings.HasPrefix(html.String(), `<span id="`+tt.props.ID+`"`))
			require.ElementsMatch(t, strings.Fields(tt.classes), strings.Fields(classAttribute.FindStringSubmatch(html.String())[1]))
			require.True(t, strings.HasSuffix(html.String(), ">x</span>"))
		})
	}
}

// TestLive checks the Live badge connects to its stream and shows and hides itself with it, its scripts not removed
// as unsafe
func TestLive(t *testing.T) {
	var html bytes.Buffer
	require.NoError(t, Live("live", "/events").Render(context.Background(), &html))
	require.Contains(t, html.String(), `hx-sse:connect="/events"`)
	require.Contains(t, html.String(), `hx-on::sse:close="this.classList.add(&#39;hidden&#39;)"`)
	require.Contains(t, html.String(), `this.dataset.connected = &#39;true&#39;"`)
	require.NotContains(t, html.String(), "failed sanitization")
	require.Contains(t, strings.Fields(classAttribute.FindStringSubmatch(html.String())[1]), "hidden")
	require.True(t, strings.HasSuffix(html.String(), ">● Live</span>"))
}
