package i18n_test

import (
	"context"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	parser "github.com/a-h/templ/parser/v2"
	"github.com/a-h/templ/parser/v2/visitor"
	"gopkg.in/yaml.v3"
	"nighttroop.t1d.nz/utils/i18n"
)

const root = "../.."

var (
	// A message id, as written in the code: words joined by dots, e.g. site.reports.title
	idLiteral = regexp.MustCompile(`"([a-z][a-z0-9_]*(?:\.[a-z0-9_]+)+)"`)
	// The id each i18n.T and i18n.N call shows
	shownID     = regexp.MustCompile(`i18n\.[TN]\([^"\n]*?"([^"]+)"`)
	placeholder = regexp.MustCompile(`{{\s*\.(\w+)\s*}}`)
	// Text with a word in it
	words = regexp.MustCompile(`[A-Za-z]{2,}`)
	// A Go string literal in a component that reads as English: a capitalised word and any after it, e.g. "Not found"
	sentence = regexp.MustCompile(`"[A-Z][a-z]+(?: [A-Za-z0-9,.'’:;!?()/%-]+)*"`)
	// Go lines whose English is for logs and errors, not people
	logged = regexp.MustCompile(`slog\.|Errorf\(|errors\.New\(`)
	// Attributes people read or hear
	readAttributes = []string{"placeholder", "title", "aria-label", "alt", "hx-confirm"}
	// Text that is the same in every language: names, units and symbols
	untranslated = map[string]bool{
		"Nighttroop": true, "mg/dL": true, "mmol/L": true, "Nightscout": true, "Dexcom": true, "LibreLinkUp": true,
	}
)

// messages are each locale file's messages by id, with each message's text or its plural forms' texts
func messages(t *testing.T) map[string]map[string][]string {
	t.Helper()
	files, err := fs.Glob(i18n.Locales, "locales/*.yaml")
	if err != nil || len(files) == 0 {
		t.Fatalf("no locale files: %v", err)
	}
	byFile := map[string]map[string][]string{}
	for _, file := range files {
		data, err := fs.ReadFile(i18n.Locales, file)
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]any
		if err := yaml.Unmarshal(data, &raw); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		byFile[file] = map[string][]string{}
		flatten(byFile[file], "", raw)
	}
	return byFile
}

var pluralForms = []string{"zero", "one", "two", "few", "many", "other"}

func flatten(into map[string][]string, prefix string, raw map[string]any) {
	for key, value := range raw {
		id := strings.TrimPrefix(prefix+"."+key, ".")
		switch value := value.(type) {
		case string:
			into[id] = []string{value}
		case map[string]any:
			if _, plural := value["other"]; plural {
				for _, form := range pluralForms {
					if text, ok := value[form].(string); ok {
						into[id] = append(into[id], text)
					}
				}
				continue
			}
			flatten(into, id, value)
		}
	}
}

// sources are the hand written Go and templ files' contents by path
func sources(t *testing.T) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if strings.HasPrefix(name, ".") && path != root || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		generated := strings.HasPrefix(name, "z_") || strings.HasSuffix(name, "_temp.go") || strings.HasSuffix(name, ".gen.go") ||
			strings.HasSuffix(name, "_templ.go")
		if generated || !(strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".templ")) || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		files[path] = string(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func english(t *testing.T) map[string][]string {
	all := map[string][]string{}
	for file, messages := range messages(t) {
		if !strings.HasSuffix(file, ".en.yaml") {
			continue
		}
		for id, texts := range messages {
			if _, ok := all[id]; ok {
				t.Errorf("%s: %s is in another English file too", file, id)
			}
			all[id] = texts
		}
	}
	return all
}

func TestEveryMessageShownIsInEnglishAndEveryOneIsShown(t *testing.T) {
	messages, used := english(t), map[string]bool{}
	for path, source := range sources(t) {
		for _, match := range idLiteral.FindAllStringSubmatch(source, -1) {
			used[match[1]] = true
		}
		for _, match := range shownID.FindAllStringSubmatch(source, -1) {
			if _, ok := messages[match[1]]; !ok {
				t.Errorf("%s: %s is not in an English locale file", path, match[1])
			}
		}
	}
	for _, id := range slices.Sorted(maps.Keys(messages)) {
		if !used[id] {
			t.Errorf("%s is in the English locale files but not in the code", id)
		}
	}
}

func TestTranslationsHaveEnglishMessagesAndPlaceholders(t *testing.T) {
	english := english(t)
	placeholders := func(texts []string) []string {
		var names []string
		for _, text := range texts {
			for _, match := range placeholder.FindAllStringSubmatch(text, -1) {
				names = append(names, match[1])
			}
		}
		slices.Sort(names)
		return slices.Compact(names)
	}
	for file, messages := range messages(t) {
		for id, texts := range messages {
			englishTexts, ok := english[id]
			if !ok {
				t.Errorf("%s: %s is not in English", file, id)
			} else if got, want := placeholders(texts), placeholders(englishTexts); !slices.Equal(got, want) {
				t.Errorf("%s: %s has placeholders %v, but in English %v", file, id, got, want)
			}
		}
	}
}

func TestTemplatesHaveNoTextOutsideTheLocales(t *testing.T) {
	report := func(path string, line uint32, text string) {
		if text = strings.TrimSpace(text); words.MatchString(text) && !untranslated[text] {
			t.Errorf("%s:%d: %q is not from i18n.T", path, line+1, text)
		}
	}
	for path, source := range sources(t) {
		if !strings.Contains(path, "server/components/") || strings.Contains(path, "/icon/") {
			continue
		}
		for i, line := range strings.Split(source, "\n") {
			for _, literal := range sentence.FindAllString(line, -1) {
				if !logged.MatchString(line) {
					report(path, uint32(i), strings.Trim(literal, `"`))
				}
			}
		}
		if !strings.HasSuffix(path, ".templ") {
			continue
		}
		file, err := parser.ParseString(source)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		v := visitor.New()
		v.Text = func(n *parser.Text) error {
			report(path, n.Range.From.Line, n.Value)
			return nil
		}
		v.ConstantAttribute = func(n *parser.ConstantAttribute) error {
			if slices.Contains(readAttributes, n.Key.String()) {
				report(path, n.Range.From.Line, n.Value)
			}
			return nil
		}
		if err := file.Visit(v); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMessagesFallBackToEnglish(t *testing.T) {
	ctx := i18n.WithLanguages(context.Background(), "xx-YY", "fr;q=0.9")
	if got := i18n.T(ctx, "common.nighttroop"); got != "Nighttroop" {
		t.Errorf("T = %q, want the English", got)
	}
	if got := i18n.T(ctx, "missing.message"); got != "missing.message" {
		t.Errorf("T of a missing message = %q, want its id", got)
	}
}
