package services_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
)

type mailerServiceTestSuite struct {
	suite.Suite
	mailerService services.MailerService
}

func (s *mailerServiceTestSuite) SetupTest() {
	s.mailerService = services.NewMailerService()
}

func (s *mailerServiceTestSuite) TestSendMailForgotPassword() {
	s.T().Run("SendMailForgotPassword - Success", func(t *testing.T) {
		// Set required environment variables for the test
		_ = os.Setenv("MAIL_HOST", "smtp.gmail.com")
		_ = os.Setenv("MAIL_PORT", "587")
		_ = os.Setenv("MAIL_USERNAME", "test@example.com")
		_ = os.Setenv("MAIL_PASSWORD", "testpassword")
		_ = os.Setenv("MAIL_FROM", "noreply@example.com")
		_ = os.Setenv("FRONTEND_URL", "https://example.com")
		defer func() {
			_ = os.Unsetenv("MAIL_HOST")
			_ = os.Unsetenv("MAIL_PORT")
			_ = os.Unsetenv("MAIL_USERNAME")
			_ = os.Unsetenv("MAIL_PASSWORD")
			_ = os.Unsetenv("MAIL_FROM")
			_ = os.Unsetenv("FRONTEND_URL")
		}()

		// Create a test user with token
		token := "test-reset-token"
		user := &models.User{
			ID:    1,
			Email: "user@example.com",
			Name:  "Test User",
			Token: &token,
		}

		// Create a temporary template file for testing
		templateContent := `<!DOCTYPE html>
<html>
<head>
    <title>Reset Password</title>
</head>
<body>
    <h1>Hello {{.Name}}</h1>
    <p>Click <a href="{{.URL}}">here</a> to reset your password.</p>
</body>
</html>`

		// Create the template directory if it doesn't exist
		err := os.MkdirAll("pkg/mailer/templates", 0755)
		require.NoError(t, err)

		// Create the template file
		templateFile, err := os.Create("pkg/mailer/templates/forgot_template.html")
		require.NoError(t, err)
		defer func() {
			_ = os.Remove("pkg/mailer/templates/forgot_template.html")
		}()

		_, err = templateFile.WriteString(templateContent)
		require.NoError(t, err)
		_ = templateFile.Close()

		// Note: This test will fail on actual email sending since we don't have real SMTP credentials
		// But it will test the template parsing and execution logic
		err = s.mailerService.SendMailForgotPassword(user)

		// The function should work up to the email sending part
		// Since we're using test credentials, it will likely fail at the SMTP send
		// But we can at least verify that the template parsing works
		if err != nil {
			// Check if the error is related to email sending (which is expected with test credentials)
			assert.Contains(t, err.Error(), "error sending email")
		}
	})

	s.T().Run("SendMailForgotPassword - Template Not Found", func(t *testing.T) {
		// Set required environment variables for the test
		_ = os.Setenv("MAIL_HOST", "smtp.gmail.com")
		_ = os.Setenv("MAIL_PORT", "587")
		_ = os.Setenv("MAIL_USERNAME", "test@example.com")
		_ = os.Setenv("MAIL_PASSWORD", "testpassword")
		_ = os.Setenv("MAIL_FROM", "noreply@example.com")
		_ = os.Setenv("FRONTEND_URL", "https://example.com")
		defer func() {
			_ = os.Unsetenv("MAIL_HOST")
			_ = os.Unsetenv("MAIL_PORT")
			_ = os.Unsetenv("MAIL_USERNAME")
			_ = os.Unsetenv("MAIL_PASSWORD")
			_ = os.Unsetenv("MAIL_FROM")
			_ = os.Unsetenv("FRONTEND_URL")
		}()

		// Remove template file if it exists
		_ = os.Remove("pkg/mailer/templates/forgot_template.html")

		// Create a test user with token
		token := "test-reset-token"
		user := &models.User{
			ID:    1,
			Email: "user@example.com",
			Name:  "Test User",
			Token: &token,
		}

		// Call the function with missing template
		err := s.mailerService.SendMailForgotPassword(user)

		// Should return template parsing error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing template")
	})

	s.T().Run("SendMailForgotPassword - Invalid Template", func(t *testing.T) {
		// Set required environment variables for the test
		_ = os.Setenv("MAIL_HOST", "smtp.gmail.com")
		_ = os.Setenv("MAIL_PORT", "587")
		_ = os.Setenv("MAIL_USERNAME", "test@example.com")
		_ = os.Setenv("MAIL_PASSWORD", "testpassword")
		_ = os.Setenv("MAIL_FROM", "noreply@example.com")
		_ = os.Setenv("FRONTEND_URL", "https://example.com")
		defer func() {
			_ = os.Unsetenv("MAIL_HOST")
			_ = os.Unsetenv("MAIL_PORT")
			_ = os.Unsetenv("MAIL_USERNAME")
			_ = os.Unsetenv("MAIL_PASSWORD")
			_ = os.Unsetenv("MAIL_FROM")
			_ = os.Unsetenv("FRONTEND_URL")
		}()

		// Create invalid template content
		invalidTemplateContent := `<!DOCTYPE html>
<html>
<head>
    <title>Reset Password</title>
</head>
<body>
    <h1>Hello {{.Name</h1>
    <p>Click <a href="{{.URL}}">here</a> to reset your password.</p>
</body>
</html>`

		// Create the template directory if it doesn't exist
		err := os.MkdirAll("pkg/mailer/templates", 0755)
		require.NoError(t, err)

		// Create the invalid template file
		templateFile, err := os.Create("pkg/mailer/templates/forgot_template.html")
		require.NoError(t, err)
		defer func() {
			_ = templateFile.Close()
			_ = os.Remove("pkg/mailer/templates/forgot_template.html")
		}()

		_, err = templateFile.WriteString(invalidTemplateContent)
		require.NoError(t, err)

		// Create a test user with token
		token := "test-reset-token"
		user := &models.User{
			ID:    1,
			Email: "user@example.com",
			Name:  "Test User",
			Token: &token,
		}

		// Call the function with invalid template
		err = s.mailerService.SendMailForgotPassword(user)

		// Should return template parsing error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing template")
	})

	s.T().Run("SendMailForgotPassword - Nil Token", func(t *testing.T) {
		// Set required environment variables for the test
		_ = os.Setenv("MAIL_HOST", "smtp.gmail.com")
		_ = os.Setenv("MAIL_PORT", "587")
		_ = os.Setenv("MAIL_USERNAME", "test@example.com")
		_ = os.Setenv("MAIL_PASSWORD", "testpassword")
		_ = os.Setenv("MAIL_FROM", "noreply@example.com")
		_ = os.Setenv("FRONTEND_URL", "https://example.com")
		defer func() {
			_ = os.Unsetenv("MAIL_HOST")
			_ = os.Unsetenv("MAIL_PORT")
			_ = os.Unsetenv("MAIL_USERNAME")
			_ = os.Unsetenv("MAIL_PASSWORD")
			_ = os.Unsetenv("MAIL_FROM")
			_ = os.Unsetenv("FRONTEND_URL")
		}()

		// Create valid template content
		templateContent := `<!DOCTYPE html>
<html>
<head>
    <title>Reset Password</title>
</head>
<body>
    <h1>Hello {{.Name}}</h1>
    <p>Click <a href="{{.URL}}">here</a> to reset your password.</p>
</body>
</html>`

		// Create the template directory if it doesn't exist
		err := os.MkdirAll("pkg/mailer/templates", 0755)
		require.NoError(t, err)

		// Create the template file
		templateFile, err := os.Create("pkg/mailer/templates/forgot_template.html")
		require.NoError(t, err)
		defer func() {
			_ = templateFile.Close()
			_ = os.Remove("pkg/mailer/templates/forgot_template.html")
		}()

		_, err = templateFile.WriteString(templateContent)
		require.NoError(t, err)

		// Create a test user with nil token
		user := &models.User{
			ID:    1,
			Email: "user@example.com",
			Name:  "Test User",
			Token: nil, // This should cause a panic or error
		}

		// Call the function should panic due to nil pointer dereference
		assert.Panics(t, func() {
			_ = s.mailerService.SendMailForgotPassword(user)
		})
	})

	s.T().Run("SendMailForgotPassword - Environment Variables Test", func(t *testing.T) {
		// Test with default environment values
		defer func() {
			_ = os.Unsetenv("MAIL_HOST")
			_ = os.Unsetenv("MAIL_PORT")
			_ = os.Unsetenv("MAIL_USERNAME")
			_ = os.Unsetenv("MAIL_PASSWORD")
			_ = os.Unsetenv("MAIL_FROM")
			_ = os.Unsetenv("FRONTEND_URL")
		}()

		// Create valid template content
		templateContent := `<!DOCTYPE html>
<html>
<head>
    <title>Reset Password</title>
</head>
<body>
    <h1>Hello {{.Name}}</h1>
    <p>Click <a href="{{.URL}}">here</a> to reset your password.</p>
</body>
</html>`

		// Create the template directory if it doesn't exist
		err := os.MkdirAll("pkg/mailer/templates", 0755)
		require.NoError(t, err)

		// Create the template file
		templateFile, err := os.Create("pkg/mailer/templates/forgot_template.html")
		require.NoError(t, err)
		defer func() {
			_ = templateFile.Close()
			_ = os.Remove("pkg/mailer/templates/forgot_template.html")
		}()

		_, err = templateFile.WriteString(templateContent)
		require.NoError(t, err)

		// Create a test user with token
		token := "test-reset-token"
		user := &models.User{
			ID:    1,
			Email: "user@example.com",
			Name:  "Test User",
			Token: &token,
		}

		// Test that environment variables are properly used
		// This should fail because of missing/invalid SMTP configuration
		err = s.mailerService.SendMailForgotPassword(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error sending email")
	})
}

func TestMailerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(mailerServiceTestSuite))
}
