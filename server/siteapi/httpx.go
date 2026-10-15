package siteapi

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

func FormToMap(form url.Values) map[string]any {
	doc := make(map[string]any)

	for key, values := range form {
		if len(values) == 0 {
			continue
		}
		setNestedValue(doc, strings.Split(key, "."), values[0])
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

// coerceFormValues makes a form submission look like the JSON the API takes:
// form values are strings, so bool and number fields of model (by json tag)
// are converted, and an absent checkbox becomes false.
func coerceFormValues(doc map[string]any, model any) {
	t := reflect.TypeOf(model).Elem()
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
}
