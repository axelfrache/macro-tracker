package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	ServerPort  string
	FDCApiKey   string
}

func Load() (*Config, error) {
	config := &Config{
		DatabaseURL: getEnvOrDefault("DATABASE_URL", "postgres://localhost:5432/macro_tracker?sslmode=disable"),
		ServerPort:  getEnvOrDefault("SERVER_PORT", "8080"),
		FDCApiKey:   getEnvOrDefault("FDC_API_KEY", "VkIvae2DDaLi0qdVhHgk0vhG216IgfDlqBGgDOwU"),
	}
	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
