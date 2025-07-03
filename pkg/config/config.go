package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// OpenStack Authentication
	OSAuthURL            string
	OSUsername           string
	OSPassword           string
	OSProjectID          string
	OSProjectName        string
	OSDomainName         string
	OSUserDomainName     string
	OSProjectDomainName  string
	OSRegionName         string
	OSIdentityAPIVersion string

	// API Configuration
	APIPort   string
	JWTSecret string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		// OpenStack Authentication
		OSAuthURL:            getEnv("OS_AUTH_URL", ""),
		OSUsername:           getEnv("OS_USERNAME", ""),
		OSPassword:           getEnv("OS_PASSWORD", ""),
		OSProjectID:          getEnv("OS_PROJECT_ID", ""),
		OSProjectName:        getEnv("OS_PROJECT_NAME", ""),
		OSDomainName:         getEnv("OS_DOMAIN_NAME", "Default"),
		OSUserDomainName:     getEnv("OS_USER_DOMAIN_NAME", "Default"),
		OSProjectDomainName:  getEnv("OS_PROJECT_DOMAIN_NAME", "Default"),
		OSRegionName:         getEnv("OS_REGION_NAME", "RegionOne"),
		OSIdentityAPIVersion: getEnv("OS_IDENTITY_API_VERSION", "3"),

		// API Configuration
		APIPort:   getEnv("API_PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "lineserve-secret-key"),
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(strings.TrimSpace(value)) == 0 {
		return defaultValue
	}
	return value
}
