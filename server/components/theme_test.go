package components

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// paletteClass is a Tailwind class naming a colour from a palette, e.g. text-gray-500, hover:bg-red-50 or bg-black/50
var paletteClass = regexp.MustCompile(`\b(?:bg|text|border(?:-[xytrbls])?|ring(?:-offset)?|fill|stroke|outline|divide|from|via|to|placeholder|accent|caret|decoration|shadow)-((?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-\d{2,3}|black|white|brand-night|brand-gold)\b`)

// TestThemeHasEveryColourUsed checks every palette colour a page's classes use is in css/theme.css. The theme clears
// Tailwind's own palette, so a class with a colour it does not define is left out of the CSS without any error.
func TestThemeHasEveryColourUsed(t *testing.T) {
	theme, err := os.ReadFile("css/theme.css")
	if err != nil {
		t.Fatal(err)
	}
	// The pages' templates and Go, and the static scripts but for the minified and vendored ones
	scanned := func(path string) bool {
		switch filepath.Ext(path) {
		case ".templ":
			return true
		case ".go":
			return !strings.HasSuffix(path, "_templ.go") && !strings.HasPrefix(filepath.Base(path), "z_")
		case ".js":
			name := filepath.Base(path)
			return !strings.HasSuffix(name, ".min.js") && name != "htmx.js" && name != "echarts.min.js"
		}
		return false
	}
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !scanned(path) {
			return err
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range paletteClass.FindAllStringSubmatch(string(source), -1) {
			if !strings.Contains(string(theme), "--color-"+match[1]+":") {
				t.Errorf("%s uses %s, but css/theme.css has no --color-%s", path, match[0], match[1])
			}
		}
		return nil
	}
	for _, root := range []string{".", "../assets/static"} {
		if err := filepath.WalkDir(root, walk); err != nil {
			t.Fatal(err)
		}
	}
}
