package security

import (
	"fmt"
	"net/http"
	"time"
)

const CONTROL_TOKEN_COOKIE = "control_token"
const SITE_TOKEN_COOKIE = "user_token"
const PRE_AUTH_TOKEN_COOKIE = "pre_auth_token"

func (cfg *SecurityConfig) SetControlLoginCookie(w http.ResponseWriter, tokenString string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     CONTROL_TOKEN_COOKIE,
		Value:    tokenString,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   cfg.HTTPS, // Set to true if using HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
	})
}

func (cfg *SecurityConfig) ClearControlLoginCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CONTROL_TOKEN_COOKIE,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   cfg.HTTPS, // Set to true if using HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
	})
}

func GetControlTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(CONTROL_TOKEN_COOKIE)
	if err != nil {

		return "", fmt.Errorf("no token cookie found")
	}
	return cookie.Value, nil
}

//////
// User
//////

func (cfg *SecurityConfig) SetUserPreAuthCookie(w http.ResponseWriter, deviceName string) error {
	tempToken, expiresAt, err := cfg.GeneratePreAuthJWTToken(deviceName)
	if err != nil {
		return fmt.Errorf("error generating pre-auth token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     PRE_AUTH_TOKEN_COOKIE,
		Value:    tempToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   cfg.HTTPS, // Set to true if using HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
	})
	return nil
}

func (cfg *SecurityConfig) ClearUserPreAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     PRE_AUTH_TOKEN_COOKIE,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   cfg.HTTPS, // Set to true if using HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
	})
}

func GetUserPreAuthTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(PRE_AUTH_TOKEN_COOKIE)
	if err != nil {
		return "", fmt.Errorf("no pre-auth token cookie found")
	}
	return cookie.Value, nil
}

func (cfg *SecurityConfig) SetUserLoginCookie(w http.ResponseWriter, tokenString string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     SITE_TOKEN_COOKIE,
		Value:    tokenString,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   cfg.HTTPS, // Set to true if using HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
	})
}

func (cfg *SecurityConfig) ClearUserLoginCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SITE_TOKEN_COOKIE,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   cfg.HTTPS, // Set to true if using HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
	})
}

func GetUserTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SITE_TOKEN_COOKIE)
	if err != nil {
		return "", fmt.Errorf("no user token cookie found")
	}
	return cookie.Value, nil
}
