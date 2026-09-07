package utils

import (
	"errors"
	"reflect"
	"strings"

	"github.com/bkjonathan/go-authentication/internal/apperror"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// UseJSONFieldNames makes the validator report the name the client actually
// sent - "firstName", not "FirstName". Called once, before any request is
// bound.
func UseJSONFieldNames() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" || name == "" {
			return field.Name
		}
		return name
	})
}

// BindingError turns whatever gin's binder rejected into a 400 the caller can
// act on: a per-field message when the body parsed but failed the rules, and a
// single message when it did not parse at all.
func BindingError(err error) *apperror.Error {
	var invalid validator.ValidationErrors
	if errors.As(err, &invalid) {
		fields := make(map[string]string, len(invalid))
		for _, field := range invalid {
			fields[field.Field()] = validationMessage(field)
		}
		return apperror.BadRequest("validation_failed", "Some fields are invalid.").WithFields(fields)
	}
	return apperror.BadRequest("invalid_request", "The request body could not be read.").Wrap(err)
}

func validationMessage(field validator.FieldError) string {
	switch field.Tag() {
	case "required":
		return "This field is required."
	case "email":
		return "Enter a valid email address."
	case "min":
		return "Must be at least " + field.Param() + " characters."
	case "max":
		return "Must be at most " + field.Param() + " characters."
	case "eqfield":
		return "Does not match " + field.Param() + "."
	case "nefield":
		return "Must differ from " + field.Param() + "."
	case "uuid":
		return "Must be a valid identifier."
	case "oneof":
		return "Must be one of: " + strings.ReplaceAll(field.Param(), " ", ", ") + "."
	default:
		return "This value is invalid."
	}
}
