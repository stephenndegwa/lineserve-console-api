package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// VolumeHandler handles volume related endpoints
type VolumeHandler struct {
	Client *client.OpenStackClient
}

// NewVolumeHandler creates a new volume handler
func NewVolumeHandler(client *client.OpenStackClient) *VolumeHandler {
	return &VolumeHandler{
		Client: client,
	}
}

// ListVolumes handles listing all volumes
func (h *VolumeHandler) ListVolumes(c *fiber.Ctx) error {
	// Create volume service
	volumeService := services.NewVolumeService(h.Client)

	// Get volumes
	volumes, err := volumeService.ListVolumes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to list volumes: " + err.Error(),
		})
	}

	// Return volumes
	return c.JSON(volumes)
}

// CreateVolume handles creating a new volume
func (h *VolumeHandler) CreateVolume(c *fiber.Ctx) error {
	// Parse request body
	var req models.CreateVolumeRequest
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
	if req.Size <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Size must be greater than 0",
		})
	}

	// Create volume service
	volumeService := services.NewVolumeService(h.Client)

	// Create volume
	volume, err := volumeService.CreateVolume(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create volume: " + err.Error(),
		})
	}

	// Return volume
	return c.Status(fiber.StatusCreated).JSON(volume)
}

// GetVolume handles getting a volume by ID
func (h *VolumeHandler) GetVolume(c *fiber.Ctx) error {
	// Get volume ID
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Volume ID is required",
		})
	}

	// Create volume service
	volumeService := services.NewVolumeService(h.Client)

	// Get volume
	volume, err := volumeService.GetVolume(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to get volume: " + err.Error(),
		})
	}

	// Return volume
	return c.JSON(volume)
}
