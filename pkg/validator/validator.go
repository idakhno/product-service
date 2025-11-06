package validator

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// DecodeAndValidate decodes JSON from request body and validates the structure.
// Returns an error if decoding or validation fails.
func DecodeAndValidate(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return validate.Struct(v)
}

// HandleValidationError handles validation errors and sends JSON response to client.
// If error is ValidationErrors, returns detailed field information.
// Otherwise returns a generic error message.
func HandleValidationError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		errors := make(map[string]string)
		for _, err := range validationErrors {
			errors[err.Field()] = "failed on the '" + err.Tag() + "' tag"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"errors": errors})
		return
	}

	http.Error(w, "invalid request body", http.StatusBadRequest)
}
