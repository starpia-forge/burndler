package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Database
	DBHost               string
	DBPort               string
	DBName               string
	DBUser               string
	DBPassword           string
	DBSSLMode            string
	DBMaxConnections     int
	DBMaxIdleConnections int
	DBConnectionLifetime time.Duration

	// Storage
	StorageMode        string
	S3Endpoint         string
	S3Region           string
	S3Bucket           string
	S3AccessKeyID      string
	S3SecretAccessKey  string
	S3UseSSL           bool
	S3PathPrefix       string
	LocalStoragePath   string
	LocalStorageMaxSize string

	// JWT
	JWTSecret            string
	JWTIssuer           string
	JWTAudience         string
	JWTExpiration       time.Duration
	JWTRefreshExpiration time.Duration

	// Server
	ServerPort          string
	ServerHost          string
	ServerReadTimeout   time.Duration
	ServerWriteTimeout  time.Duration
	ServerMaxRequestSize int64

	// CORS
	CORSAllowedOrigins []string

	// Build Worker
	BuildWorkerCount   int
	BuildTimeout       time.Duration
	BuildTempDir       string
	BuildRetentionDays int

	// Logging
	LogLevel  string
	LogFormat string
}

func Load() *Config {
	return &Config{
		// Database
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBName:               getEnv("DB_NAME", "burndler"),
		DBUser:               getEnv("DB_USER", "burndler"),
		DBPassword:           getEnv("DB_PASSWORD", "changeme"),
		DBSSLMode:            getEnv("DB_SSL_MODE", "disable"),
		DBMaxConnections:     getEnvAsInt("DB_MAX_CONNECTIONS", 25),
		DBMaxIdleConnections: getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5),
		DBConnectionLifetime: getEnvAsDuration("DB_CONNECTION_LIFETIME", "300s"),

		// Storage
		StorageMode:         getEnv("STORAGE_MODE", "local"),
		S3Endpoint:          getEnv("S3_ENDPOINT", "https://s3.amazonaws.com"),
		S3Region:            getEnv("S3_REGION", "us-east-1"),
		S3Bucket:            getEnv("S3_BUCKET", "burndler-artifacts"),
		S3AccessKeyID:       getEnv("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey:   getEnv("S3_SECRET_ACCESS_KEY", ""),
		S3UseSSL:            getEnvAsBool("S3_USE_SSL", true),
		S3PathPrefix:        getEnv("S3_PATH_PREFIX", "packages/"),
		LocalStoragePath:    getEnv("LOCAL_STORAGE_PATH", "/tmp/burndler/storage"),
		LocalStorageMaxSize: getEnv("LOCAL_STORAGE_MAX_SIZE", "10GB"),

		// JWT
		JWTSecret:            getEnv("JWT_SECRET", "changeme-generate-secure-secret"),
		JWTIssuer:           getEnv("JWT_ISSUER", "burndler"),
		JWTAudience:         getEnv("JWT_AUDIENCE", "burndler-api"),
		JWTExpiration:       getEnvAsDuration("JWT_EXPIRATION", "24h"),
		JWTRefreshExpiration: getEnvAsDuration("JWT_REFRESH_EXPIRATION", "168h"),

		// Server
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		ServerHost:          getEnv("SERVER_HOST", "0.0.0.0"),
		ServerReadTimeout:   getEnvAsDuration("SERVER_READ_TIMEOUT", "30s"),
		ServerWriteTimeout:  getEnvAsDuration("SERVER_WRITE_TIMEOUT", "30s"),
		ServerMaxRequestSize: getEnvAsInt64("SERVER_MAX_REQUEST_SIZE", 100*1024*1024), // 100MB

		// CORS
		CORSAllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),

		// Build Worker
		BuildWorkerCount:   getEnvAsInt("BUILD_WORKER_COUNT", 4),
		BuildTimeout:       getEnvAsDuration("BUILD_TIMEOUT", "30m"),
		BuildTempDir:       getEnv("BUILD_TEMP_DIR", "/tmp/burndler-builds"),
		BuildRetentionDays: getEnvAsInt("BUILD_RETENTION_DAYS", 7),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// Fallback to default
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}