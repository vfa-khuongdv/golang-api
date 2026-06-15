package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type (
	Config struct {
		Server   ServerConfig
		Database DatabaseConfig
		JWT      JWTConfig
		Mail     MailConfig
		CORS     CORSConfig
		App      AppConfig
	}

	ServerConfig struct {
		Port    string
		GinMode string
		Stage   string
	}

	JWTConfig struct {
		Secret string
	}

	MailConfig struct {
		Host     string
		Port     int
		Username string
		Password string
		From     string
	}

	CORSConfig struct {
		AllowedOrigins []string
	}

	AppConfig struct {
		FrontendURL string
		RunMigrate  bool
	}
)

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:    GetEnv("PORT", "3000"),
			GinMode: GetEnv("GIN_MODE", "release"),
			Stage:   GetEnv("STAGE", "dev"),
		},
		Database: DatabaseConfig{
			Host:         GetEnv("DB_HOST", "127.0.0.1"),
			Port:         GetEnv("DB_PORT", "3306"),
			User:         GetEnv("DB_USERNAME", ""),
			Password:     GetEnv("DB_PASSWORD", ""),
			DBName:       GetEnv("DB_DATABASE", ""),
			MaxOpenConns: GetEnvAsInt("DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns: GetEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		},
		JWT: JWTConfig{
			Secret: strings.TrimSpace(GetEnv("JWT_KEY", "")),
		},
		Mail: MailConfig{
			Host:     GetEnv("MAIL_HOST", "smtp.gmail.com"),
			Port:     GetEnvAsInt("MAIL_PORT", 587),
			Username: GetEnv("MAIL_USERNAME", ""),
			Password: GetEnv("MAIL_PASSWORD", ""),
			From:     GetEnv("MAIL_FROM", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: strings.Split(GetEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"), ","),
		},
		App: AppConfig{
			FrontendURL: GetEnv("FRONTEND_URL", ""),
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

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
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
