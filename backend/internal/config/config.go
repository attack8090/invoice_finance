package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL              string
	FabricLedgerServiceURL   string
	FabricChannelName        string
	FabricChaincodeName      string
	FabricConnectionProfile  string
	FabricWallet             string
	FabricUser               string
	JWTSecret                string
	AIModelEndpoint          string
	RedisURL                 string
	Environment              string
	Port                     int
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))

	return &Config{
		DatabaseURL:              getEnv("DATABASE_URL", "mongodb://localhost:27017/invoice_financing"),
		FabricLedgerServiceURL:   getEnv("FABRIC_LEDGER_SERVICE_URL", "http://localhost:8086"),
		FabricChannelName:        getEnv("FABRIC_CHANNEL_NAME", "invoice-financing-channel"),
		FabricChaincodeName:      getEnv("FABRIC_CHAINCODE_NAME", "invoice-financing"),
		FabricConnectionProfile:  getEnv("FABRIC_CONNECTION_PROFILE", "./fabric-config/connection.yaml"),
		FabricWallet:             getEnv("FABRIC_WALLET", "./fabric-config/wallet"),
		FabricUser:               getEnv("FABRIC_USER", "appUser"),
		JWTSecret:                getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		AIModelEndpoint:          getEnv("AI_MODEL_ENDPOINT", "http://localhost:5000/api/ml"),
		RedisURL:                 getEnv("REDIS_URL", "redis://localhost:6379"),
		Environment:              getEnv("ENVIRONMENT", "development"),
		Port:                     port,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
