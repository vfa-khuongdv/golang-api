package routes

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/handlers"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	ginMode := configs.GetEnv("GIN_MODE", "release")
	gin.SetMode(ginMode)

	// Initialize the new Gin router
	router := gin.New()

	// By default Gin trusts ALL proxies and therefore honors a client-supplied
	// X-Forwarded-For header, which would let anyone spoof ClientIP() — the key
	// used by the rate limiter and stored on refresh tokens. Disable proxy
	// trust unless the operator explicitly opts in via TRUSTED_PROXIES
	// (comma-separated proxy/LB CIDRs, e.g. "10.0.0.0/8").
	trustedProxies := configs.GetEnv("TRUSTED_PROXIES", "")
	if trustedProxies == "" {
		if err := router.SetTrustedProxies(nil); err != nil {
			logger.Fatalf("Failed to disable trusted proxies: %v", err)
		}
	} else {
		proxies := strings.Split(trustedProxies, ",")
		for i := range proxies {
			proxies[i] = strings.TrimSpace(proxies[i])
		}
		if err := router.SetTrustedProxies(proxies); err != nil {
			logger.Fatalf("Failed to configure trusted proxies %q: %v", trustedProxies, err)
		}
	}

	stage := configs.GetEnv("STAGE", "dev")

	// Set up Swagger documentation only in non-production environments
	if stage != "prod" {
		router.StaticFile("/docs/swagger.json", "./docs/swagger.json")
		router.StaticFile("/swagger", "./docs/swagger.html")
		router.StaticFile("/api-docs", "./docs/swagger.html")
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	refreshRepo := repositories.NewRefreshTokenRepository(db)

	// Initialize services
	refreshTokenService := services.NewRefreshTokenService(refreshRepo)
	mailerService := services.NewMailerService()
	userService := services.NewUserService(userRepo, mailerService)
	jwtService, err := services.NewJWTService()
	if err != nil {
		logger.Fatalf("Failed to initialize JWT service: %v", err)
	}
	authService := services.NewAuthService(userRepo, refreshTokenService, jwtService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, mailerService)

	// Add middleware
	router.Use(
		middlewares.RequestIDMiddleware(),
		middlewares.CORSMiddleware(),
		middlewares.LogMiddleware(),
		gin.Recovery(),
	)

	router.GET("/healthz", handlers.HealthCheck)
	router.GET("/api/v1/version", handlers.VersionInfo)

	// Setup API routes
	api := router.Group("/api/v1")
	{
		// Public routes with rate limiting.
		// EmptyBodyMiddleware is applied here (not globally) because it would
		// otherwise reject legitimate body-less POSTs like /logout.
		public := api.Group("/")
		public.Use(middlewares.RateLimiter(10, time.Minute), middlewares.EmptyBodyMiddleware())
		{
			public.POST("/login", authHandler.Login)
			public.POST("/refresh-token", authHandler.RefreshToken)
			public.POST("/forgot-password", userHandler.ForgotPassword)
			public.POST("/reset-password", userHandler.ResetPassword)
		}

		authenticated := api.Group("/")
		authenticated.Use(middlewares.AuthMiddleware(jwtService))
		{
			authenticated.POST("/logout", authHandler.Logout)
			authenticated.POST("/change-password", middlewares.EmptyBodyMiddleware(), userHandler.ChangePassword)
			authenticated.GET("/profile", userHandler.GetProfile)
			authenticated.PATCH("/profile", middlewares.EmptyBodyMiddleware(), userHandler.UpdateProfile)
		}
	}

	return router
}
