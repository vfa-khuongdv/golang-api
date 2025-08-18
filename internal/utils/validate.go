package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

// InitValidator initializes the validator engine and registers custom validation rules.
// This function is called during the application startup to ensure that
func InitValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("valid_birthday", ValidateBirthday)
		_ = v.RegisterValidation("not_blank", ValidateNotBlank)
	}
}

// Custom validation func to check no spaces at all in the string
func ValidateNotBlank(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	trimmed := strings.TrimSpace(str)
	return trimmed != ""
}

// ValidateBirthday checks if the birthday is in a valid format and not a future date.
func ValidateBirthday(fl validator.FieldLevel) bool {
	birthdayStr := fl.Field().String()
	layout := "2006-01-02" // Format: YYYY-MM-DD

	// Parse the birthday to check the format
	parsedDate, err := time.Parse(layout, birthdayStr)
	if err != nil {
		return false // Invalid date format
	}

	// Check if the birthday is in the future
	if parsedDate.After(time.Now()) {
		return false // Invalid: birthday can't be in the future
	}

	return true // Valid birthday
}

// TranslateValidationErrors converts validation errors from the validator package
// into a structured ValidationError that can be returned in API responses.
func TranslateValidationErrors(err error, obj any) *apperror.ValidationError {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return &apperror.ValidationError{
			Code:    apperror.ErrValidationFailed,
			Message: err.Error(),
			Fields:  []apperror.FieldError{},
		}
	}

	var fieldErrors []apperror.FieldError

	// Reflect type for JSON tag lookup
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	for _, fe := range ve {
		ns := fe.StructNamespace() // e.g. "Settings[0].Value"
		parts := strings.Split(ns, ".")

		jsonParts := []string{}
		currType := objType

		for i, part := range parts {
			fieldName := part
			indexSuffix := ""

			// Handle slice index, e.g. Settings[0]
			if idx := strings.Index(part, "["); idx != -1 {
				fieldName = part[:idx]
				indexSuffix = part[idx:]
			}

			field, found := currType.FieldByName(fieldName)
			if !found {
				// Join the rest with dots and append
				jsonParts = append(jsonParts, strings.Join(parts[i:], "."))
				break
			}

			jsonTag := field.Tag.Get("json")
			jsonName := strings.Split(jsonTag, ",")[0]
			if jsonName == "" || jsonName == "-" {
				jsonName = fieldName
			}

			jsonParts = append(jsonParts, jsonName+indexSuffix)

			currType = field.Type
			// Dereference pointers
			for currType.Kind() == reflect.Ptr {
				currType = currType.Elem()
			}
			// If slice or array, go to element type
			if currType.Kind() == reflect.Slice || currType.Kind() == reflect.Array {
				currType = currType.Elem()
			}
		}

		fieldName := strings.Join(jsonParts, ".")

		param := fe.Param()
		var msg string

		switch fe.Tag() {
		case "required":
			msg = fmt.Sprintf("%s is required", fieldName)
		case "email":
			msg = fmt.Sprintf("%s must be a valid email address", fieldName)
		case "url":
			msg = fmt.Sprintf("%s must be a valid URL", fieldName)
		case "uuid":
			msg = fmt.Sprintf("%s must be a valid UUID", fieldName)
		case "len":
			msg = fmt.Sprintf("%s must be exactly %s characters long", fieldName, param)
		case "min":
			msg = fmt.Sprintf("%s must be at least %s characters long or numeric", fieldName, param)
		case "max":
			msg = fmt.Sprintf("%s must be at most %s characters long or numeric", fieldName, param)
		case "eq":
			msg = fmt.Sprintf("%s must be equal to %s", fieldName, param)
		case "ne":
			msg = fmt.Sprintf("%s must not be equal to %s", fieldName, param)
		case "lt":
			msg = fmt.Sprintf("%s must be less than %s", fieldName, param)
		case "lte":
			msg = fmt.Sprintf("%s must be less than or equal to %s", fieldName, param)
		case "gt":
			msg = fmt.Sprintf("%s must be greater than %s", fieldName, param)
		case "gte":
			msg = fmt.Sprintf("%s must be greater than or equal to %s", fieldName, param)
		case "oneof":
			msg = fmt.Sprintf("%s must be one of [%s]", fieldName, param)
		case "contains":
			msg = fmt.Sprintf("%s must contain '%s'", fieldName, param)
		case "excludes":
			msg = fmt.Sprintf("%s must not contain '%s'", fieldName, param)
		case "startswith":
			msg = fmt.Sprintf("%s must start with '%s'", fieldName, param)
		case "endswith":
			msg = fmt.Sprintf("%s must end with '%s'", fieldName, param)
		case "ip":
			msg = fmt.Sprintf("%s must be a valid IP address", fieldName)
		case "ipv4":
			msg = fmt.Sprintf("%s must be a valid IPv4 address", fieldName)
		case "ipv6":
			msg = fmt.Sprintf("%s must be a valid IPv6 address", fieldName)
		case "datetime":
			msg = fmt.Sprintf("%s must be a valid datetime (format: %s)", fieldName, param)
		case "numeric":
			msg = fmt.Sprintf("%s must be a numeric value", fieldName)
		case "boolean":
			msg = fmt.Sprintf("%s must be a boolean value", fieldName)
		case "alpha":
			msg = fmt.Sprintf("%s must contain only letters", fieldName)
		case "alphanum":
			msg = fmt.Sprintf("%s must contain only letters and numbers", fieldName)
		case "alphanumunicode":
			msg = fmt.Sprintf("%s must contain only unicode letters and numbers", fieldName)
		case "ascii":
			msg = fmt.Sprintf("%s must contain only ASCII characters", fieldName)
		case "printascii":
			msg = fmt.Sprintf("%s must contain only printable ASCII characters", fieldName)
		case "base64":
			msg = fmt.Sprintf("%s must be a valid base64 string", fieldName)
		case "containsany":
			msg = fmt.Sprintf("%s must contain at least one of the characters in '%s'", fieldName, param)
		case "excludesall":
			msg = fmt.Sprintf("%s must not contain any of the characters in '%s'", fieldName, param)
		case "excludesrune":
			msg = fmt.Sprintf("%s must not contain the rune '%s'", fieldName, param)
		case "isdefault":
			msg = fmt.Sprintf("%s must be the default value", fieldName)
		case "unique":
			msg = fmt.Sprintf("%s must contain unique values", fieldName)
		case "valid_birthday":
			msg = fmt.Sprintf("%s must be a valid date (YYYY-MM-DD) and not in the future", fieldName)
		case "not_blank":
			msg = fmt.Sprintf("%s must not be blank", fieldName)
		default:
			msg = fmt.Sprintf("%s is invalid", fieldName)
		}

		fieldErrors = append(fieldErrors, apperror.FieldError{
			Field:   fieldName,
			Message: msg,
		})
	}

	return apperror.NewValidationError("Validation failed", fieldErrors)
}

// The utility function to map JSON errors to FieldError structs.
func ToFieldErrors(json any) []apperror.FieldError {
	var fieldErrors []apperror.FieldError

	if items, ok := json.([]any); ok {
		for _, item := range items {
			if fieldMap, ok := item.(map[string]any); ok {
				field, _ := fieldMap["field"].(string)
				message, _ := fieldMap["message"].(string)

				fieldErrors = append(fieldErrors, apperror.FieldError{
					Field:   field,
					Message: message,
				})
			}
		}
	}

	return fieldErrors
}
