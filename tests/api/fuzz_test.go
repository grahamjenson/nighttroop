package api_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/tests/dsl/jsonapi"
	"nighttroop.t1d.nz/utils/testutils"
)

// FuzzRequestFiles parses all request files to build bodies, then we mangle them and send them against all create endpoints to test for panics and 500s
func FuzzRequestFiles(f *testing.F) {
	//// SETUP DEFUALT OBJECTS
	config, controlDB, siteInformation, err := testutils.GetConfigControlDBAndSiteDBForTest()
	if err != nil {
		f.Fatalf("failed to get config and databases: %v", err)
	}
	subdomain := siteInformation.SiteDB.Subdomain
	token := siteInformation.PwdNSToken
	if err != nil {
		f.Fatalf("failed to get config and databases: %v", err)
	}
	bodies := []map[string]any{}
	//// COLLECT BODIES AND PATHS AND SUCH
	err = fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".http" {
			return nil
		}

		requestFileBytes, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return nil
		}
		requestFile, err := jsonapi.ParseRequestFile(string(requestFileBytes), map[string]any{}, path)
		if err != nil {
			return nil
		}

		for _, requests := range requestFile.Requests {
			if requests.Body == nil {
				continue
			}
			var bodyJSON map[string]any

			err := json.Unmarshal([]byte(*requests.Body), &bodyJSON)
			if err != nil {
				continue
			}

			bodies = append(bodies, bodyJSON)

		}
		return nil
	})

	if err != nil {
		f.Fatalf("failed walking embedded files: %v", err)
	}

	// The mangled bodies are meant to fail, so discard the logs they cause until this target ends
	testutils.UseLogHandler(f, slog.DiscardHandler)
	router := server.GetMainHandler(controlDB, config)

	if len(bodies) == 0 {
		f.Fatalf("no request bodies found")
	}
	for i := range bodies {
		f.Add(int64(100), 0, i)
	}
	// Each input sends one body, so an input stays well under the fuzzer's 10s hang limit
	f.Fuzz(func(t *testing.T, seed int64, modifications int, bodyIndex int) {
		rrr := rand.New(rand.NewSource(seed))

		body := bodies[(bodyIndex%len(bodies)+len(bodies))%len(bodies)]
		if modifications != 0 {
			body = mutateBody(body, rrr, modifications)
		}
		mutatedBodyBytes, err := json.Marshal(body)
		if err != nil {
			t.Skipf("failed to marshal mutated body: %v", err)
		}
		newBody := string(mutatedBodyBytes)

		// Entries
		runFuzzTest(t, "CreateEntryV1Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v1/entries?token=%s", subdomain, token), newBody)

		runFuzzTest(t, "CreateEntryV3Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v3/entries?token=%s", subdomain, token), newBody)

		// Treatments
		runFuzzTest(t, "CreateTreatmentV1Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v1/treatments?token=%s", subdomain, token), newBody)

		runFuzzTest(t, "CreateTreatmentV3Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v3/treatments?token=%s", subdomain, token), newBody)

		// Device Status
		runFuzzTest(t, "CreateDeviceStatusV1Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v1/devicestatus?token=%s", subdomain, token), newBody)

		runFuzzTest(t, "CreateDeviceStatusV3Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v3/devicestatus?token=%s", subdomain, token), newBody)

		// profile
		runFuzzTest(t, "CreateProfileV1Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v1/profile?token=%s", subdomain, token), newBody)

		runFuzzTest(t, "CreateProfileV3Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v3/profile?token=%s", subdomain, token), newBody)

		// food
		runFuzzTest(t, "CreateFoodV1Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v1/food?token=%s", subdomain, token), newBody)

		runFuzzTest(t, "CreateFoodV3Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v3/food?token=%s", subdomain, token), newBody)

		// settings
		runFuzzTest(t, "CreateSettingsV3Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v3/settings?token=%s", subdomain, token), newBody)

		// activities
		runFuzzTest(t, "CreateActivitiesV1Mod", router, fmt.Sprintf("http://%s.ns.localhost/api/v1/activities?token=%s", subdomain, token), newBody)
	})
}

func runFuzzTest(t *testing.T, name string, router http.Handler, path, body string) {
	t.Run(name, func(t *testing.T) {

		t.Logf("body=%v", string(body))
		req := httptest.NewRequest("POST", path, strings.NewReader(body))

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code == 404 {
			t.Errorf("not found: %d", rr.Code)
		}

		if rr.Code >= 500 {
			t.Errorf("server error: %d: %s", rr.Code, rr.Body.String())
		}

		if rr.Code == 429 {
			t.Errorf("rate limited: %d", rr.Code)
		}

		t.Logf("response code=%v, body=%v", rr.Code, rr.Body.String())
	})
}

func mutateBody(body map[string]any, r *rand.Rand, modifications int) map[string]any {
	body, _ = mutateObject(cloneObject(body), r, modifications)
	return body
}

func mutateObject(body map[string]any, r *rand.Rand, modifications int) (map[string]any, int) {
	if modifications <= 0 {
		return body, 0
	}

	// Mutate existing fields.
	for key, value := range body {
		if modifications <= 0 {
			break
		}

		// Occasionally delete fields entirely.
		if r.Intn(20) == 0 {
			delete(body, key)
			modifications--
			continue
		}

		// Occasionally replace the entire value with a different JSON type.
		if r.Intn(5) == 0 {
			body[key] = randomJSONValue(r)
			modifications--
			continue
		}

		body[key], modifications = mutateJSONValue(value, r, modifications)
	}

	// Add unexpected fields.
	for modifications > 0 && r.Intn(3) == 0 {
		body[randomKey(r)] = randomJSONValue(r)
		modifications--
	}

	// Occasionally rename an existing key.
	if modifications > 0 && len(body) > 0 && r.Intn(10) == 0 {
		keys := mapKeys(body)
		oldKey := keys[r.Intn(len(keys))]
		newKey := randomKey(r)

		if newKey != oldKey {
			body[newKey] = body[oldKey]
			delete(body, oldKey)
			modifications--
		}
	}

	return body, modifications
}

