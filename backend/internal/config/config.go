package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL      string
	EthereumRPC      string
	ContractAddress  string
	JWTSecret        string
	AIModelEndpoint  string
	RedisURL         string
	Environment      string
	Port             int
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))

	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://localhost/invoice_financing?sslmode=disable"),
		EthereumRPC:     getEnv("ETHEREUM_RPC", "http://localhost:8545"),
		ContractAddress: getEnv("CONTRACT_ADDRESS", "0x0000000000000000000000000000000000000000"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		AIModelEndpoint: getEnv("AI_MODEL_ENDPOINT", "http://localhost:5000/api/ml"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		Port:           port,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
