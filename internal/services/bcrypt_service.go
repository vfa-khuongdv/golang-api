package services

import (
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

type BcryptService interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hashPassword string) bool
	HashPasswordWithCost(password string, cost int) (string, error)
}

type bcryptServiceImpl struct{}

func NewBcryptService() BcryptService {
	return &bcryptServiceImpl{}
}

// HashPassword hashes a password using bcrypt with the default cost
// Returns the hashed password as a string, or an error if hashing fails
func (s *bcryptServiceImpl) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperror.NewInternalError(err.Error())
	}
	return string(hashedPassword), nil
}

// CheckPasswordHash compares a plain text password with a hashed password
// Returns true if they match, false otherwise
func (s *bcryptServiceImpl) CheckPasswordHash(password, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

// HashPasswordWithCost hashes a password using bcrypt with a specified cost
// Returns the hashed password as a string, or an error if hashing fails
func (s *bcryptServiceImpl) HashPasswordWithCost(password string, cost int) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", apperror.NewInternalError(err.Error())
	}
	return string(hashedPassword), nil
}
