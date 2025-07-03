package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// ProjectHandler handles project related endpoints
type ProjectHandler struct {
	Client *client.OpenStackClient
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(client *client.OpenStackClient) *ProjectHandler {
	return &ProjectHandler{
		Client: client,
	}
}

// ListProjects handles listing all projects
func (h *ProjectHandler) ListProjects(c *fiber.Ctx) error {
	// Create identity service
	identityService := services.NewIdentityService(h.Client)

	// Get projects
	projects, err := identityService.ListProjects()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to list projects: " + err.Error(),
		})
	}

	// Return projects
	return c.JSON(projects)
}

// GetProject handles getting a project by ID
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	// Get project ID
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Project ID is required",
		})
	}

	// Create identity service
	identityService := services.NewIdentityService(h.Client)

	// Get project
	project, err := identityService.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to get project: " + err.Error(),
		})
	}

	// Return project
	return c.JSON(project)
}
