package stat

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumbersShowUnitsSmallerAndGrey(t *testing.T) {
	const unit = `<span class="text-sm font-medium text-gray-500">`
	tests := []struct{ text, html string }{
		{"0.7 U", "0.7" + unit + " U</span>"},
		{"84%", "84" + unit + "%</span>"},
		{"2 days 1 hour ago", "2" + unit + " days </span>1" + unit + " hour ago</span>"},
		{"just now", "just now"},
		{"–", "–"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		require.NoError(t, Numbers(tt.text, "text-sm font-medium text-gray-500").Render(context.Background(), &buf))
		require.Equal(t, tt.html, buf.String(), tt.text)
	}
}
