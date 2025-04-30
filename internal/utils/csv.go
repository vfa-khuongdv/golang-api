// csv.go
package utils

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
)

// ParseUserCSV parses a CSV file and returns a slice of User models
// Expected CSV format:
// Email,Password,Name,Birthday,Address,Gender
// john@example.com,password123,John Doe,1990-01-01,123 Main St,1
func ParseUserCSV(file multipart.File) ([]models.User, error) {
	reader := csv.NewReader(file)

	// Read header row
	header, err := reader.Read()
	if err != nil {
		return nil, errors.New(errors.ErrInvalidData, "Failed to read CSV header")
	}

	// Validate header columns
	expectedHeaders := []string{"email", "password", "name", "birthday", "address", "gender"}
	for i, h := range header {
		if strings.ToLower(h) != expectedHeaders[i] {
			return nil, errors.New(errors.ErrInvalidData, fmt.Sprintf("Invalid CSV header format. Expected: %v", strings.Join(expectedHeaders, ",")))
		}
	}

	var users []models.User
	lineNumber := 1 // Start at 1 to account for header already read

	// Read data rows
	for {
		lineNumber++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.New(errors.ErrInvalidData, fmt.Sprintf("Error reading line %d: %v", lineNumber, err))
		}

		// Validate record has correct number of fields
		if len(record) != len(expectedHeaders) {
			return nil, errors.New(errors.ErrInvalidData, fmt.Sprintf("Line %d has incorrect number of fields", lineNumber))
		}

		// Parse gender
		gender, err := strconv.ParseInt(record[5], 10, 16)
		if err != nil {
			return nil, errors.New(errors.ErrInvalidData, fmt.Sprintf("Invalid gender value at line %d: %v", lineNumber, err))
		}
		if gender < 0 || gender > 2 {
			return nil, errors.New(errors.ErrInvalidData, fmt.Sprintf("Gender at line %d must be 0, 1, or 2", lineNumber))
		}

		// Hash password
		hashedPassword := HashPassword(record[1])
		if hashedPassword == "" {
			return nil, errors.New(errors.ErrAuthPasswordHashFailed, fmt.Sprintf("Failed to hash password at line %d", lineNumber))
		}

		// Create birthday pointer if not empty
		var birthday *string
		if record[3] != "" {
			birthday = &record[3]
		}

		// Create address pointer if not empty
		var address *string
		if record[4] != "" {
			address = &record[4]
		}

		user := models.User{
			Email:    record[0],
			Password: hashedPassword,
			Name:     record[2],
			Birthday: birthday,
			Address:  address,
			Gender:   int16(gender),
		}

		users = append(users, user)
	}

	return users, nil
}
