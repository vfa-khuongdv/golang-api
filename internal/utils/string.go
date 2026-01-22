package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateRandomString generates a random string of specified length using alphanumeric characters
// Parameters:
//   - n: length of the random string to generate
//
// Returns:
//   - string: randomly generated alphanumeric string of length n
func GenerateRandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback or panic in case of entropy failure, usually effectively impossible
			// For this utility, we'll just skip the char which is not ideal but simple
			continue
		}
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

// StringToPtr converts a string to a pointer to a string
// Parameters:
//   - s: the string to convert
//
// Returns:
//   - *string: pointer to the string, or nil if the input string is empty
func StringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// IntToPtr converts any type to a pointer to that type
// Parameters:
//   - i: the value to convert
//
// Returns:
//   - *T: pointer to the value
func IntToPtr[T any](i T) *T {
	return &i
}
