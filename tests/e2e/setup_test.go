package e2e

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/routes"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
