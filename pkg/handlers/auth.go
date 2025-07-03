package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/config"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// AuthHandler handles authentication
type AuthHandler struct {
	JWTSecret string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(jwtSecret string) *AuthHandler {
	fmt.Printf("Creating AuthHandler with JWT secret: %s\n", jwtSecret)
	return &AuthHandler{
		JWTSecret: jwtSecret,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// Parse request body
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" || req.AuthURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username, password, and authURL are required",
		})
	}

	// Create auth options
	authOptions := gophercloud.AuthOptions{
		IdentityEndpoint: req.AuthURL,
		Username:         req.Username,
		Password:         req.Password,
		TenantID:         req.ProjectID,
		TenantName:       req.ProjectName,
		DomainName:       req.DomainName,
	}

	// Create context
	ctx := c.Context()

	// Authenticate with OpenStack
	_, err := config.NewProviderClient(ctx, authOptions)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication failed",
		})
	}

	// Create JWT token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = req.Username
	claims["project_id"] = req.ProjectID
	claims["project_name"] = req.ProjectName
	claims["domain_name"] = req.DomainName
	claims["region_name"] = req.RegionName
	claims["auth_url"] = req.AuthURL
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// Generate encoded token
	fmt.Printf("Signing token with JWT secret: %s\n", h.JWTSecret)
	encodedToken, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// Return token
	return c.JSON(fiber.Map{
		"token": encodedToken,
		"user": fiber.Map{
			"username":     req.Username,
			"project_id":   req.ProjectID,
			"project_name": req.ProjectName,
			"domain_name":  req.DomainName,
			"region_name":  req.RegionName,
		},
	})
}
