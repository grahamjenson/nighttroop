package alert

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

// TestAlert checks each variant's colours, that one without a variant is a warning, and that a title is shown bold
// above the children
func TestAlert(t *testing.T) {
	const base = "rounded-lg border p-3 text-sm "
	tests := []struct {
		name    string
		props   Props
		classes string
		inside  string
	}{
		{"warning", Props{Variant: VariantWarning}, base + "border-amber-400 bg-amber-100 text-amber-800", ">x</div>"},
		{"error", Props{Variant: VariantError}, base + "border-red-200 bg-red-50 text-red-700", ">x</div>"},
		{"info", Props{Variant: VariantInfo}, base + "border-blue-200 bg-blue-50 text-blue-700", ">x</div>"},
		{"no variant", Props{}, base + "border-amber-400 bg-amber-100 text-amber-800", ">x</div>"},
		{"title", Props{ID: "a", Title: "Careful", Class: "mt-2"}, base + "border-amber-400 bg-amber-100 text-amber-800 mt-2",
			`><p class="font-semibold">Careful</p>x</div>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var html bytes.Buffer
			ctx := templ.WithChildren(context.Background(), templ.Raw("x"))
			require.NoError(t, Alert(tt.props).Render(ctx, &html))
			require.Equal(t, tt.props.ID != "", strings.HasPrefix(html.String(), `<div id="`+tt.props.ID+`"`))
			require.ElementsMatch(t, strings.Fields(tt.classes), strings.Fields(classAttribute.FindStringSubmatch(html.String())[1]))
			require.True(t, strings.HasSuffix(html.String(), tt.inside), html.String())
		})
	}
}
