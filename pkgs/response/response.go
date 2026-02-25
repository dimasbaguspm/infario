package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// ErrorResponse defines the standard structure for all API errors.
// @Description Standard API error response
// @name ErrorResponse
type ErrorResponse struct {
	Status  int               `json:"status"`  // HTTP Status Code
	Message string            `json:"message"` // Human readable summary
	Errors  map[string]string `json:"errors"`  // Detailed field errors
}

func Error(w http.ResponseWriter, code int, message string, details ...map[string]string) {
	resp := ErrorResponse{
		Status:  code,
		Message: message,
	}

	if len(details) > 0 {
		resp.Errors = details[0]
	}

	JSON(w, code, resp)
}
