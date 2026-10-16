package siteapiutils

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func ReadJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	return decoder.Decode(dst)
}

// WriteJSON writes a JSON response with the provided HTTP status code.
func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(body)
}

// FormToMap turns a form submission into the document a JSON client would
// send for T. Dotted names nest, and because form values are only strings,
// bool and number fields of T (matched by json tag) are converted; an absent
// checkbox becomes false.
func FormToMap[T any](form url.Values) map[string]any {
	doc := make(map[string]any)

	for key, values := range form {
		if len(values) == 0 {
			continue
		}
		setNestedValue(doc, strings.Split(key, "."), values[0])
	}

	t := reflect.TypeFor[T]()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			name = f.Name
		}
		str, present := doc[name].(string)
		switch f.Type.Kind() {
		case reflect.Bool:
			doc[name] = present && (str == "on" || str == "true" || str == "1")
		case reflect.Int, reflect.Int64, reflect.Float64:
			if n, err := strconv.ParseFloat(str, 64); present && err == nil {
				doc[name] = n
			}
		}
	}
	return doc
}

func setNestedValue(m map[string]any, parts []string, value any) {
	for _, part := range parts[:len(parts)-1] {
		next, ok := m[part].(map[string]any)
		if !ok {
			next = make(map[string]any)
			m[part] = next
		}
		m = next
	}

	m[parts[len(parts)-1]] = value
}

type Model interface {
}

func WriteModelsJSON(w http.ResponseWriter, status int, models []Model) {
	writeJSON(w, map[string]any{"status": status, "result": models})
}

func WriteModelJSON(w http.ResponseWriter, status int, model Model) {
	writeJSON(w, map[string]any{"status": status, "result": model})
}
