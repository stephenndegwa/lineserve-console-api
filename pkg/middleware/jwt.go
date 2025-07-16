package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// JWTMiddleware creates a JWT middleware
func JWTMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: "Authorization header is required",
			})
		}

		// Check if the header has the Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: "Invalid authorization header format",
			})
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: fmt.Sprintf("Invalid token: %v", err),
			})
		}

		// Check if the token is valid
		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: "Invalid token",
			})
		}

		// Get claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: "Invalid token claims",
			})
		}

		// Check if the token is expired
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
					Error: "Token expired",
				})
			}
		}

		// Store claims in context
		c.Locals("user", claims)
		c.Locals("user_id", claims["user_id"])
		c.Locals("username", claims["username"])
		c.Locals("project_id", claims["project_id"])
		c.Locals("domain_name", claims["domain_name"])

		// If this is a project-scoped token, mark it in the context
		if projectID, ok := claims["project_id"].(string); ok && projectID != "" {
			// Just log that we have a project-scoped token
			c.Locals("has_project_scope", true)
		}

		// Continue
		return c.Next()
	}
}

// IsAdmin checks if the user has admin role
func IsAdmin(c *fiber.Ctx) bool {
	role := c.Locals("role")
	if role == nil {
		return false
	}

	// Check if role is admin
	return role.(string) == "admin"
}

// AdminRequired is middleware that checks if the user has admin role
func AdminRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if user is admin
		if !IsAdmin(c) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}

		// Continue
		return c.Next()
	}
}
