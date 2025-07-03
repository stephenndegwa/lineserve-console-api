package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// ComputeHandler handles compute related endpoints
type ComputeHandler struct {
	ComputeService *services.ComputeService
}

// NewComputeHandler creates a new compute handler
func NewComputeHandler(client *client.OpenStackClient) *ComputeHandler {
	return &ComputeHandler{
		ComputeService: services.NewComputeService(client),
	}
}

// ListInstances lists all instances
func (h *ComputeHandler) ListInstances(c *fiber.Ctx) error {
	// Get instances from service
	instances, err := h.ComputeService.ListInstances()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list instances",
		})
	}

	// Return instances
	return c.JSON(instances)
}

// CreateInstance creates a new instance
func (h *ComputeHandler) CreateInstance(c *fiber.Ctx) error {
	// Parse request body
	var req models.CreateInstanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}
	if req.FlavorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "FlavorID is required",
		})
	}
	if req.ImageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ImageID is required",
		})
	}
	if req.NetworkID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "NetworkID is required",
		})
	}

	// Create instance
	instance, err := h.ComputeService.CreateInstance(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create instance",
		})
	}

	// Return instance
	return c.Status(fiber.StatusCreated).JSON(instance)
}

// GetInstance gets an instance by ID
func (h *ComputeHandler) GetInstance(c *fiber.Ctx) error {
	// Get ID from params
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Instance ID is required",
		})
	}

	// Get instance from service
	instance, err := h.ComputeService.GetInstance(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get instance",
		})
	}

	// Return instance
	return c.JSON(instance)
}

// ListFlavors lists all flavors
func (h *ComputeHandler) ListFlavors(c *fiber.Ctx) error {
	// Get flavors from service
	flavors, err := h.ComputeService.ListFlavors()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list flavors",
		})
	}

	// Return flavors
	return c.JSON(flavors)
}
