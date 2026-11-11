package controlapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// testConfig allows one login attempt an IP; a config can only be made once per process
var testConfig *config.ServerConfig

func TestMain(m *testing.M) {
	os.Setenv("SITE_LOGIN_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_LOGIN_PER_MINUTE_BURST", "1")
	cfg, err := config.NewServerConfig()
	if err != nil {
		panic(err)
	}
	testConfig = cfg
	os.Exit(m.Run())
}

// A name longer than any identity's is refused before it reaches the rate limiter, which keys on it
func TestControlLoginRejectsLongNameBeforeRateLimit(t *testing.T) {
	longName := strings.Repeat("a", 33)
	tests := []struct {
		name        string
		ip          string
		contentType string
		body        string
	}{
		{"json", "203.0.113.1", "application/json", `{"name":"` + longName + `","password":"pw"}`},
		{"form", "203.0.113.2", "application/x-www-form-urlencoded", "name=" + longName + "&password=pw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/login", strings.NewReader(tt.body))
			r.RemoteAddr = tt.ip + ":1234"
			r.Header.Set("Content-Type", tt.contentType)
			r = r.WithContext(httputils.WithClientIP(r))

			w := httptest.NewRecorder()
			POSTLogin(nil, testConfig)(w, r)
			require.Equal(t, http.StatusUnauthorized, w.Code)

			// The IP still has its one attempt, so the refused login never reached the limiter
			allowed, _ := httputils.ControlLoginAttemptsWithinRateLimit(testConfig, r, "admin"+tt.name)
			require.True(t, allowed)
		})
	}
}

// A refused login answers the login page with an error toast, and the API with the same message as text
func TestControlLoginErrors(t *testing.T) {
	longName := strings.Repeat("a", 33)
	tests := []struct {
		name    string
		ip      string
		htmx    bool
		body    string
		limited string // the name whose one attempt is used up first, if any
		status  int
		message string
	}{
		{"wrong login, page", "203.0.113.11", true, "name=" + longName + "&password=pw", "", http.StatusUnauthorized, "Wrong name, password or code. Codes change every 30 seconds, so use the newest one."},
		{"wrong login, api", "203.0.113.12", false, `{"name":"` + longName + `","password":"pw"}`, "", http.StatusUnauthorized, "Wrong name, password or code. Codes change every 30 seconds, so use the newest one."},
		{"too many, page", "203.0.113.13", true, "name=limitedpage&password=pw", "limitedpage", http.StatusTooManyRequests, "Too many login attempts. Wait a minute, then try again."},
		{"too many, api", "203.0.113.14", false, `{"name":"limitedapi","password":"pw"}`, "limitedapi", http.StatusTooManyRequests, "Too many login attempts. Wait a minute, then try again."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/login", strings.NewReader(tt.body))
			r.RemoteAddr = tt.ip + ":1234"
			r.Header.Set("Content-Type", "application/json")
			if tt.htmx {
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				r.Header.Set("HX-Request", "true")
			}
			r = r.WithContext(httputils.WithClientIP(r))
			if tt.limited != "" {
				allowed, _ := httputils.ControlLoginAttemptsWithinRateLimit(testConfig, r, tt.limited)
				require.True(t, allowed)
			}

			w := httptest.NewRecorder()
			POSTLogin(nil, testConfig)(w, r)
			require.Equal(t, tt.status, w.Code)
			if tt.htmx {
				require.Contains(t, w.Body.String(), `data-variant="error"`)
				require.Contains(t, w.Body.String(), tt.message)
			} else {
				require.Equal(t, tt.message+"\n", w.Body.String())
			}
		})
	}
}

// A name with no admin is refused as a wrong password is, with a password checked, so neither its answer nor how long
// it takes tells which names exist
func TestControlLoginMissingNameIsLikeWrongPassword(t *testing.T) {
	// A config from testutils cannot be made beside TestMain's, as each reads the flags
	controlDB, err := db.NewControlDB(t.Context(), testConfig)
	require.NoError(t, err)
	defer controlDB.Close()

	tests := []struct {
		name string
		htmx bool
	}{{"apiadmin", false}, {"pageadmin", true}}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Its own admin, as login attempts are limited to one for each name
			_, err := controlDB.CreateControlIdentity(t.Context(), tt.name, consts.ControlRoleAdmin, nil)
			require.NoError(t, err)
			// Each from its own IP, as login attempts are limited to one for each IP
			login := func(ip string, name string) (int, string) {
				r := httptest.NewRequest("POST", "/login", strings.NewReader(`{"name":"`+name+`","password":"wrong","totp":"123123"}`))
				r.Header.Set("Content-Type", "application/json")
				if tt.htmx {
					r = httptest.NewRequest("POST", "/login", strings.NewReader("name="+name+"&password=wrong&totp=123123"))
					r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					r.Header.Set("HX-Request", "true")
				}
				r.RemoteAddr = ip + ":1234"
				before := testConfig.Security.PasswordsVerified()
				w := httptest.NewRecorder()
				POSTLogin(controlDB, testConfig)(w, r.WithContext(httputils.WithClientIP(r)))
				require.Equal(t, int64(1), testConfig.Security.PasswordsVerified()-before, "passwords checked logging in as %s", name)
				// A toast has an ID of its own each time
				return w.Code, regexp.MustCompile(`id="id-[A-Z0-9]+"`).ReplaceAllString(w.Body.String(), "")
			}

			wrongStatus, wrongBody := login(fmt.Sprintf("192.0.2.%d", 10+2*i), tt.name)
			missingStatus, missingBody := login(fmt.Sprintf("192.0.2.%d", 11+2*i), "nosuch"+tt.name)
			require.Equal(t, http.StatusUnauthorized, wrongStatus)
			require.Equal(t, wrongStatus, missingStatus)
			require.Equal(t, wrongBody, missingBody)
		})
	}
}
