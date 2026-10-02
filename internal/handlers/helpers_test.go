package handlers_test

import "github.com/vfa-khuongdv/golang-cms/pkg/apperror"

// toFieldErrors converts the "fields" array of a decoded error response into
// the field errors it was built from, for comparing with the expected ones.
func toFieldErrors(json any) []apperror.FieldError {
	var fieldErrors []apperror.FieldError
	items, _ := json.([]any)
	for _, item := range items {
		if fieldMap, ok := item.(map[string]any); ok {
			field, _ := fieldMap["field"].(string)
			message, _ := fieldMap["message"].(string)
			fieldErrors = append(fieldErrors, apperror.FieldError{Field: field, Message: message})
		}
	}
	return fieldErrors
}
