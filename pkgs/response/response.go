package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
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

func MapValidationErrors(err error) map[string]string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make(map[string]string, len(ve))
		for _, fe := range ve {
			out[fe.Field()] = msgForTag(fe.Tag(), fe.Param())
		}
		return out
	}
	return nil
}

func msgForTag(tag string, param string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "url":
		return "Must be a valid URL"
	case "min":
		return fmt.Sprintf("Must be at least %s characters", param)
	case "max":
		return fmt.Sprintf("Must not exceed %s characters", param)
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", param)
	case "alphanumhyphen":
		return "Only alphanumeric characters and hyphens are allowed"
	}
	return fmt.Sprintf("Field failed on tag: %s", tag)
}
