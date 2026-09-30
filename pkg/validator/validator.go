package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shikagari/api/internal/domain"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// kenyacounty validates that a location is one of Kenya's 47 counties.
	// Registered here because county names contain spaces, which the
	// built-in `oneof` rule cannot express.
	_ = validate.RegisterValidation("kenyacounty", func(fl validator.FieldLevel) bool {
		return domain.IsKenyanCounty(fl.Field().String())
	})

	// vehicleyear validates a model year between 1980 and next year.
	// A static `lte` tag would silently start rejecting new cars every January.
	_ = validate.RegisterValidation("vehicleyear", func(fl validator.FieldLevel) bool {
		return domain.IsValidVehicleYear(int(fl.Field().Int()))
	})
}

// Validate validates a struct against its binding tags.
// Returns a map of field → error message for easy JSON serialisation.
func Validate(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		field := toSnakeCase(e.Field())
		errors[field] = formatError(e)
	}
	return errors
}

// formatError converts a ValidationError into a human-readable message.
func formatError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", toSnakeCase(e.Field()))
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", e.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", e.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", e.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(e.Param(), " ", ", "))
	case "kenyacounty":
		return "must be a valid Kenyan county"
	case "vehicleyear":
		return "must be a valid model year"
	case "url":
		return "must be a valid URL"
	default:
		return fmt.Sprintf("failed validation: %s", e.Tag())
	}
}

// toSnakeCase converts a PascalCase field name to snake_case.
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, []rune(strings.ToLower(string(r)))...)
	}
	return string(result)
}
