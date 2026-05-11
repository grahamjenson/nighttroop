package nightscout

import "net/http"

// /
// AUTH
func NSRefreshNightscoutJWTToken(w http.ResponseWriter, r *http.Request) {
	// Get the existing JWT token and return a new one with an extended expiration time

}

func NSRevokeNightscoutJWTToken(w http.ResponseWriter, r *http.Request) {
	// Because the JWT is stateless we wont be implementing this
	http.Error(w, "Not Implemented: JWT revocation is not supported", http.StatusNotImplemented)
}

// //
// ENTRIES
// //
func NSGetEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSCreateEntry(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetTreatments(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSCreateTreatment(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetDeviceStatus(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}
func NSCreateDeviceStatus(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetProfiles(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetCurrentProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSCreateProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetStatus(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetProperties(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetNightscoutJWTToken(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}
