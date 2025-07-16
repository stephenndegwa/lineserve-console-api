package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// FloatingIPHandler handles floating IP related endpoints
type FloatingIPHandler struct {
	Client *client.OpenStackClient
}

// NewFloatingIPHandler creates a new floating IP handler
func NewFloatingIPHandler(client *client.OpenStackClient) *FloatingIPHandler {
	return &FloatingIPHandler{
		Client: client,
	}
}

// ListFloatingIPs handles listing all floating IPs
func (h *FloatingIPHandler) ListFloatingIPs(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Network == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack network service unavailable",
		})
	}

	// Create floating IP service
	floatingIPService := services.NewFloatingIPService(h.Client)

	// Get floating IPs
	floatingIPs, err := floatingIPService.ListFloatingIPs()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to list floating IPs: %v", err),
		})
	}

	// Ensure we return an empty array instead of null
	if floatingIPs == nil {
		floatingIPs = []models.FloatingIP{}
	}

	// Return floating IPs
	return c.JSON(floatingIPs)
}

// CreateFloatingIP handles creating a new floating IP
func (h *FloatingIPHandler) CreateFloatingIP(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Network == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack network service unavailable",
		})
	}

	// Parse request body
	var req models.CreateFloatingIPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Create floating IP service
	floatingIPService := services.NewFloatingIPService(h.Client)

	// Create floating IP
	floatingIP, err := floatingIPService.CreateFloatingIP(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create floating IP: %v", err),
		})
	}

	// Return floating IP
	return c.Status(fiber.StatusCreated).JSON(floatingIP)
}

// GetFloatingIP handles getting a floating IP by ID
func (h *FloatingIPHandler) GetFloatingIP(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Network == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack network service unavailable",
		})
	}

	// Get floating IP ID
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Floating IP ID is required",
		})
	}

	// Create floating IP service
	floatingIPService := services.NewFloatingIPService(h.Client)

	// Get floating IP
	floatingIP, err := floatingIPService.GetFloatingIP(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to get floating IP: %v", err),
		})
	}

	// Return floating IP
	return c.JSON(floatingIP)
}

// UpdateFloatingIP handles updating a floating IP
func (h *FloatingIPHandler) UpdateFloatingIP(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Network == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack network service unavailable",
		})
	}

	// Get floating IP ID
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Floating IP ID is required",
		})
	}

	// Parse request body
	var req models.UpdateFloatingIPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Create floating IP service
	floatingIPService := services.NewFloatingIPService(h.Client)

	// Update floating IP
	floatingIP, err := floatingIPService.UpdateFloatingIP(id, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to update floating IP: %v", err),
		})
	}

	// Return floating IP
	return c.JSON(floatingIP)
}

// DeleteFloatingIP handles deleting a floating IP
func (h *FloatingIPHandler) DeleteFloatingIP(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Network == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack network service unavailable",
		})
	}

	// Get floating IP ID
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Floating IP ID is required",
		})
	}

	// Create floating IP service
	floatingIPService := services.NewFloatingIPService(h.Client)

	// Delete floating IP
	err := floatingIPService.DeleteFloatingIP(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to delete floating IP: %v", err),
		})
	}

	// Return success
	return c.Status(fiber.StatusNoContent).Send(nil)
}
