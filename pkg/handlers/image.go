package handlers

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
	"github.com/lineserve/lineserve-api/pkg/openstack"
)

// ImageHandler handles image related endpoints
type ImageHandler struct {
	JWTSecret string
}

// NewImageHandler creates a new image handler
func NewImageHandler(jwtSecret string) *ImageHandler {
	return &ImageHandler{
		JWTSecret: jwtSecret,
	}
}

// getProviderFromToken extracts the project-scoped provider from the JWT token
func (h *ImageHandler) getProviderFromToken(c *fiber.Ctx) (*gophercloud.ProviderClient, error) {
	// Get token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return nil, fmt.Errorf("invalid or missing Authorization header")
	}

	// Parse token
	tokenString := authHeader[7:]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Extract username, password, domain, and project ID
	username, _ := claims["username"].(string)
	projectID, ok := claims["project_id"].(string)
	if !ok || projectID == "" {
		return nil, fmt.Errorf("project_id not found in token")
	}

	domainName, ok := claims["domain_name"].(string)
	if !ok || domainName == "" {
		domainName = "Default" // Use default domain if not specified
	}

	// Get provider from context (if available)
	if provider, ok := c.Locals("provider").(*gophercloud.ProviderClient); ok && provider != nil {
		return provider, nil
	}

	// Try to get OpenStack token from JWT claims
	openstackToken, ok := claims["openstack_token"].(string)
	if ok && openstackToken != "" {
		// Use the OpenStack token to authenticate
		ctx := c.Context()
		provider, err := openstack.AuthenticateWithToken(ctx, openstackToken, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with OpenStack token: %v", err)
		}
		return provider, nil
	}

	// If no OpenStack token in claims, try header
	tokenID := c.Get("X-Auth-Token")
	if tokenID != "" {
		ctx := c.Context()
		provider, err := openstack.AuthenticateWithToken(ctx, tokenID, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with X-Auth-Token: %v", err)
		}
		return provider, nil
	}

	// Last resort: try to authenticate with username (will likely fail without password)
	ctx := c.Context()
	provider, err := openstack.AuthenticateScoped(ctx, username, "", domainName, projectID)
	if err != nil {
		return nil, fmt.Errorf("no authenticated provider found: %v", err)
	}

	return provider, nil
}

// ListImages lists all images
func (h *ImageHandler) ListImages(c *fiber.Ctx) error {
	// Get provider from token
	provider, err := h.getProviderFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Authentication error: %v", err),
		})
	}

	// Create image client
	imageClient, err := openstack.NewImageClient(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create image client: %v", err),
		})
	}

	// List images
	listOpts := images.ListOpts{
		Status: images.ImageStatusActive,
	}

	ctx := context.Background()
	allPages, err := images.List(imageClient, listOpts).AllPages(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to list images: %v", err),
		})
	}

	allImages, err := images.ExtractImages(allPages)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to extract images: %v", err),
		})
	}

	// Convert to our model
	result := make([]models.Image, len(allImages))
	for i, img := range allImages {
		// Convert properties
		properties := make(map[string]string)
		for k, v := range img.Properties {
			if strVal, ok := v.(string); ok {
				properties[k] = strVal
			}
		}

		result[i] = models.Image{
			ID:         img.ID,
			Name:       img.Name,
			Status:     string(img.Status),
			Size:       img.SizeBytes,
			Visibility: string(img.Visibility),
			Tags:       img.Tags,
			CreatedAt:  img.CreatedAt,
			UpdatedAt:  img.UpdatedAt,
			Properties: properties,
		}
	}

	// Return images
	return c.JSON(result)
}

// GetImage gets an image by ID
func (h *ImageHandler) GetImage(c *fiber.Ctx) error {
	// Get provider from token
	provider, err := h.getProviderFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Authentication error: %v", err),
		})
	}

	// Get ID from params
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Image ID is required",
		})
	}

	// Create image client
	imageClient, err := openstack.NewImageClient(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create image client: %v", err),
		})
	}

	// Get image
	ctx := context.Background()
	img, err := images.Get(ctx, imageClient, id).Extract()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to get image: %v", err),
		})
	}

	// Convert properties
	properties := make(map[string]string)
	for k, v := range img.Properties {
		if strVal, ok := v.(string); ok {
			properties[k] = strVal
		}
	}

	// Convert to our model
	result := models.Image{
		ID:         img.ID,
		Name:       img.Name,
		Status:     string(img.Status),
		Size:       img.SizeBytes,
		Visibility: string(img.Visibility),
		Tags:       img.Tags,
		CreatedAt:  img.CreatedAt,
		UpdatedAt:  img.UpdatedAt,
		Properties: properties,
	}

	// Return image
	return c.JSON(result)
}

// CreateImage creates a new image
func (h *ImageHandler) CreateImage(c *fiber.Ctx) error {
	// Get provider from token
	provider, err := h.getProviderFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Authentication error: %v", err),
		})
	}

	// Create OpenStack client
	openStackClient, err := client.NewOpenStackClient(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create OpenStack client: %v", err),
		})
	}

	// Parse request
	var createReq models.CreateImageRequest
	if err := c.BodyParser(&createReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Validate request
	if createReq.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Image name is required",
		})
	}

	// Set default values if not provided
	if createReq.DiskFormat == "" {
		createReq.DiskFormat = "raw"
	}
	if createReq.ContainerFormat == "" {
		createReq.ContainerFormat = "bare"
	}
	if createReq.Visibility == "" {
		createReq.Visibility = string(images.ImageVisibilityPrivate)
	}

	// Create image service
	imageService := services.NewImageService(openStackClient)

	// Create image
	image, err := imageService.CreateImage(createReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create image: %v", err),
		})
	}

	// If there's image data in the request, upload it
	file, err := c.FormFile("file")
	if err == nil && file != nil {
		// Open the uploaded file
		fileHandle, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: fmt.Sprintf("Failed to open uploaded file: %v", err),
			})
		}
		defer fileHandle.Close()

		// Upload the image data
		err = imageService.UploadImageData(image.ID, fileHandle)
		if err != nil {
			// If upload fails, try to delete the created image
			_ = imageService.DeleteImage(image.ID)
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: fmt.Sprintf("Failed to upload image data: %v", err),
			})
		}

		// Refresh image to get updated status and size
		image, err = imageService.GetImage(image.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: fmt.Sprintf("Image created but failed to refresh image data: %v", err),
			})
		}
	}

	// Return the created image
	return c.Status(fiber.StatusCreated).JSON(image)
}

// DeleteImage deletes an image by ID
func (h *ImageHandler) DeleteImage(c *fiber.Ctx) error {
	// Get provider from token
	provider, err := h.getProviderFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Authentication error: %v", err),
		})
	}

	// Get ID from params
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Image ID is required",
		})
	}

	// Create image client
	imageClient, err := openstack.NewImageClient(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create image client: %v", err),
		})
	}

	// Delete the image
	ctx := context.Background()
	err = images.Delete(ctx, imageClient, id).ExtractErr()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to delete image: %v", err),
		})
	}

	// Return success
	return c.JSON(models.SuccessResponse{
		Message: fmt.Sprintf("Image %s deleted successfully", id),
	})
}
