package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Feishu   FeishuConfig
	AI       AIConfig
}

type ServerConfig struct {
	Port         int
	Host         string
	Environment  string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret          string
	AccessTokenExp  int // hours
	RefreshTokenExp int // days
}

type FeishuConfig struct {
	AppID       string
	AppSecret   string
	RedirectURI string
}

type AIConfig struct {
	Mode            string // cloud, local, hybrid
	CloudAPIKey     string
	CloudModel      string
	LocalModelPath  string
	LocalModelPort  int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:        getEnvInt("SERVER_PORT", 8080),
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "testmind"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "testmind"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "testmind-secret-key"),
			AccessTokenExp:  getEnvInt("JWT_ACCESS_TOKEN_EXP", 2),
			RefreshTokenExp: getEnvInt("JWT_REFRESH_TOKEN_EXP", 7),
		},
		Feishu: FeishuConfig{
			AppID:       getEnv("FEISHU_APP_ID", ""),
			AppSecret:   getEnv("FEISHU_APP_SECRET", ""),
			RedirectURI: getEnv("FEISHU_REDIRECT_URI", ""),
		},
		AI: AIConfig{
			Mode:           getEnv("AI_MODE", "cloud"),
			CloudAPIKey:    getEnv("AI_CLOUD_API_KEY", ""),
			CloudModel:     getEnv("AI_CLOUD_MODEL", "gpt-4"),
			LocalModelPath: getEnv("AI_LOCAL_MODEL_PATH", ""),
			LocalModelPort: getEnvInt("AI_LOCAL_MODEL_PORT", 8000),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}