func mutateJSONValue(value any, r *rand.Rand, modifications int) (any, int) {
	if modifications <= 0 {
		return value, 0
	}

	// Sometimes completely replace the value with another JSON type.
	if r.Intn(5) == 0 {
		return randomJSONValue(r), modifications - 1
	}

	switch v := value.(type) {
	case nil:
		return randomJSONValue(r), modifications - 1

	case bool:
		if r.Intn(10) == 0 {
			return randomJSONValue(r), modifications - 1
		}
		return !v, modifications - 1

	case float64:
		return mutateNumber(v, r), modifications - 1

	case string:
		return mutateString(v, r), modifications - 1

	case []any:
		return mutateArray(v, r, modifications)

	case map[string]any:
		return mutateObject(v, r, modifications)

	default:
		return randomJSONValue(r), modifications - 1
	}
}

func mutateArray(array []any, r *rand.Rand, modifications int) ([]any, int) {
	if modifications <= 0 {
		return array, 0
	}

	// Delete random elements.
	for modifications > 0 && len(array) > 0 && r.Intn(10) == 0 {
		i := r.Intn(len(array))
		array = append(array[:i], array[i+1:]...)
		modifications--
	}

	// Mutate existing elements.
	for i := range array {
		if modifications <= 0 {
			break
		}

		if r.Intn(5) == 0 {
			array[i] = randomJSONValue(r)
			modifications--
		} else {
			array[i], modifications = mutateJSONValue(
				array[i],
				r,
				modifications,
			)
		}
	}

	// Add random elements.
	for modifications > 0 && r.Intn(4) == 0 {
		array = append(array, randomJSONValue(r))
		modifications--
	}

	// Occasionally insert a null.
	if modifications > 0 && r.Intn(5) == 0 {
		i := r.Intn(len(array) + 1)
		array = append(array, nil)
		copy(array[i+1:], array[i:])
		array[i] = nil
		modifications--
	}

	return array, modifications
}

func randomJSONValue(r *rand.Rand) any {
	switch r.Intn(6) {
	case 0:
		return r.Intn(2) == 0

	case 1:
		return randomNumber(r)

	case 2:
		return randomString(r)

	case 3:
		// Array
		n := r.Intn(5)
		array := make([]any, 0, n)
		for i := 0; i < n; i++ {
			array = append(array, randomJSONValue(r))
		}
		return array

	case 4:
		// Object
		object := make(map[string]any)
		n := r.Intn(5)
		for i := 0; i < n; i++ {
			object[randomKey(r)] = randomJSONValue(r)
		}
		return object

	default:
		return nil
	}
}

func mutateNumber(v float64, r *rand.Rand) float64 {
	switch r.Intn(10) {
	case 0:
		return 0
	case 1:
		return -v
	case 2:
		return v + 1
	case 3:
		return v - 1
	case 4:
		return v * 2
	case 5:
		return v / 2
	case 6:
		return float64(r.Intn(10))
	case 7:
		return float64(r.Int63())
	case 8:
		return -float64(r.Int63())
	default:
		return r.Float64()*2e10 - 1e10
	}
}

func mutateString(s string, r *rand.Rand) string {
	switch r.Intn(12) {
	case 0:
		return ""
	case 1:
		return randomString(r)
	case 2:
		return s + s
	case 3:
		return s + randomString(r)
	case 4:
		if len(s) > 0 {
			return s[:len(s)/2]
		}
		return s
	case 5:
		return "null"
	case 6:
		return "true"
	case 7:
		return "123"
	case 8:
		return "{}"
	case 9:
		return "[]"
	case 10:
		return "\x00\x01\x02"
	default:
		// Unicode and unusual characters.
		return s + " \t\n\r \u0000 😀 💩"
	}
}

func randomString(r *rand.Rand) string {
	values := []string{
		"",
		"a",
		"null",
		"true",
		"false",
		"0",
		"{}",
		"[]",
		" ",
		"\t\n\r",
		"\x00",
		"😀",
		"💩",
		"日本語",
		"é",
		strings.Repeat("x", 100),
	}

	if r.Intn(3) == 0 {
		return values[r.Intn(len(values))]
	}

	n := r.Intn(50)
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(byte(r.Intn(256)))
	}

	return b.String()
}

func randomKey(r *rand.Rand) string {
	switch r.Intn(8) {
	case 0:
		return ""
	case 1:
		return "unknown"
	case 2:
		return "unexpected"
	case 3:
		return "null"
	case 4:
		return strings.Repeat("x", r.Intn(100))
	case 5:
		return "😀"
	default:
		return randomString(r)
	}
}

func randomNumber(r *rand.Rand) float64 {
	switch r.Intn(10) {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		return -1
	case 3:
		return 0.000000001
	case 4:
		return 9999999999999999
	case 5:
		return -9999999999999999
	default:
		return r.Float64()*2e12 - 1e12
	}
}

func cloneObject(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))

	for key, value := range src {
		dst[key] = cloneJSONValue(value)
	}

	return dst
}

func cloneJSONValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneObject(v)

	case []any:
		dst := make([]any, len(v))
		for i, value := range v {
			dst[i] = cloneJSONValue(value)
		}
		return dst

	default:
		return v
	}
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
