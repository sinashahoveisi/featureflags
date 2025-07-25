package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom validations
	validate.RegisterValidation("flag_name", validateFlagName)
}

// FlagCreateRequest represents the request payload for creating a flag
type FlagCreateRequest struct {
	Name         string  `json:"name" validate:"required,flag_name,min=3,max=100"`
	Dependencies []int64 `json:"dependencies,omitempty" validate:"dive,gt=0"`
}

// FlagToggleRequest represents the request payload for toggling a flag
type FlagToggleRequest struct {
	Enable bool   `json:"enable"`
	Reason string `json:"reason" validate:"required,min=3,max=500"`
}

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, ", ")
}

// ValidateFlagCreateRequest validates a flag creation request
func ValidateFlagCreateRequest(req FlagCreateRequest) error {
	if err := validate.Struct(req); err != nil {
		return formatValidationErrors(err)
	}
	return nil
}

// ValidateFlagToggleRequest validates a flag toggle request
func ValidateFlagToggleRequest(req FlagToggleRequest) error {
	if err := validate.Struct(req); err != nil {
		return formatValidationErrors(err)
	}
	return nil
}

// ValidateFlagID validates a flag ID
func ValidateFlagID(id int64) error {
	if id <= 0 {
		return errors.New("flag ID must be greater than 0")
	}
	return nil
}

// ValidateActor validates an actor name
func ValidateActor(actor string) error {
	if actor == "" {
		return errors.New("actor is required")
	}
	if len(actor) > 100 {
		return errors.New("actor name too long (max 100 characters)")
	}
	return nil
}

// ValidateDependencies validates a list of dependency IDs
func ValidateDependencies(dependencies []int64) error {
	for _, dep := range dependencies {
		if dep <= 0 {
			return errors.New("dependency IDs must be greater than 0")
		}
	}
	return nil
}

// validateFlagName is a custom validation function for flag names
func validateFlagName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	
	// Flag name should only contain alphanumeric characters, underscores, and hyphens
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '_' || char == '-') {
			return false
		}
	}
	
	// Should not start or end with underscore or hyphen
	if strings.HasPrefix(name, "_") || strings.HasPrefix(name, "-") ||
	   strings.HasSuffix(name, "_") || strings.HasSuffix(name, "-") {
		return false
	}
	
	return true
}

// formatValidationErrors formats validator errors into a custom error format
func formatValidationErrors(err error) error {
	var validationErrors []ValidationError
	
	for _, err := range err.(validator.ValidationErrors) {
		var message string
		
		switch err.Tag() {
		case "required":
			message = "This field is required"
		case "flag_name":
			message = "Flag name must contain only alphanumeric characters, underscores, and hyphens, and cannot start or end with underscore or hyphen"
		case "min":
			message = fmt.Sprintf("Must be at least %s characters long", err.Param())
		case "max":
			message = fmt.Sprintf("Must be at most %s characters long", err.Param())
		case "gt":
			message = fmt.Sprintf("Must be greater than %s", err.Param())
		default:
			message = "Invalid value"
		}
		
		validationErrors = append(validationErrors, ValidationError{
			Field:   err.Field(),
			Message: message,
		})
	}
	
	return ValidationErrors{Errors: validationErrors}
} 