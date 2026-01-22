package utils

import (
	"math/rand"
	"time"
)

// GenerateRandomString generates a random string of specified length using alphanumeric characters
// Parameters:
//   - n: length of the random string to generate
//
// Returns:
//   - string: randomly generated alphanumeric string of length n
func GenerateRandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := make([]byte, n)
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
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
