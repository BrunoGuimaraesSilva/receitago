package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write json: %v\n", err)
	}
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"status": "error", "error": err.Error()})
}
