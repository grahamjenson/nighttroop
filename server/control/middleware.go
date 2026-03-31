package control

import (
	"fmt"
	"net/http"
	"strings"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

func JWTAuthMiddleware(controlDB *db.ControlDB, config config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
			contentType := httputils.GetAcceptType(r)
			tokenStr := ""
			switch contentType {
			case httputils.AcceptTypeJSON:
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
					http.Error(w, "Unauthorized Header", http.StatusUnauthorized)
					return
				}

				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			default:
				tokenStr, err = httputils.GetControlTokenFromRequest(r)
				if err != nil {
					fmt.Println("Control Router: No token cookie found, redirecting to login")
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
			}

			adminName, role, err := config.VerifyControlJWTToken(tokenStr)

			if err != nil {
				http.Error(w, "Bad Token", http.StatusInternalServerError)
				return
			}

			// Optional: Inject claims into context
			ctx := httputils.SetAdminNameInHTTPContext(r.Context(), adminName)
			ctx = httputils.SetRoleInHTTPContext(ctx, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
