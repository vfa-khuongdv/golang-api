package routes

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/handlers"
	"github.com/vfa-khuongdv/golang-cms/internal/middlewares"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/constants"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

// maxRequestBodyBytes caps request bodies; the API only takes small JSON.
const maxRequestBodyBytes = 1 << 20 // 1 MB

func SetupRouter(db *gorm.DB) *gin.Engine {
	ginMode := configs.GetEnv("GIN_MODE", "release")
	gin.SetMode(ginMode)

	// Initialize the new Gin router
	router := gin.New()

	// ClientIP() is the key used by the rate limiter and stored on refresh
	// tokens. Behind a load balancer (AWS ALB) the peer is the balancer, so
	// without trusting it every client would share one IP and one rate limit.
	// The default trusts every peer (0.0.0.0/0): the real client IP is read from
	// X-Forwarded-For, but a client can also spoof it, which weakens the per-IP
	// rate limit. Set TRUSTED_PROXIES to the proxy CIDRs (e.g. "10.0.0.0/8") to
	// close that, or to an empty value to trust nothing when exposed directly.
	trustedProxies := configs.GetEnv("TRUSTED_PROXIES", "0.0.0.0/0")
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

	// The default is kept so existing deployments behind a load balancer keep
	// working, but in prod it lets clients spoof their IP, so say so.
	if _, set := os.LookupEnv("TRUSTED_PROXIES"); !set && stage == "prod" {
		logger.Warnf("TRUSTED_PROXIES is not set, so X-Forwarded-For is trusted from every peer: clients can spoof their IP and bypass the per-IP rate limit. Set it to the load balancer CIDR, e.g. 10.0.0.0/16")
	}

	// Set up Swagger documentation only in non-production environments
	if stage != "prod" {
		router.StaticFile("/docs/swagger.json", "./docs/swagger.json")
		router.StaticFile("/swagger", "./docs/swagger.html")
		router.StaticFile("/api-docs", "./docs/swagger.html")
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	refreshRepo := repositories.NewRefreshTokenRepository(db)
	settingRepo := repositories.NewSettingRepository(db)
	roleRepo := repositories.NewRoleRepository(db)

	// Initialize services
	refreshTokenService := services.NewRefreshTokenService(refreshRepo)
	settingsEncryptionKey := strings.TrimSpace(configs.GetEnv("SETTINGS_ENCRYPTION_KEY", ""))
	if len(settingsEncryptionKey) < utils.MinSecretKeyLength {
		logger.Fatalf("SETTINGS_ENCRYPTION_KEY must be at least %d characters", utils.MinSecretKeyLength)
	}
	mailerService := services.NewMailerService(settingRepo, settingsEncryptionKey)
	userService := services.NewUserService(userRepo, mailerService, refreshTokenService)
	settingService := services.NewSettingService(settingRepo, settingsEncryptionKey)
	roleService := services.NewRoleService(roleRepo, userRepo)
	jwtService, err := services.NewJWTService()
	if err != nil {
		logger.Fatalf("Failed to initialize JWT service: %v", err)
	}
	authService := services.NewAuthService(userRepo, refreshTokenService, jwtService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, mailerService)
	settingHandler := handlers.NewSettingHandler(settingService)
	roleHandler := handlers.NewRoleHandler(roleService)

	// Add middleware
	// Recovery goes first so a panic in any later middleware still gets a
	// response; the body limit goes before LogMiddleware, which reads the body.
	router.Use(
		gin.Recovery(),
		middlewares.RequestIDMiddleware(),
		middlewares.CORSMiddleware(),
		middlewares.BodySizeLimit(maxRequestBodyBytes),
		middlewares.LogMiddleware(),
	)

	router.GET("/healthz", handlers.HealthCheck)
	router.GET("/api/v1/version", handlers.VersionInfo)

	// Setup API routes
	api := router.Group("/api/v1")
	{
		// Public routes with rate limiting. Empty request bodies are rejected
		// centrally by TranslateValidationErrors (io.EOF -> ErrEmptyData), so
		// body-less POSTs like /logout are unaffected.
		public := api.Group("/")
		public.Use(middlewares.RateLimiter(10, time.Minute))
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
			authenticated.POST("/change-password", userHandler.ChangePassword)
			authenticated.GET("/profile", userHandler.GetProfile)
			authenticated.PATCH("/profile", userHandler.UpdateProfile)

			// Routes below require a permission granted by one of the user's roles
			require := func(permission string) gin.HandlerFunc {
				return middlewares.RequirePermission(roleService, permission)
			}
			authenticated.GET("/settings", require(constants.PermissionSettingsRead), settingHandler.GetSettings)
			authenticated.PUT("/settings", require(constants.PermissionSettingsUpdate), settingHandler.UpdateSettings)

			authenticated.GET("/roles", require(constants.PermissionRolesRead), roleHandler.ListRoles)
			authenticated.POST("/roles", require(constants.PermissionRolesCreate), roleHandler.CreateRole)
			authenticated.GET("/roles/:id", require(constants.PermissionRolesRead), roleHandler.GetRole)
			authenticated.PUT("/roles/:id", require(constants.PermissionRolesUpdate), roleHandler.UpdateRole)
			authenticated.DELETE("/roles/:id", require(constants.PermissionRolesDelete), roleHandler.DeleteRole)
			authenticated.GET("/permissions", require(constants.PermissionPermissionsRead), roleHandler.ListPermissions)
			authenticated.PUT("/users/:id/roles", require(constants.PermissionUsersAssign), roleHandler.SetUserRoles)
		}
	}

	return router
}
