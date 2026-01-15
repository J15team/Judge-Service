package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	Server  ServerConfig
	Sandbox SandboxConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port    string
	Timeout time.Duration
}

// SandboxConfig holds sandbox configuration
type SandboxConfig struct {
	DefaultTimeLimit   int // milliseconds
	DefaultMemoryLimit int // MB
	MaxTimeLimit       int // milliseconds
	MaxMemoryLimit     int // MB
	CompileTimeout     time.Duration
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8081"),
			Timeout: time.Duration(getEnvInt("SERVER_TIMEOUT", 30)) * time.Second,
		},
		Sandbox: SandboxConfig{
			DefaultTimeLimit:   getEnvInt("DEFAULT_TIME_LIMIT", 2000),
			DefaultMemoryLimit: getEnvInt("DEFAULT_MEMORY_LIMIT", 256),
			MaxTimeLimit:       getEnvInt("MAX_TIME_LIMIT", 10000),
			MaxMemoryLimit:     getEnvInt("MAX_MEMORY_LIMIT", 512),
			CompileTimeout:     time.Duration(getEnvInt("COMPILE_TIMEOUT", 10)) * time.Second,
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
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
