package components

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// statusTarget is where an element sends an error response, e.g. hx-status:4xx="target:#toaster swap:beforeend" in a
// template or "hx-status:5xx": "..." in its attributes
var statusTarget = regexp.MustCompile(`hx-status:([45]xx)"?\s*[=:]\s*"([^"]*)"`)

// TestErrorsGoToTheToaster checks every element sending a 4xx to #toaster also sends a 5xx there, as one without would
// swap a server error's toast into its own target, and that both add to the toasts shown rather than replace them
func TestErrorsGoToTheToaster(t *testing.T) {
	const toaster = "target:#toaster swap:beforeend"
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".templ" {
			return err
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		counts := map[string]int{}
		for _, match := range statusTarget.FindAllStringSubmatch(string(source), -1) {
			if strings.Contains(match[2], "#toaster") && match[2] != toaster {
				t.Errorf("%s sends a %s to %q, not %q", path, match[1], match[2], toaster)
			}
			counts[match[1]+" "+match[2]]++
		}
		if counts["4xx "+toaster] != counts["5xx "+toaster] {
			t.Errorf("%s sends %d 4xx but %d 5xx to the toaster", path, counts["4xx "+toaster], counts["5xx "+toaster])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
