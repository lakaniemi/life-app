package server

import (
	"errors"
	"fmt"
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
// `validate` tags. Normalizing first means "  " fails "required". Client
// mistakes are returned as an *apiError.
func decodeAndValidateRequest(w http.ResponseWriter, r *http.Request, v any) error {
	if err := decodeJSON(w, r, v); err != nil {
		return err
	}
	if err := conform.Struct(r.Context(), v); err != nil {
		// A misuse, such as an unknown tag or a non-pointer v.
		return fmt.Errorf("normalize request: %w", err)
	}
	return validateRequest(v)
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

// validateRequest checks v. If it's invalid, it returns a 400 *apiError
// listing the failing fields.
func validateRequest(v any) error {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		// Not a validation failure but a misuse, such as passing a non-struct.
		return fmt.Errorf("validate request: %w", err)
	}
	apiErr := newAPIError(http.StatusBadRequest, "invalid_body", "request body failed validation")
	apiErr.fields = mapSlice(validationErrs, func(fieldErr validator.FieldError) fieldError {
		return fieldError{Field: fieldErr.Field(), Rule: fieldErr.Tag()}
	})
	return apiErr
}
