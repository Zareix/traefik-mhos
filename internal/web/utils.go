package web

import (
	"encoding/json"
	"net/http"
)

type UntypedJSON map[string]any

func responseJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func responseJSONWithStatus(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	responseJSON(w, data)
}
