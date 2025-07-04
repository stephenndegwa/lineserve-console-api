package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// ProjectScopeRequired is a middleware that ensures the request has a project-scoped token
func ProjectScopeRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if the token has a project scope
		hasProjectScope, ok := c.Locals("has_project_scope").(bool)
		if !ok || !hasProjectScope {
			return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
				Error: "This endpoint requires a project-scoped token",
			})
		}

		// Get the project ID from the token
		projectID, ok := c.Locals("project_id").(string)
		if !ok || projectID == "" {
			return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
				Error: "No project ID found in token",
			})
		}

		// Add the project ID to the request context for logging
		fmt.Printf("Request to %s with project scope: %s\n", c.Path(), projectID)

		// Continue
		return c.Next()
	}
}
