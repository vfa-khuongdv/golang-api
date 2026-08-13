package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
)

func newDTOValidator() *validator.Validate {
	v := validator.New()
	v.SetTagName("binding")
	_ = v.RegisterValidation("valid_birthday", utils.ValidateBirthday)
	_ = v.RegisterValidation("not_blank", utils.ValidateNotBlank)
	_ = v.RegisterValidation("password_complexity", utils.ValidatePasswordComplexity)
	return v
}

func TestDTOStructs(t *testing.T) {
	t.Run("JwtResult", func(t *testing.T) {
		r := dto.JwtResult{Token: "abc", ExpiresAt: 123}
		assert.Equal(t, "abc", r.Token)
		assert.Equal(t, int64(123), r.ExpiresAt)
	})

	t.Run("LoginResponse", func(t *testing.T) {
		r := dto.LoginResponse{
			AccessToken:  dto.JwtResult{Token: "a", ExpiresAt: 1},
			RefreshToken: dto.JwtResult{Token: "b", ExpiresAt: 2},
		}
		b, err := json.Marshal(r)
		require.NoError(t, err)

		var m map[string]any
		require.NoError(t, json.Unmarshal(b, &m))
		assert.Equal(t, "a", m["access_token"].(map[string]any)["token"])
		assert.Equal(t, float64(2), m["refresh_token"].(map[string]any)["expires_at"])
	})

	t.Run("RefreshTokenResult", func(t *testing.T) {
		r := dto.RefreshTokenResult{UserId: 1, Token: &dto.JwtResult{Token: "t"}}
		assert.Equal(t, uint(1), r.UserId)
		assert.Equal(t, "t", r.Token.Token)

		b, err := json.Marshal(r)
		require.NoError(t, err)
		assert.Contains(t, string(b), `"token"`)
	})

	t.Run("RefreshTokenResult omitempty", func(t *testing.T) {
		r := dto.RefreshTokenResult{UserId: 1}
		b, err := json.Marshal(r)
		require.NoError(t, err)
		assert.NotContains(t, string(b), `"token"`)
		assert.Contains(t, string(b), `"user_id":1`)
	})

	t.Run("Pagination instantiation", func(t *testing.T) {
		p := dto.Pagination[string]{Page: 1, Limit: 10, TotalItems: 100, TotalPages: 10, Data: []string{"a"}}
		assert.Equal(t, 1, p.Page)
		assert.Equal(t, 1, len(p.Data))

		b, err := json.Marshal(p)
		require.NoError(t, err)
		assert.Contains(t, string(b), `"total_items":100`)
		assert.Contains(t, string(b), `"data":["a"]`)
	})
}

func TestDTOBindingRules(t *testing.T) {
	v := newDTOValidator()

	t.Run("LoginInput valid", func(t *testing.T) {
		err := v.Struct(dto.LoginInput{Email: "a@b.com", Password: "secret"})
		assert.NoError(t, err)
	})

	t.Run("LoginInput invalid email", func(t *testing.T) {
		err := v.Struct(dto.LoginInput{Email: "not-an-email", Password: "secret"})
		assert.Error(t, err)
	})

	t.Run("LoginInput short password", func(t *testing.T) {
		err := v.Struct(dto.LoginInput{Email: "a@b.com", Password: "123"})
		assert.Error(t, err)
	})

	t.Run("ResetPasswordInput requires both fields", func(t *testing.T) {
		err := v.Struct(dto.ResetPasswordInput{})
		assert.Error(t, err)
	})

	t.Run("ResetPasswordInput new_password min length", func(t *testing.T) {
		err := v.Struct(dto.ResetPasswordInput{Token: "t", NewPassword: "short"})
		assert.Error(t, err)
	})

	t.Run("CreateUserInput gender oneof", func(t *testing.T) {
		valid := dto.CreateUserInput{
			Email:    "a@b.com",
			Password: "Secret@1",
			Name:     "John",
			Birthday: strPtr("2000-01-01"),
			Address:  strPtr("HN"),
			Gender:   2,
		}
		assert.NoError(t, v.Struct(valid))

		valid.Gender = 4
		assert.Error(t, v.Struct(valid))
	})

	t.Run("CreateUserInput weak password rejected", func(t *testing.T) {
		in := dto.CreateUserInput{
			Email:    "a@b.com",
			Password: "lowercaseonly",
			Name:     "John",
			Birthday: strPtr("2000-01-01"),
			Address:  strPtr("HN"),
			Gender:   1,
		}
		err := v.Struct(in)
		assert.Error(t, err)
		// The error must be the password_complexity rule, not a generic one.
		assert.Contains(t, err.Error(), "password_complexity")
	})

	t.Run("CreateUserInput invalid birthday", func(t *testing.T) {
		in := dto.CreateUserInput{
			Email:    "a@b.com",
			Password: "Secret@1",
			Name:     "John",
			Birthday: strPtr("2000-13-01"),
			Address:  strPtr("HN"),
			Gender:   1,
		}
		assert.Error(t, v.Struct(in))
	})

	t.Run("UpdateProfileInput omitempty fields pass when empty", func(t *testing.T) {
		err := v.Struct(dto.UpdateProfileInput{})
		assert.NoError(t, err)
	})
}

func strPtr(s string) *string {
	return &s
}
