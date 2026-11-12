package security

import (
	"fmt"
	"net/http"
	"time"
)

const CONTROL_TOKEN_COOKIE = "control_token"
const SITE_TOKEN_COOKIE = "user_token"
const PRE_AUTH_TOKEN_COOKIE = "pre_auth_token"

// tokenCookie is a token cookie for the whole host that scripts cannot read, sent only over HTTPS when the server
// has it, and Strict as nothing should link into a signed in page
func (cfg *SecurityConfig) tokenCookie(name string, value string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   cfg.HTTPS,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}
}

func (cfg *SecurityConfig) setCookie(w http.ResponseWriter, name string, value string, expiresAt time.Time) {
	http.SetCookie(w, cfg.tokenCookie(name, value, expiresAt))
}

// clearCookie tells the browser to drop a cookie setCookie set
func (cfg *SecurityConfig) clearCookie(w http.ResponseWriter, name string) {
	cookie := cfg.tokenCookie(name, "", time.Unix(0, 0))
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

// cookieValue is the value of the request's cookie called name, or an error when it has none
func cookieValue(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", fmt.Errorf("no %s cookie found", name)
	}
	return cookie.Value, nil
}

func (cfg *SecurityConfig) SetControlLoginCookie(w http.ResponseWriter, tokenString string, expiresAt time.Time) {
	cfg.setCookie(w, CONTROL_TOKEN_COOKIE, tokenString, expiresAt)
}

func (cfg *SecurityConfig) ClearControlLoginCookie(w http.ResponseWriter) {
	cfg.clearCookie(w, CONTROL_TOKEN_COOKIE)
}

func GetControlTokenFromRequest(r *http.Request) (string, error) {
	return cookieValue(r, CONTROL_TOKEN_COOKIE)
}

//////
// User
//////

func (cfg *SecurityConfig) SetUserPreAuthCookie(w http.ResponseWriter, identity Credentialed, subdomain string, deviceName string) error {
	tempToken, expiresAt, err := cfg.GeneratePreAuthJWTToken(identity, subdomain, deviceName)
	if err != nil {
		return fmt.Errorf("error generating pre-auth token: %w", err)
	}
	cfg.setCookie(w, PRE_AUTH_TOKEN_COOKIE, tempToken, expiresAt)
	return nil
}

func (cfg *SecurityConfig) ClearUserPreAuthCookie(w http.ResponseWriter) {
	cfg.clearCookie(w, PRE_AUTH_TOKEN_COOKIE)
}

func GetUserPreAuthTokenFromRequest(r *http.Request) (string, error) {
	return cookieValue(r, PRE_AUTH_TOKEN_COOKIE)
}

func (cfg *SecurityConfig) SetUserLoginCookie(w http.ResponseWriter, tokenString string, expiresAt time.Time) {
	cfg.setCookie(w, SITE_TOKEN_COOKIE, tokenString, expiresAt)
}

func (cfg *SecurityConfig) ClearUserLoginCookie(w http.ResponseWriter) {
	cfg.clearCookie(w, SITE_TOKEN_COOKIE)
}

func GetUserTokenFromRequest(r *http.Request) (string, error) {
	return cookieValue(r, SITE_TOKEN_COOKIE)
}
