package server

import (
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/mold/v4/modifiers"
	"github.com/go-playground/validator/v10"
)

// conform and validate are package state, unlike handler dependencies, because
// they only cache struct metadata and are safe for concurrent use (like a
// regexp).
var (
	conform  = modifiers.New()
	validate = newValidator()
)

// decodeAndValidateRequest decodes the request body into v, which must be a
// pointer to a struct, applies its `mod` tags (e.g. trim) and then checks its
// `validate` tags. Normalizing first means "  " fails "required". It reports
// whether v is ready to use; on false the response is written.
func decodeAndValidateRequest(w http.ResponseWriter, r *http.Request, logger *slog.Logger, v any) bool {
	if !decodeJSON(w, r, logger, v) {
		return false
	}
	if err := conform.Struct(r.Context(), v); err != nil {
		// A misuse, such as an unknown tag or a non-pointer v.
		writeInternalError(w, r, logger, err)
		return false
	}
	return validateRequest(w, r, logger, v)
}

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	// Report fields by their JSON name, which is what clients send.
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}

type fieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

// validateRequest checks v and, if it's invalid, responds 400 with the failing
// fields. It reports whether v was valid; on false the response is written.
func validateRequest(w http.ResponseWriter, r *http.Request, logger *slog.Logger, v any) bool {
	err := validate.Struct(v)
	if err == nil {
		return true
	}

	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		// Not a validation failure but a misuse, such as passing a non-struct.
		writeInternalError(w, r, logger, err)
		return false
	}
	fields := make([]fieldError, len(validationErrs))
	for i, fieldErr := range validationErrs {
		fields[i] = fieldError{Field: fieldErr.Field(), Rule: fieldErr.Tag()}
	}
	writeJSON(w, r, logger, http.StatusBadRequest, errorResponse{Error: errorBody{
		Code:    "invalid_body",
		Message: "request body failed validation",
		Fields:  fields,
	}})
	return false
}
