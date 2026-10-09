package server

import (
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate checks request structs against their `validate` struct tags. It's
// shared package state, unlike handler dependencies, because it's a
// stateless tool (like a compiled regexp): it only caches struct metadata and
// is safe for concurrent use, so one instance is the intended usage.
var validate = newValidator()

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

// fieldError says which request field failed which rule, e.g. "name" and
// "required".
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

	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		// Not a validation failure but a misuse, such as passing a non-struct.
		writeInternalError(w, r, logger, err)
		return false
	}
	fields := make([]fieldError, len(verrs))
	for i, fe := range verrs {
		fields[i] = fieldError{Field: fe.Field(), Rule: fe.Tag()}
	}
	writeJSON(w, r, logger, http.StatusBadRequest, errorResponse{Error: errorBody{
		Code:    "invalid_body",
		Message: "request body failed validation",
		Fields:  fields,
	}})
	return false
}
