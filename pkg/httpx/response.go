package httpx

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Error errorMessage `json:"error"`
}

type errorMessage struct {
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, errorResponse{
		Error: errorMessage{Message: message},
	})
}
