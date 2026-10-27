package apiutils

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/httputils"
)

func ReadJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	return decoder.Decode(dst)
}

// WriteStatus sets the Content-Type for the response, then the status: headers set after WriteHeader are dropped
func WriteStatus(w http.ResponseWriter, acceptType httputils.AcceptType, status int) {
	if acceptType == httputils.AcceptTypeHTMX || acceptType == httputils.AcceptTypeHTML {
		w.Header().Set("Content-Type", "text/html")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
}

// WriteJSON writes body as JSON with the status, setting the Content-Type first as headers set after it are dropped
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encodeJSON(w, body)
}

// encodeJSON writes body as JSON, after WriteStatus has set the Content-Type and status
func encodeJSON(w http.ResponseWriter, body any) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(body)
}

// FormToMap turns a form submission into the document a JSON client would
// send for T. Dotted names nest, and because form values are only strings,
// bool and number fields of T, or pointers to them, (matched by json tag) are converted; an absent
// checkbox becomes false, and a number left blank is not sent.
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
		kind := f.Type.Kind()
		if kind == reflect.Pointer {
			kind = f.Type.Elem().Kind()
		}
		switch kind {
		case reflect.Bool:
			doc[name] = present && (str == "on" || str == "true" || str == "1")
		case reflect.Int, reflect.Int64, reflect.Float64:
			if n, err := strconv.ParseFloat(str, 64); present && err == nil {
				doc[name] = n
			} else if present && str == "" {
				// An empty number input was left blank, so it is not sent
				delete(doc, name)
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

func WriteModelJSON(w http.ResponseWriter, status int, model Model) {
	encodeJSON(w, map[string]any{"status": status, "result": model})
}

// ExtractModelFromRequest reads a JSON or form body into T
func ExtractModelFromRequest[T any](contentType httputils.ContentType, r *http.Request) (*T, *HTTPError) {
	var doc map[string]any
	switch contentType {
	case httputils.ContentTypeJSON:
		if err := ReadJSON(r, &doc); err != nil || len(doc) == 0 {
			slog.Info("failed to read JSON body", "error", err)
			return nil, ErrBadBody

		}
	case httputils.ContentTypeForm:

		if err := r.ParseForm(); err != nil || len(r.Form) == 0 {
			slog.Debug("failed to parse form data", "error", err)
			return nil, ErrBadForm
		}
		doc = FormToMap[T](r.Form)
	default:
		return nil, ErrUnknownContent
	}

	model := new(T)
	err := utils.FromMap(doc, model)
	if err != nil {
		slog.Info("failed to convert doc to model", "error", err)
		return nil, ErrBadBody
	}
	return model, nil
}
