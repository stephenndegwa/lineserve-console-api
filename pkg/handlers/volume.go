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
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Volume == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack volume service unavailable",
		})
	}

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

// ListVolumeTypes handles listing all volume types
func (h *VolumeHandler) ListVolumeTypes(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Volume == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack volume service unavailable",
		})
	}

	// Create volume service
	volumeService := services.NewVolumeService(h.Client)

	// Get volume types
	volumeTypes, err := volumeService.ListVolumeTypes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to list volume types: " + err.Error(),
		})
	}

	// Return volume types
	return c.JSON(volumeTypes)
}

// CreateVolume handles creating a new volume
func (h *VolumeHandler) CreateVolume(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Volume == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack volume service unavailable",
		})
	}

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
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Volume == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack volume service unavailable",
		})
	}

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

// DeleteVolume handles deleting a volume by ID
func (h *VolumeHandler) DeleteVolume(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Volume == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack volume service unavailable",
		})
	}

	// Get volume ID
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Volume ID is required",
		})
	}

	// Create volume service
	volumeService := services.NewVolumeService(h.Client)

	// Delete volume
	err := volumeService.DeleteVolume(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to delete volume: " + err.Error(),
		})
	}

	// Return success
	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: "Volume deleted successfully",
	})
}

// AttachVolume handles attaching a volume to an instance
func (h *VolumeHandler) AttachVolume(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Compute == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack compute service unavailable",
		})
	}

	// Get volume ID
	volumeID := c.Params("id")
	if volumeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Volume ID is required",
		})
	}

	// Parse request body
	var req models.VolumeAttachRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate request
	if req.InstanceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Instance ID is required",
		})
	}

	// Create volume service
	volumeService := services.NewVolumeService(h.Client)

	// Attach volume
	attachment, err := volumeService.AttachVolume(volumeID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to attach volume: " + err.Error(),
		})
	}

	// Return attachment
	return c.Status(fiber.StatusOK).JSON(attachment)
}

// DetachVolume handles detaching a volume from an instance
func (h *VolumeHandler) DetachVolume(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Compute == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack compute service unavailable",
		})
	}

	// Get volume ID
	volumeID := c.Params("id")
	if volumeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Volume ID is required",
		})
	}

	// Parse request body
	var req models.VolumeDetachRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Create volume service
	volumeService := services.NewVolumeService(h.Client)

	// Detach volume
	err := volumeService.DetachVolume(volumeID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to detach volume: " + err.Error(),
		})
	}

	// Return success
	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: "Volume detached successfully",
	})
}

// ResizeVolume handles resizing a volume
func (h *VolumeHandler) ResizeVolume(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Volume == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack volume service unavailable",
		})
	}

	// Get volume ID
	volumeID := c.Params("id")
	if volumeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Volume ID is required",
		})
	}

	// Parse request body
	var req models.VolumeResizeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate request
	if req.NewSize <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "New size must be greater than 0",
		})
	}

	// Create volume service
	volumeService := services.NewVolumeService(h.Client)

	// Get current volume to check current size
	volume, err := volumeService.GetVolume(volumeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to get volume: " + err.Error(),
		})
	}

	// Check if new size is greater than current size
	if req.NewSize <= volume.Size {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "New size must be greater than current size",
		})
	}

	// Resize volume
	err = volumeService.ResizeVolume(volumeID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to resize volume: " + err.Error(),
		})
	}

	// Return success
	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: "Volume resize operation started",
	})
}
