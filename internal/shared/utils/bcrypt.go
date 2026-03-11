package utils

import "golang.org/x/crypto/bcrypt"

// CheckPasswordHash compares a plain text password with a hashed password
// Returns true if they match, false otherwise
func CheckPasswordHash(password, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

func HashPasswordWithCost(password string, cost int) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return ""
	}
	return string(hashedPassword)
}

// Default wrapper to use in production
func HashPassword(password string) string {
	return HashPasswordWithCost(password, bcrypt.DefaultCost)
}
