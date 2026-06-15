package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrBcryptHashFailed = errors.New("bcrypt hashing failed")

	// CheckPasswordHash compares a plain text password with a hashed password.
	// Returns true if they match, false otherwise.
	CheckPasswordHash = func(password, hashPassword string) bool {
		err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
		return err == nil
	}

	// HashPasswordWithCost hashes a password using the given bcrypt cost.
	HashPasswordWithCost = func(password string, cost int) (string, error) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
		if err != nil {
			return "", errors.Join(ErrBcryptHashFailed, err)
		}
		return string(hashedPassword), nil
	}

	// HashPassword hashes a password using the default bcrypt cost.
	HashPassword = func(password string) (string, error) {
		return HashPasswordWithCost(password, bcrypt.DefaultCost)
	}
)
