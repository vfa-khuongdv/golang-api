package dto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
)

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
		assert.NotNil(t, r)
	})

	t.Run("RefreshTokenResult", func(t *testing.T) {
		r := dto.RefreshTokenResult{UserId: 1, Token: &dto.JwtResult{Token: "t"}}
		assert.Equal(t, uint(1), r.UserId)
		assert.Equal(t, "t", r.Token.Token)
	})

	t.Run("Pagination instantiation", func(t *testing.T) {
		p := dto.Pagination[string]{Page: 1, Limit: 10, TotalItems: 100, TotalPages: 10, Data: []string{"a"}}
		assert.Equal(t, 1, p.Page)
		assert.Equal(t, 1, len(p.Data))
	})

	t.Run("LoginInput", func(t *testing.T) {
		i := dto.LoginInput{Email: "a@b.com", Password: "secret"}
		assert.Equal(t, "a@b.com", i.Email)
		assert.Equal(t, "secret", i.Password)
	})

	t.Run("UpdateProfileInput", func(t *testing.T) {
		n := "John"
		g := int16(1)
		i := dto.UpdateProfileInput{Name: &n, Gender: &g}
		assert.Equal(t, "John", *i.Name)
		assert.Equal(t, int16(1), *i.Gender)
	})
}
