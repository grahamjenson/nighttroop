package httputils

import (
	"fmt"
	"net/http"
	"time"

	"nighttroop.t1d.nz/utils/config"
)

const CONTROL_TOKEN_COOKIE = "control_token"

func SetControlLoginCookie(config config.ServerConfig, w http.ResponseWriter, tokenString string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     CONTROL_TOKEN_COOKIE,
		Value:    tokenString,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   config.IsHTTPS(), // Set to true if using HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
	})
}

func ClearControlLoginCookie(config config.ServerConfig, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CONTROL_TOKEN_COOKIE,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   config.IsHTTPS(), // Set to true if using HTTPS
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

const USER_TOKEN_COOKIE = "user_token"
const PRE_AUTH_TOKEN_COOKIE = "pre_auth_token"

func SetUserPreAuthCookie(config config.ServerConfig, w http.ResponseWriter, deviceName string) error {
	tempToken, expiresAt, err := config.GeneratePreAuthJWTToken(deviceName)
	if err != nil {
		return fmt.Errorf("error generating pre-auth token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     PRE_AUTH_TOKEN_COOKIE,
		Value:    tempToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   config.IsHTTPS(), // Set to true if using HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
	})
	return nil
}

func ClearUserPreAuthCookie(config config.ServerConfig, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     PRE_AUTH_TOKEN_COOKIE,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   config.IsHTTPS(), // Set to true if using HTTPS
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

func SetUserLoginCookie(config config.ServerConfig, w http.ResponseWriter, tokenString string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     USER_TOKEN_COOKIE,
		Value:    tokenString,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   config.IsHTTPS(), // Set to true if using HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
	})
}

func ClearUserLoginCookie(config config.ServerConfig, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     USER_TOKEN_COOKIE,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   config.IsHTTPS(), // Set to true if using HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Going Strict here because we shouldnt have external links to the control plane
	})
}

func GetUserTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(USER_TOKEN_COOKIE)
	if err != nil {
		return "", fmt.Errorf("no user token cookie found")
	}
	return cookie.Value, nil
}
