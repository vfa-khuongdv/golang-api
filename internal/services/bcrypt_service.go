package services

import (
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
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

func (s *bcryptServiceImpl) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperror.NewInternalServerError(err.Error())
	}
	return string(hashedPassword), nil
}

func (s *bcryptServiceImpl) CheckPasswordHash(password, hashPassword string) bool {
	return utils.CheckPasswordHash(password, hashPassword)
}

func (s *bcryptServiceImpl) HashPasswordWithCost(password string, cost int) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", apperror.NewInternalServerError(err.Error())
	}
	return string(hashedPassword), nil
}
