package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
)

// ImageHandler handles image related endpoints
type ImageHandler struct {
	ImageService *services.ImageService
}

// NewImageHandler creates a new image handler
func NewImageHandler(client *client.OpenStackClient) *ImageHandler {
	return &ImageHandler{
		ImageService: services.NewImageService(client),
	}
}

// ListImages lists all images
func (h *ImageHandler) ListImages(c *fiber.Ctx) error {
	// Get images from service
	images, err := h.ImageService.ListImages()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list images",
		})
	}

	// Return images
	return c.JSON(images)
}

// GetImage gets an image by ID
func (h *ImageHandler) GetImage(c *fiber.Ctx) error {
	// Get ID from params
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image ID is required",
		})
	}

	// Get image from service
	image, err := h.ImageService.GetImage(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get image",
		})
	}

	// Return image
	return c.JSON(image)
}
