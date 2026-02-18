package utils_test

import (
	"errors"
	"testing"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

type User struct {
	Birthday string `validate:"required,valid_birthday"`
}

type TestStruct struct {
	Field string `validate:"%s=%s"`
}

func TestValidateBirthday(t *testing.T) {
	validate := validator.New()
	_ = validate.RegisterValidation("valid_birthday", utils.ValidateBirthday)

	tests := []struct {
		name     string
		birthday string
		wantErr  bool
	}{
		{
			name:     "Valid birthday",
			birthday: "2000-01-01",
			wantErr:  false,
		},
		{
			name:     "Invalid format",
			birthday: "01-01-2000",
			wantErr:  true,
		},
		{
			name:     "Future date",
			birthday: time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := User{Birthday: tt.birthday}
			err := validate.Struct(u)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestTranslateValidationErrors(t *testing.T) {
	validate := validator.New()

	testCases := []struct {
		name     string
		tag      string
		value    any
		expected []apperror.FieldError
	}{
		// required
		{name: "required", tag: "required", value: struct {
			Field string `validate:"required"`
		}{Field: ""}, expected: []apperror.FieldError{{Field: "Field", Message: "Field is required"}}},

		// email
		{name: "email", tag: "email", value: struct {
			Email string `validate:"email"`
		}{Email: "invalid-email"}, expected: []apperror.FieldError{{Field: "Email", Message: "Email must be a valid email address"}}},

		// url
		{name: "url", tag: "url", value: struct {
			URL string `validate:"url"`
		}{URL: "invalid-url"}, expected: []apperror.FieldError{{Field: "URL", Message: "URL must be a valid URL"}}},

		// uuid
		{name: "uuid", tag: "uuid", value: struct {
			UUID string `validate:"uuid"`
		}{UUID: "invalid-uuid"}, expected: []apperror.FieldError{{Field: "UUID", Message: "UUID must be a valid UUID"}}},

		// len
		{name: "len", tag: "len", value: struct {
			Field string `validate:"len=5"`
		}{Field: "123"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be exactly 5 characters long"}}},

		// min
		{name: "min", tag: "min", value: struct {
			Field string `validate:"min=5"`
		}{Field: "123"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be at least 5 characters long or numeric"}}},

		// max
		{name: "max", tag: "max", value: struct {
			Field string `validate:"max=2"`
		}{Field: "123"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be at most 2 characters long or numeric"}}},

		// eq
		{name: "eq", tag: "eq", value: struct {
			Field string `validate:"eq=admin"`
		}{Field: "user"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be equal to admin"}}},

		// ne
		{name: "ne", tag: "ne", value: struct {
			Field string `validate:"ne=admin"`
		}{Field: "admin"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must not be equal to admin"}}},

		// lt
		{name: "lt", tag: "lt", value: struct {
			Field int `validate:"lt=10"`
		}{Field: 10}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be less than 10"}}},

		// lte
		{name: "lte", tag: "lte", value: struct {
			Field int `validate:"lte=10"`
		}{Field: 11}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be less than or equal to 10"}}},

		// gt
		{name: "gt", tag: "gt", value: struct {
			Field int `validate:"gt=10"`
		}{Field: 10}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be greater than 10"}}},

		// gte
		{name: "gte", tag: "gte", value: struct {
			Field int `validate:"gte=10"`
		}{Field: 9}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be greater than or equal to 10"}}},

		// oneof
		{name: "oneof", tag: "oneof", value: struct {
			Field string `validate:"oneof=admin user"`
		}{Field: "guest"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be one of [admin user]"}}},

		// contains
		{name: "contains", tag: "contains", value: struct {
			Field string `validate:"contains=@"`
		}{Field: "test.com"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must contain '@'"}}},

		// excludes
		{name: "excludes", tag: "excludes", value: struct {
			Field string `validate:"excludes=@"`
		}{Field: "test@com"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must not contain '@'"}}},

		// startswith
		{name: "startswith", tag: "startswith", value: struct {
			Field string `validate:"startswith=abc"`
		}{Field: "xyz"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must start with 'abc'"}}},

		// endswith
		{name: "endswith", tag: "endswith", value: struct {
			Field string `validate:"endswith=xyz"`
		}{Field: "abc"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must end with 'xyz'"}}},

		// ip - invalid to trigger error
		{name: "ip", tag: "ip", value: struct {
			Field string `validate:"ip"`
		}{Field: "invalid-ip"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be a valid IP address"}}},

		// ipv4 - invalid to trigger error
		{name: "ipv4", tag: "ipv4", value: struct {
			Field string `validate:"ipv4"`
		}{Field: "999.999.999.999"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be a valid IPv4 address"}}},

		// ipv6 - invalid to trigger error
		{name: "ipv6", tag: "ipv6", value: struct {
			Field string `validate:"ipv6"`
		}{Field: "invalid-ipv6"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be a valid IPv6 address"}}},

		// datetime
		{name: "datetime", tag: "datetime", value: struct {
			Field string `validate:"datetime=2006-01-02"`
		}{Field: "invalid-date"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be a valid datetime (format: 2006-01-02)"}}},

		// numeric
		{name: "numeric", tag: "numeric", value: struct {
			Field string `validate:"numeric"`
		}{Field: "abc"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be a numeric value"}}},

		// boolean - use string to trigger error because bool type with false is valid
		{name: "boolean", tag: "boolean", value: struct {
			Field string `validate:"boolean"`
		}{Field: "notbool"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be a boolean value"}}},

		// alpha
		{name: "alpha", tag: "alpha", value: struct {
			Field string `validate:"alpha"`
		}{Field: "abc123"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must contain only letters"}}},

		// alphanum
		{name: "alphanum", tag: "alphanum", value: struct {
			Field string `validate:"alphanum"`
		}{Field: "abc!@#"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must contain only letters and numbers"}}},

		// alphanumunicode
		{name: "alphanumunicode", tag: "alphanumunicode", value: struct {
			Field string `validate:"alphanumunicode"`
		}{Field: "abc123!@#"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must contain only unicode letters and numbers"}}},

		// ascii
		{name: "ascii", tag: "ascii", value: struct {
			Field string `validate:"ascii"`
		}{Field: "abcé"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must contain only ASCII characters"}}},

		// printascii
		{name: "printascii", tag: "printascii", value: struct {
			Field string `validate:"printascii"`
		}{Field: "abc\x00"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must contain only printable ASCII characters"}}},

		// base64
		{name: "base64", tag: "base64", value: struct {
			Field string `validate:"base64"`
		}{Field: "invalid-base64"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be a valid base64 string"}}},

		// containsany
		{name: "containsany", tag: "containsany", value: struct {
			Field string `validate:"containsany=abc"`
		}{Field: "xyz"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must contain at least one of the characters in 'abc'"}}},

		// excludesall
		{name: "excludesall", tag: "excludesall", value: struct {
			Field string `validate:"excludesall=abc"`
		}{Field: "abc"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must not contain any of the characters in 'abc'"}}},

		// excludesrune
		{name: "excludesrune", tag: "excludesrune", value: struct {
			Field string `validate:"excludesrune=あ"`
		}{Field: "あtest"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must not contain the rune 'あ'"}}},

		// isdefault
		{name: "isdefault", tag: "isdefault", value: struct {
			Field string `validate:"isdefault"`
		}{Field: "non-default"}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must be the default value"}}},

		// unique
		{name: "unique", tag: "unique", value: struct {
			Field []string `validate:"unique"`
		}{Field: []string{"a", "b", "a"}}, expected: []apperror.FieldError{{Field: "Field", Message: "Field must contain unique values"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Struct(tc.value)
			assert.Error(t, err)

			validationErr := utils.TranslateValidationErrors(err, tc.value)
			assert.NotNil(t, validationErr)
			assert.Equal(t, tc.expected, validationErr.Fields)
		})
	}
}

func TestTranslateValidationErrors_ExtraCases(t *testing.T) {
	validate := validator.New()

	_ = validate.RegisterValidation("valid_birthday", utils.ValidateBirthday)
	_ = validate.RegisterValidation("not_blank", utils.ValidateNotBlank)

	tests := []struct {
		name     string
		input    any
		expected []apperror.FieldError
	}{
		{
			name: "valid_birthday (future date)",
			input: struct {
				Birthday string `validate:"valid_birthday"`
			}{Birthday: "3000-01-01"},
			expected: []apperror.FieldError{
				{
					Field:   "Birthday",
					Message: "Birthday must be a valid date (YYYY-MM-DD) and not in the future",
				},
			},
		},
		{
			name: "not_blank (blank field)",
			input: struct {
				NewPassword string `validate:"not_blank"`
			}{NewPassword: "   "},
			expected: []apperror.FieldError{
				{
					Field:   "NewPassword",
					Message: "NewPassword must not be blank",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Struct(tc.input)
			assert.Error(t, err)

			result := utils.TranslateValidationErrors(err, tc.input)
			assert.Equal(t, result.Fields, tc.expected)
			assert.Equal(t, result.Code, apperror.ErrValidationFailed)
			assert.Equal(t, result.Message, "Validation failed")
		})
	}
}

func TestInitValidator(t *testing.T) {
	// Initialize the validator and register custom validations
	utils.InitValidator()

	// Get the validator engine from gin binding
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		t.Fatal("Failed to get validator engine")
	}

	// Check if the "valid_birthday" validation function is registered
	err := v.Var("3000-01-01", "valid_birthday") // Future date - should fail validation
	if err == nil {
		t.Error("Expected validation error for future birthday, got nil")
	}

	err = v.Var("2000-01-01", "valid_birthday") // Valid date - should pass validation
	if err != nil {
		t.Errorf("Expected no validation error for valid birthday, got: %v", err)
	}
}

func TestTranslateValidationErrors_DefaultCase(t *testing.T) {
	validate := validator.New()

	// Register a custom tag that is NOT handled in TranslateValidationErrors
	_ = validate.RegisterValidation("custom_unhandled", func(fl validator.FieldLevel) bool {
		return false // always fail to trigger the error
	})

	// Struct with the unhandled custom validation tag
	input := struct {
		Field string `validate:"custom_unhandled"`
	}{
		Field: "any value",
	}

	// Validate it
	err := validate.Struct(input)
	assert.Error(t, err)

	// Run through TranslateValidationErrors
	result := utils.TranslateValidationErrors(err, input)

	assert.Equal(t, result.Code, apperror.ErrValidationFailed)
	assert.Equal(t, result.Message, "Validation failed")
	assert.Equal(t, result.Fields, []apperror.FieldError{
		{
			Field:   "Field",
			Message: "Field is invalid",
		},
	})
}

func TestToFieldErrors(t *testing.T) {
	t.Run("MapArrayToFieldErrors", func(t *testing.T) {
		input := []any{
			map[string]any{"field": "Email", "message": "Email is required"},
			map[string]any{"field": "Password", "message": "Password is too short"},
		}

		result := utils.ToFieldErrors(input)

		assert.Equal(t, []apperror.FieldError{
			{Field: "Email", Message: "Email is required"},
			{Field: "Password", Message: "Password is too short"},
		}, result)
	})

	t.Run("UnsupportedInputReturnsEmpty", func(t *testing.T) {
		assert.Empty(t, utils.ToFieldErrors(map[string]any{"field": "Email"}))
		assert.Empty(t, utils.ToFieldErrors("not-an-array"))
	})
}

func TestTranslateValidationErrors_InternalBranches(t *testing.T) {
	t.Run("NonValidationError", func(t *testing.T) {
		result := utils.TranslateValidationErrors(errors.New("plain error"), struct{}{})
		assert.Equal(t, apperror.ErrValidationFailed, result.Code)
		assert.Equal(t, "plain error", result.Message)
		assert.Empty(t, result.Fields)
	})

	t.Run("PointerObjectAndStructNameHandling", func(t *testing.T) {
		type LoginInput struct {
			Email string `json:"email" validate:"required,email"`
		}

		validate := validator.New()
		input := &LoginInput{Email: "invalid"}
		err := validate.Struct(input)
		assert.Error(t, err)

		result := utils.TranslateValidationErrors(err, input)
		assert.Equal(t, "email", result.Fields[0].Field)
	})

	t.Run("MissingFieldFallback", func(t *testing.T) {
		type Source struct {
			Email string `json:"email" validate:"required"`
		}
		type Different struct {
			Name string `json:"name"`
		}

		validate := validator.New()
		err := validate.Struct(Source{})
		assert.Error(t, err)

		result := utils.TranslateValidationErrors(err, Different{})
		assert.Equal(t, "Source.Email", result.Fields[0].Field)
	})

	t.Run("SliceIndexTraversal", func(t *testing.T) {
		type Child struct {
			Email string `json:"email" validate:"required,email"`
		}
		type Parent struct {
			Items []Child `json:"items" validate:"dive"`
		}

		validate := validator.New()
		input := &Parent{
			Items: []Child{
				{Email: "bad-email"},
			},
		}

		err := validate.Struct(input)
		assert.Error(t, err)

		result := utils.TranslateValidationErrors(err, input)
		assert.Equal(t, "items[0].email", result.Fields[0].Field)
	})

	t.Run("PointerFieldTraversal", func(t *testing.T) {
		type Profile struct {
			Email string `json:"email" validate:"required,email"`
		}
		type User struct {
			Profile *Profile `json:"profile" validate:"required"`
		}

		validate := validator.New()
		input := User{
			Profile: &Profile{Email: "bad-email"},
		}

		err := validate.Struct(input)
		assert.Error(t, err)

		result := utils.TranslateValidationErrors(err, input)
		assert.Equal(t, "profile.email", result.Fields[0].Field)
	})
}
