package nightscout

import (
	"fmt"
	"net/http"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/user"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func NightscoutAuthMiddleware(controlDB *db.ControlDB, cfg config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			deviceName, deviceRole, keepGoing, err := user.TryLoginWithJWTToken(cfg, r)
			if err == nil {
				ctx = httputils.SetDeviceNameInHTTPContext(ctx, deviceName)
				ctx = httputils.SetRoleInHTTPContext(ctx, deviceRole)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// If the JWT token is present but incorrect we need to stop
			if !keepGoing {
				http.Error(w, "Unauthorized: Invalid JWT token", http.StatusUnauthorized)
				return
			}

			// Last Chance
			deviceName, deviceRole, err = tryLoginWithLegacyToken(controlDB, cfg, r)
			if err == nil {
				ctx = httputils.SetDeviceNameInHTTPContext(ctx, deviceName)
				ctx = httputils.SetRoleInHTTPContext(ctx, deviceRole)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			http.Error(w, "Unauthorized: No valid authentication method found", http.StatusUnauthorized)
		})
	}
}

func tryLoginWithLegacyToken(controlDB *db.ControlDB, config config.ServerConfig, r *http.Request) (string, string, error) {

	tokenParam := r.URL.Query().Get("token")
	if tokenParam == "" {
		return "", "", fmt.Errorf("no token provided")
	}

	subdomain := httputils.GetSubdomainFromHTTPContext(r.Context())
	userDB, _, err := controlDB.GetUser(subdomain, false)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	deviceName, role, err := config.Security.VerifyLegacyTokenForDevice(tokenParam, subdomain, userDB.Stmts.DeviceGet)
	if err != nil {
		return "", "", fmt.Errorf("bad token: %w", err)
	}

	return deviceName, role, nil
}
