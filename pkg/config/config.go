package config

import (
	"fmt"
	"os"
	"regexp"
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
	OSInterface          string
	OSMemberRoleID       string

	// PostgreSQL Configuration
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string
	PostgresSSLMode  string
	DatabaseURL      string

	// API Configuration
	APIPort   string
	JWTSecret string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Warning: .env file not found or could not be loaded: %v\n", err)
	}

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
		OSInterface:          getEnv("OS_INTERFACE", "public"),
		OSMemberRoleID:       getEnv("OPENSTACK_MEMBER_ROLE_ID", ""),

		// PostgreSQL Configuration
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", ""),
		PostgresDBName:   getEnv("POSTGRES_DB", "lineserve"),
		PostgresSSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),

		// API Configuration
		APIPort:   getEnv("API_PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "lineserve-secret-key"),
	}

	return config, nil
}

// GetPostgresConnectionString returns the PostgreSQL connection string
func (c *Config) GetPostgresConnectionString() string {
	// If DATABASE_URL is provided, use it directly
	if c.DatabaseURL != "" {
		fmt.Println("Using DATABASE_URL for PostgreSQL connection")
		return c.DatabaseURL
	}

	// Otherwise, build the connection string from individual parameters
	fmt.Println("Building PostgreSQL connection string from individual parameters")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresDBName, c.PostgresSSLMode)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(strings.TrimSpace(value)) == 0 {
		return defaultValue
	}
	return value
}

// ParseDatabaseURL parses a DATABASE_URL into its components
// Format: postgresql://username:password@host:port/database
func ParseDatabaseURL(url string) (host, port, user, password, dbname, sslmode string) {
	// Default values
	sslmode = "require" // Default for most cloud databases

	// Extract components using regex
	re := regexp.MustCompile(`postgresql://([^:]+):([^@]+)@([^:]+):(\d+)/(.+)`)
	matches := re.FindStringSubmatch(url)

	if len(matches) >= 6 {
		user = matches[1]
		password = matches[2]
		host = matches[3]
		port = matches[4]
		dbname = matches[5]
	}

	return
}
