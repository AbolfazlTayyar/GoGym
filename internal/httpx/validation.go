package httpx

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// init teaches gin's shared validator to report a field by its json tag name
// ("first_name") instead of its Go struct field name ("FirstName"), so the
// keys in ErrorBody.Fields match the keys the client actually sent.
//
// It runs here, in an init, rather than from an exported setup function
// called at startup, because every package that writes an envelope imports
// httpx — which makes the registration impossible to forget in a test that
// builds its own router, and a test asserting on "FirstName" instead of
// "first_name" is exactly the kind of drift this avoids.
func init() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidationFields translates a binding error from c.ShouldBindJSON into the
// envelope's error.fields map. It returns nil for an error that isn't
// per-field — a syntactically invalid JSON body, say — in which case the
// caller still reports a validation_failed, just without field detail.
func ValidationFields(err error) map[string]string {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return nil
	}

	fields := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		fields[fe.Field()] = validationMessage(fe)
	}
	return fields
}

// validationMessage renders one field error as a sentence fragment. Only the
// binding tags the request structs actually use get a tailored message; the
// rest fall back to naming the rule, which is still more useful to a client
// than "invalid request".
func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	default:
		return fmt.Sprintf("failed the %q rule", fe.Tag())
	}
}
