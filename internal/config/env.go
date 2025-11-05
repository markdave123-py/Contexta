package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	AwsAccessKey string
	AwsSecretKey string
	AwsRegion    string
	BucketName   string
	SslCertPath  string
	AIAPIKey     string
	EmbedModel   string
	EmbedDim     int
	GenModel     string
	Port         string
}

// LoadConfig loads the environment variables and return config
func LoadConfig() *Config {

	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		AwsAccessKey: getEnv("AWS_ACCESS_KEY", ""),
		AwsSecretKey: getEnv("AWS_SECRET_KEY", ""),
		AwsRegion:    getEnv("AWS_REGION", "us-east-2"),
		BucketName:   getEnv("BUCKET_NAME", "contexta-docs"),
		SslCertPath:  getEnv("SSL_CERT_PATH", ""),
		AIAPIKey:     getEnv("GEMINI_API_KEY", ""),
		EmbedModel:   getEnv("EMBED_MODEL", "text-embedding-004"),
		EmbedDim:     getEnvInt("EMBED_DIM", 1536),
		GenModel:     getEnv("GEN_MODEL", "gemini-1.5-flash"),
		Port:         getEnv("PORT", "8080"),
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

func getEnvInt(key string, def int) int {
	v := getEnv(key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("WARN: %s=%q not an int, using default %d", key, v, def)
		return def
	}
	return n
}
