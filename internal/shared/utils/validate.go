package utils

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

const birthdayLayout = "2006-01-02"

// InitValidator registers custom validation rules with the Gin validator engine.
func InitValidator() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	registrations := []struct {
		name string
		fn   validator.Func
	}{
		{"valid_birthday", ValidateBirthday},
		{"not_blank", ValidateNotBlank},
		{"password_complexity", ValidatePasswordComplexity},
	}
	for _, r := range registrations {
		_ = v.RegisterValidation(r.name, r.fn)
	}
}

func ValidateNotBlank(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}
	return strings.TrimSpace(fl.Field().String()) != ""
}

func ValidatePasswordComplexity(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case strings.Contains("!@#$%^&*()_+-=[]{}|;':\",./<>?", string(ch)):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func ValidateBirthday(fl validator.FieldLevel) bool {
	birthdayStr := fl.Field().String()
	parsedDate, err := time.Parse(birthdayLayout, birthdayStr)
	if err != nil {
		return false
	}
	if parsedDate.After(time.Now()) {
		return false
	}
	return true
}

// TranslateValidationErrors converts binding/validation errors into a structured ValidationError.
func TranslateValidationErrors(err error, obj any) *apperror.ValidationError {
	// Empty or whitespace-only request bodies surface as io.EOF from the JSON
	// decoder. Return a clean, dedicated error instead of leaking "EOF" to
	// clients (previously handled by the now-removed EmptyBody middleware).
	if errors.Is(err, io.EOF) {
		return &apperror.ValidationError{
			Code:    apperror.ErrEmptyData,
			Message: "Request body cannot be empty",
			Fields:  []apperror.FieldError{},
		}
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		// Anything that is neither io.EOF nor a validation error is a JSON
		// decode failure (e.g. malformed body). Return a generic message so
		// decoder internals are never echoed to clients.
		return &apperror.ValidationError{
			Code:    apperror.ErrValidationFailed,
			Message: "Invalid request body",
			Fields:  []apperror.FieldError{},
		}
	}

	var fieldErrors []apperror.FieldError
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Pointer {
		objType = objType.Elem()
	}

	for _, fe := range ve {
		fieldName := resolveFieldJSONPath(fe.StructNamespace(), objType)
		msg := validationMessage(fieldName, fe.Tag(), fe.Param())
		fieldErrors = append(fieldErrors, apperror.FieldError{
			Field:   fieldName,
			Message: msg,
		})
	}

	return apperror.NewValidationError("Validation failed", fieldErrors)
}

// resolveFieldJSONPath converts a struct namespace like "User.Settings[0].Value"
// into its JSON-equivalent path like "settings[0].value".
func resolveFieldJSONPath(ns string, objType reflect.Type) string {
	parts := strings.Split(ns, ".")
	jsonParts := []string{}
	currType := objType

	startIndex := 0
	if len(parts) > 0 && parts[0] == objType.Name() {
		startIndex = 1
	}

	for i := startIndex; i < len(parts); i++ {
		part := parts[i]
		fieldName := part
		indexSuffix := ""

		if idx := strings.Index(part, "["); idx != -1 {
			fieldName = part[:idx]
			indexSuffix = part[idx:]
		}

		field, found := currType.FieldByName(fieldName)
		if !found || !field.IsExported() {
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
		for currType.Kind() == reflect.Pointer {
			currType = currType.Elem()
		}
		if currType.Kind() == reflect.Slice || currType.Kind() == reflect.Array {
			currType = currType.Elem()
		}
	}

	return strings.Join(jsonParts, ".")
}

func validationMessage(fieldName, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fieldName)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fieldName)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fieldName)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fieldName, param)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long or numeric", fieldName, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long or numeric", fieldName, param)
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", fieldName, param)
	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", fieldName, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fieldName, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fieldName, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fieldName, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fieldName, param)
	case "eqfield":
		return fmt.Sprintf("%s must be equal to %s", fieldName, param)
	case "nefield":
		return fmt.Sprintf("%s must not be equal to %s", fieldName, param)
	case "gtfield":
		return fmt.Sprintf("%s must be greater than %s", fieldName, param)
	case "gtefield":
		return fmt.Sprintf("%s must be greater than or equal to %s", fieldName, param)
	case "ltfield":
		return fmt.Sprintf("%s must be less than %s", fieldName, param)
	case "ltefield":
		return fmt.Sprintf("%s must be less than or equal to %s", fieldName, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", fieldName, param)
	case "contains":
		return fmt.Sprintf("%s must contain '%s'", fieldName, param)
	case "excludes":
		return fmt.Sprintf("%s must not contain '%s'", fieldName, param)
	case "startswith":
		return fmt.Sprintf("%s must start with '%s'", fieldName, param)
	case "endswith":
		return fmt.Sprintf("%s must end with '%s'", fieldName, param)
	case "ip":
		return fmt.Sprintf("%s must be a valid IP address", fieldName)
	case "ipv4":
		return fmt.Sprintf("%s must be a valid IPv4 address", fieldName)
	case "ipv6":
		return fmt.Sprintf("%s must be a valid IPv6 address", fieldName)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime (format: %s)", fieldName, param)
	case "numeric":
		return fmt.Sprintf("%s must be a numeric value", fieldName)
	case "boolean":
		return fmt.Sprintf("%s must be a boolean value", fieldName)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", fieldName)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", fieldName)
	case "alphanumunicode":
		return fmt.Sprintf("%s must contain only unicode letters and numbers", fieldName)
	case "ascii":
		return fmt.Sprintf("%s must contain only ASCII characters", fieldName)
	case "printascii":
		return fmt.Sprintf("%s must contain only printable ASCII characters", fieldName)
	case "base64":
		return fmt.Sprintf("%s must be a valid base64 string", fieldName)
	case "containsany":
		return fmt.Sprintf("%s must contain at least one of the characters in '%s'", fieldName, param)
	case "excludesall":
		return fmt.Sprintf("%s must not contain any of the characters in '%s'", fieldName, param)
	case "excludesrune":
		return fmt.Sprintf("%s must not contain the rune '%s'", fieldName, param)
	case "isdefault":
		return fmt.Sprintf("%s must be the default value", fieldName)
	case "unique":
		return fmt.Sprintf("%s must contain unique values", fieldName)
	case "valid_birthday":
		return fmt.Sprintf("%s must be a valid date (YYYY-MM-DD) and not in the future", fieldName)
	case "not_blank":
		return fmt.Sprintf("%s must not be blank", fieldName)
	case "password_complexity":
		return fmt.Sprintf("%s must be at least 8 characters and contain uppercase, lowercase, digit, and special character", fieldName)
	default:
		return fmt.Sprintf("%s is invalid", fieldName)
	}
}

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
