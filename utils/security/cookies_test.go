package security_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// Each token cookie is set for the whole host, unreadable to scripts, Strict, and Secure on HTTPS; a clear drops it
func TestTokenCookies(t *testing.T) {
	expiresAt := time.Unix(2000000000, 0)
	tests := []struct {
		name    string
		cookie  string
		write   func(cfg *security.SecurityConfig, w http.ResponseWriter)
		read    func(r *http.Request) (string, error)
		value   string
		cleared bool
	}{
		{"set control", security.CONTROL_TOKEN_COOKIE, func(cfg *security.SecurityConfig, w http.ResponseWriter) {
			cfg.SetControlLoginCookie(w, "control-jwt", expiresAt)
		}, security.GetControlTokenFromRequest, "control-jwt", false},
		{"clear control", security.CONTROL_TOKEN_COOKIE, func(cfg *security.SecurityConfig, w http.ResponseWriter) {
			cfg.ClearControlLoginCookie(w)
		}, security.GetControlTokenFromRequest, "", true},
		{"set user", security.SITE_TOKEN_COOKIE, func(cfg *security.SecurityConfig, w http.ResponseWriter) {
			cfg.SetUserLoginCookie(w, "user-jwt", expiresAt)
		}, security.GetUserTokenFromRequest, "user-jwt", false},
		{"clear user", security.SITE_TOKEN_COOKIE, func(cfg *security.SecurityConfig, w http.ResponseWriter) {
			cfg.ClearUserLoginCookie(w)
		}, security.GetUserTokenFromRequest, "", true},
		{"set pre-auth", security.PRE_AUTH_TOKEN_COOKIE, func(cfg *security.SecurityConfig, w http.ResponseWriter) {
			if err := cfg.SetUserPreAuthCookie(w, &controldb.SiteIdentity{ID: 1}, "site", "device"); err != nil {
				t.Fatal(err)
			}
		}, security.GetUserPreAuthTokenFromRequest, "", false},
		{"clear pre-auth", security.PRE_AUTH_TOKEN_COOKIE, func(cfg *security.SecurityConfig, w http.ResponseWriter) {
			cfg.ClearUserPreAuthCookie(w)
		}, security.GetUserPreAuthTokenFromRequest, "", true},
	}
	for _, https := range []bool{false, true} {
		cfg := security.NewSecurityConfig("", consts.Test, "localhost", https)
		for _, tt := range tests {
			rec := httptest.NewRecorder()
			tt.write(cfg, rec)
			cookies := rec.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("%s: got %d cookies, want 1", tt.name, len(cookies))
			}
			c := cookies[0]
			if c.Name != tt.cookie || c.Path != "/" || !c.HttpOnly || c.Secure != https || c.SameSite != http.SameSiteStrictMode {
				t.Errorf("%s (https %v): got %q", tt.name, https, rec.Header().Get("Set-Cookie"))
			}
			if tt.cleared && (c.Value != "" || c.MaxAge != -1 || !c.Expires.Equal(time.Unix(0, 0))) {
				t.Errorf("%s (https %v): not cleared, got %q", tt.name, https, rec.Header().Get("Set-Cookie"))
			}
			if !tt.cleared && (c.Value == "" || c.MaxAge != 0 || !c.Expires.After(time.Now())) {
				t.Errorf("%s (https %v): not set, got %q", tt.name, https, rec.Header().Get("Set-Cookie"))
			}
			if tt.value != "" && c.Value != tt.value {
				t.Errorf("%s (https %v): got value %q, want %q", tt.name, https, c.Value, tt.value)
			}

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if _, err := tt.read(r); err == nil {
				t.Errorf("%s: read a cookie the request does not have", tt.name)
			}
			r.AddCookie(&http.Cookie{Name: tt.cookie, Value: "sent"})
			if got, err := tt.read(r); err != nil || got != "sent" {
				t.Errorf("%s: read %q, %v, want sent", tt.name, got, err)
			}
		}
	}
}
