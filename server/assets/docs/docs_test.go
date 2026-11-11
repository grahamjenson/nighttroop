package docs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The terms and privacy policy name the operator and link the support email given, escaped, and never T1DNZ or New
// Zealand unless given
func TestWithOperator(t *testing.T) {
	tests := []struct {
		name, operator, email string
		// shown is the operator, and mailto the support email's link, as the HTML has them
		shown, mailto string
		notWant       []string
	}{
		{"another operator", "Example Health", "help@example.org", "Example Health", `href="mailto:help@example.org"`,
			[]string{"T1DNZ", "nighttroop@t1d.nz", "New Zealand"}},
		{"the project's own", "T1DNZ", "nighttroop@t1d.nz", "T1DNZ", `href="mailto:nighttroop@t1d.nz"`, nil},
		{"HTML is escaped", `<b>A & B</b>`, `"x"@example.org`, "&lt;b&gt;A &amp; B&lt;/b&gt;", `href="mailto:&#34;x&#34;@example.org"`,
			[]string{"<b>A"}},
		{"a name holding a placeholder is not replaced again", "SUPPORT_EMAIL Ltd", "help@example.org", "SUPPORT_EMAIL Ltd",
			`href="mailto:help@example.org"`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terms := WithOperator(TermsOfService, tt.operator, tt.email)
			privacy := WithOperator(PrivacyPolicy, tt.operator, tt.email)
			require.Contains(t, terms, "We are "+tt.shown)
			require.Contains(t, privacy, "operated by "+tt.shown)
			for _, document := range []string{terms, privacy} {
				require.Contains(t, document, tt.mailto)
				require.NotContains(t, document, "OPERATOR_NAME")
				for _, notWant := range tt.notWant {
					require.NotContains(t, document, notWant)
				}
			}
		})
	}
}
