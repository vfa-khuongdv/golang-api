package e2e

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/routes"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// generateExpiredToken creates a JWT that is already expired for testing auth middleware
func generateExpiredToken(userID uint) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, services.CustomClaims{
		ID:    userID,
		Scope: services.TokenScopeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	})
	signed, err := token.SignedString([]byte("this-is-a-very-long-secret-key-for-e2e-testing-purposes-32-chars"))
	if err != nil {
		panic("failed to sign expired token: " + err.Error())
	}
	return signed
}

// ErrorResponse represents the standard error response structure
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func init() {
	// Change to project root to allow loading templates
	_ = os.Chdir("../..")
}

// setupTestRouter initializes the router with an in-memory SQLite database
func setupTestRouter() (*gin.Engine, *gorm.DB) {
	_ = os.Setenv("JWT_KEY", "this-is-a-very-long-secret-key-for-e2e-testing-purposes-32-chars")

	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Initialize in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
	)
	if err != nil {
		panic("failed to migrate test database")
	}

	// Initialize Validator
	utils.InitValidator()

	// Setup Router
	router := routes.SetupRouter(db)

	return router, db
}
