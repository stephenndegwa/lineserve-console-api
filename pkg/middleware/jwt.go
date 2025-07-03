package middleware

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lineserve/lineserve-api/pkg/client"
)

// JWTProtected returns a JWT middleware with the secret from environment
func JWTProtected() fiber.Handler {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_secret" // Default secret if not set
	}
	fmt.Printf("Using JWT secret: %s\n", jwtSecret)
	return JWTMiddleware(jwtSecret)
}

// JWTMiddleware creates a JWT middleware
func JWTMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		// Check if the header has the Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		fmt.Printf("Token received: %s\n", tokenString)

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				fmt.Printf("Invalid signing method: %v\n", token.Header["alg"])
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			fmt.Printf("Error parsing token: %v\n", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Check if the token is valid
		if !token.Valid {
			fmt.Println("Token is invalid")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Get claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Println("Invalid token claims")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Check if the token is expired
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				fmt.Println("Token has expired")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token expired",
				})
			}
		}

		fmt.Printf("Token is valid, username: %v\n", claims["username"])

		// Store claims in context
		c.Locals("username", claims["username"])
		c.Locals("project_id", claims["project_id"])
		c.Locals("project_name", claims["project_name"])
		c.Locals("domain_name", claims["domain_name"])
		c.Locals("region_name", claims["region_name"])
		c.Locals("auth_url", claims["auth_url"])

		// Continue
		return c.Next()
	}
}

// GetOpenStackClient creates an OpenStack client from JWT claims
func GetOpenStackClient(c *fiber.Ctx) (*client.OpenStackClient, error) {
	// Create a new OpenStack client
	return client.NewOpenStackClient()
}
