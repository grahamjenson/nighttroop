package public

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server/components/public_templ"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

type CreateUserResponse struct {
	UserID     string `json:"userID"`
	APIKey     string `json:"apiKey"`
	TOTPSecret string `json:"totpSecret"`
	BackupKey  string `json:"backupKey"`
	PWDNSToken string `json:"pwdNSToken"`
}

func GETMainPageHandler(controlDB *db.ControlDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := public_templ.CreateUser().Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func POSTCreateUserHandler(controlDB *db.ControlDB, config config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowed, remainingTime := httputils.CreateUserAllowed(config, r)
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			public_templ.CreateUserRateLimited(remainingTime+1).Render(r.Context(), w)
			return
		}

		// Check to see if there is 10k users and return error if there is
		if controlDB.GetNUsersDBs() > config.MaxUsers {
			// TODO make this a template like above
			http.Error(w, "Full: Too many users", http.StatusTooManyRequests)
			return
		}

		userCreateResult, err := controlDB.CreateUser()
		if err != nil {
			slog.Error("Unable to create user", "error", err)
			http.Error(w, "Unable to Create User", http.StatusInternalServerError)
			return
		}

		responseModel := CreateUserResponse{
			UserID:     userCreateResult.SiteDB.Subdomain,
			APIKey:     userCreateResult.AdminPassword,
			BackupKey:  userCreateResult.BackupPassword,
			TOTPSecret: userCreateResult.TotpSecret,
			PWDNSToken: userCreateResult.PwdNSToken,
		}

		slog.Info("Number of Users", "count", controlDB.GetNUsersDBs(), "routines", runtime.NumGoroutine())

		if httputils.GetAcceptType(r).API() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(responseModel)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		err = public_templ.CreateUserInfo(userCreateResult.SiteDB.URL(), userCreateResult.AdminPassword, userCreateResult.TotpSecret, userCreateResult.PwdNSToken, userCreateResult.SiteDB.NSURL(), userCreateResult.BackupPassword).Render(r.Context(), w)
		if err != nil {
			slog.Error("Unable to create user", "error", err)
			http.Error(w, "Unable to Create User", http.StatusInternalServerError)
		}

	}
}
