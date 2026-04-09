package user

import "net/http"

func NSGetEntries(w http.ResponseWriter, r *http.Request) {
	// TODO implement this properly with pagination and filtering
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSCreateEntry(w http.ResponseWriter, r *http.Request) {
	// TODO implement this properly with pagination and filtering
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetTreatments(w http.ResponseWriter, r *http.Request) {
	// TODO implement this properly with pagination and filtering
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSCreateTreatment(w http.ResponseWriter, r *http.Request) {
	// TODO implement this properly with pagination and filtering
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetDeviceStatus(w http.ResponseWriter, r *http.Request) {
	// TODO implement this properly with pagination and filtering
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}
func NSCreateDeviceStatus(w http.ResponseWriter, r *http.Request) {
	// TODO implement this properly with pagination and filtering
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetProfiles(w http.ResponseWriter, r *http.Request) {
	// TODO implement this properly with pagination and filtering
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}
func NSCreateProfile(w http.ResponseWriter, r *http.Request) {
	// TODO implement this properly with pagination and filtering
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetProperties(w http.ResponseWriter, r *http.Request) {
	// TODO implement this properly with pagination and filtering
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func NSGetNightscoutJWTToken(w http.ResponseWriter, r *http.Request) {
	// TODO implement this properly with pagination and filtering
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}
