package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     int
	Database DatabaseConfig
	LogLevel string
	JWT      JWTConfig
	Webhook  WebhookConfig
	GitHub   GitHubConfig
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	PublicKey  string
	PrivateKey string
	Issuer     string
	Subject    string
	Algorithm  string
}

type WebhookConfig struct {
	CommitBuildURL string
}

type GitHubConfig struct {
	Token    string
	ApiBase  string
	Proxy    string
}

func Load() (*Config, error) {
	// 加载 .env 文件
	_ = godotenv.Load()

	port := 32767
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	config := &Config{
		Port: port,
		Database: DatabaseConfig{
			URL: getEnvRequired("DB_URL"),
		},
		LogLevel: getEnvDefault("LOG_LEVEL", "info"),
		JWT: JWTConfig{
			PublicKey:  getEnvRequired("API_PUBLIC_KEY"),
			PrivateKey: getEnvRequired("API_PRIVATE_KEY"),
			Issuer:     getEnvDefault("API_ISSUER", "MenthaMC"),
			Subject:    getEnvDefault("API_SUBJECT", "leaves-ci"),
			Algorithm:  getEnvDefault("API_ALGO", "ES256"),
		},
		Webhook: WebhookConfig{
			CommitBuildURL: os.Getenv("COMMIT_BUILD_WEBHOOK_URL"),
		},
		GitHub: GitHubConfig{
			Token:   os.Getenv("GITHUB_TOKEN"),
			ApiBase: getEnvDefault("GITHUB_API_BASE", "https://api.github.com"),
			Proxy:   os.Getenv("GITHUB_PROXY"),
		},
	}

	return config, nil
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Required environment variable " + key + " is not set")
	}
	return value
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
