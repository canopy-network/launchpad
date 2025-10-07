package validators

import (
	"context"
	"reflect"
	"strings"

	"github.com/enielson/launchpad/internal/models"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	validate := validator.New()

	// Register custom validation tags
	validate.RegisterValidation("uppercase", validateUppercase)
	validate.RegisterValidation("github_url", validateGithubURL)

	// Register custom struct field names for better error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate: validate,
	}
}

// Validate validates a struct
func (v *Validator) Validate(s interface{}) error {
	return v.validate.Struct(s)
}

// ValidateWithContext validates a struct with context
func (v *Validator) ValidateWithContext(ctx context.Context, s interface{}) error {
	return v.validate.StructCtx(ctx, s)
}

// FormatErrors formats validation errors into a slice of ValidationErrorDetail
func (v *Validator) FormatErrors(err error) []models.ValidationErrorDetail {
	var errors []models.ValidationErrorDetail

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, validationError := range validationErrors {
			errors = append(errors, models.ValidationErrorDetail{
				Field:   validationError.Field(),
				Message: v.getErrorMessage(validationError),
			})
		}
	}

	return errors
}

func (v *Validator) getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "This field must be at least " + err.Param() + " characters long"
	case "max":
		return "This field must be no more than " + err.Param() + " characters long"
	case "email":
		return "This field must be a valid email address"
	case "url":
		return "This field must be a valid URL"
	case "uuid":
		return "This field must be a valid UUID"
	case "oneof":
		return "This field must be one of: " + err.Param()
	case "uppercase":
		return "This field must be uppercase"
	case "github_url":
		return "This field must be a valid GitHub URL"
	case "datetime":
		return "This field must be a valid datetime in RFC3339 format"
	default:
		return "This field is invalid"
	}
}

// Custom validation functions
func validateUppercase(fl validator.FieldLevel) bool {
	return fl.Field().String() == strings.ToUpper(fl.Field().String())
}

func validateGithubURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	return strings.HasPrefix(url, "https://github.com/")
}
