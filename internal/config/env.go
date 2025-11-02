package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	AwsAccessKey string
	AwsSecretKey string
	AwsRegion    string
	SslCertPath  string
}

// LoadConfig loads the environment variables and return config struct
func LoadConfig() *Config {

	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		AwsAccessKey: getEnv("AWS_ACCESS_KEY", ""),
		AwsSecretKey: getEnv("AWS_SECRET_KEY", ""),
		AwsRegion:    getEnv("AWS_REGION", "us-east-1"),
		SslCertPath:  getEnv("SSL_CERT_PATH", ""),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	return cfg
}

// Helper to read environment variables with a default fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
