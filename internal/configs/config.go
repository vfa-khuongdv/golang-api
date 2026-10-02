package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
)

type (
	Config struct {
		Server   ServerConfig
		Database DatabaseConfig
		JWT      JWTConfig
		Settings SettingsConfig
		CORS     CORSConfig
		App      AppConfig
	}

	ServerConfig struct {
		Port    string
		GinMode string
		Stage   string
		// TrustedProxies are the peers whose X-Forwarded-For is trusted; empty
		// trusts nobody. TrustedProxiesDefault is true when TRUSTED_PROXIES is
		// unset and every peer is trusted.
		TrustedProxies        []string
		TrustedProxiesDefault bool
	}

	JWTConfig struct {
		Secret string
	}

	SettingsConfig struct {
		EncryptionKey string
	}

	CORSConfig struct {
		AllowedOrigins []string
	}

	AppConfig struct {
		ServiceName string
		Version     string
		FrontendURL string
		RunMigrate  bool
	}
)

func Load() (*Config, error) {
	_ = godotenv.Load()

	trustedProxies, trustedProxiesSet := os.LookupEnv("TRUSTED_PROXIES")
	if !trustedProxiesSet {
		trustedProxies = "0.0.0.0/0"
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:                  GetEnv("PORT", "3000"),
			GinMode:               GetEnv("GIN_MODE", "release"),
			Stage:                 GetEnv("STAGE", "dev"),
			TrustedProxies:        splitList(trustedProxies),
			TrustedProxiesDefault: !trustedProxiesSet,
		},
		Database: DatabaseConfig{
			Host:            GetEnv("DB_HOST", "127.0.0.1"),
			Port:            GetEnv("DB_PORT", "3306"),
			User:            GetEnv("DB_USERNAME", ""),
			Password:        GetEnv("DB_PASSWORD", ""),
			DBName:          GetEnv("DB_DATABASE", ""),
			MaxOpenConns:    GetEnvAsInt("DB_MAX_OPEN_CONNS", DEFAULT_MAX_OPEN_CONNS),
			MaxIdleConns:    GetEnvAsInt("DB_MAX_IDLE_CONNS", DEFAULT_MAX_IDLE_CONNS),
			ConnMaxLifetime: GetEnvAsDuration("DB_CONN_MAX_LIFETIME", DEFAULT_CONN_MAX_LIFETIME),
			ConnMaxIdleTime: GetEnvAsDuration("DB_CONN_MAX_IDLE_TIME", DEFAULT_CONN_MAX_IDLE_TIME),
		},
		JWT: JWTConfig{
			Secret: strings.TrimSpace(GetEnv("JWT_KEY", "")),
		},
		Settings: SettingsConfig{
			EncryptionKey: strings.TrimSpace(GetEnv("SETTINGS_ENCRYPTION_KEY", "")),
		},
		CORS: CORSConfig{
			AllowedOrigins: splitList(GetEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		},
		App: AppConfig{
			ServiceName: GetEnv("APP_SERVICE", "golang-cms"),
			Version:     GetEnv("APP_VERSION", "dev"),
			RunMigrate:  GetEnv("RUN_MIGRATE", "false") == "true",
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	var missing []string

	if c.Server.Port == "" {
		missing = append(missing, "PORT")
	}
	if c.Database.User == "" {
		missing = append(missing, "DB_USERNAME")
	}
	if c.Database.Password == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if c.Database.DBName == "" {
		missing = append(missing, "DB_DATABASE")
	}
	if c.JWT.Secret == "" {
		missing = append(missing, "JWT_KEY")
	}
	if c.Settings.EncryptionKey == "" {
		missing = append(missing, "SETTINGS_ENCRYPTION_KEY")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	for name, secret := range map[string]string{"JWT_KEY": c.JWT.Secret, "SETTINGS_ENCRYPTION_KEY": c.Settings.EncryptionKey} {
		if len(secret) < utils.MinSecretKeyLength {
			return fmt.Errorf("%s must be at least %d characters", name, utils.MinSecretKeyLength)
		}
	}

	return nil
}

// splitList splits a comma-separated value and trims each item, dropping
// empty ones; an empty value gives an empty list.
func splitList(value string) []string {
	items := []string{}
	for item := range strings.SplitSeq(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	valueStr := GetEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func GetEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, err := time.ParseDuration(GetEnv(key, "")); err == nil {
		return value
	}
	return defaultValue
}
