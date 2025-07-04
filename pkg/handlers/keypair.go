package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// KeyPairHandler handles key pair related endpoints
type KeyPairHandler struct {
	Client *client.OpenStackClient
}

// NewKeyPairHandler creates a new key pair handler
func NewKeyPairHandler(client *client.OpenStackClient) *KeyPairHandler {
	return &KeyPairHandler{
		Client: client,
	}
}

// ListKeyPairs handles listing all key pairs
func (h *KeyPairHandler) ListKeyPairs(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Compute == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack compute service unavailable",
		})
	}

	// Create key pair service
	keyPairService := services.NewKeyPairService(h.Client)

	// Get key pairs
	keyPairs, err := keyPairService.ListKeyPairs()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to list key pairs: " + err.Error(),
		})
	}

	// Return key pairs
	return c.JSON(keyPairs)
}

// GetKeyPair handles getting a key pair by name
func (h *KeyPairHandler) GetKeyPair(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Compute == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack compute service unavailable",
		})
	}

	// Get key pair name
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Key pair name is required",
		})
	}

	// Create key pair service
	keyPairService := services.NewKeyPairService(h.Client)

	// Get key pair
	keyPair, err := keyPairService.GetKeyPair(name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to get key pair: " + err.Error(),
		})
	}

	// Return key pair
	return c.JSON(keyPair)
}

// CreateKeyPair handles creating a new key pair
func (h *KeyPairHandler) CreateKeyPair(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Compute == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack compute service unavailable",
		})
	}

	// Parse request body
	var req models.CreateKeyPairRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate request
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Name is required",
		})
	}

	// Create key pair service
	keyPairService := services.NewKeyPairService(h.Client)

	// Create key pair
	keyPair, err := keyPairService.CreateKeyPair(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create key pair: " + err.Error(),
		})
	}

	// Return key pair
	return c.Status(fiber.StatusCreated).JSON(keyPair)
}

// DeleteKeyPair handles deleting a key pair by name
func (h *KeyPairHandler) DeleteKeyPair(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Compute == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack compute service unavailable",
		})
	}

	// Get key pair name
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Key pair name is required",
		})
	}

	// Create key pair service
	keyPairService := services.NewKeyPairService(h.Client)

	// Delete key pair
	err := keyPairService.DeleteKeyPair(name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to delete key pair: " + err.Error(),
		})
	}

	// Return success
	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: "Key pair deleted successfully",
	})
}
