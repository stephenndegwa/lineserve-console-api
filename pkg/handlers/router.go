package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// RouterHandler handles router related requests
type RouterHandler struct {
	Service *services.RouterService
}

// NewRouterHandler creates a new router handler
func NewRouterHandler(client *client.OpenStackClient) *RouterHandler {
	return &RouterHandler{
		Service: services.NewRouterService(client),
	}
}

// ListRouters handles GET /api/v1/routers
func (h *RouterHandler) ListRouters(c *fiber.Ctx) error {
	routers, err := h.Service.ListRouters()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(routers)
}

// GetRouter handles GET /api/v1/routers/:id
func (h *RouterHandler) GetRouter(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "router ID is required",
		})
	}

	router, err := h.Service.GetRouter(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(router)
}

// CreateRouter handles POST /api/v1/routers
func (h *RouterHandler) CreateRouter(c *fiber.Ctx) error {
	var req models.CreateRouterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	router, err := h.Service.CreateRouter(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(router)
}

// DeleteRouter handles DELETE /api/v1/routers/:id
func (h *RouterHandler) DeleteRouter(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "router ID is required",
		})
	}

	err := h.Service.DeleteRouter(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// UpdateRouterInterfaces handles PUT /api/v1/routers/:id/interfaces
func (h *RouterHandler) UpdateRouterInterfaces(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "router ID is required",
		})
	}

	var req models.RouterInterfaceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check action parameter
	action := c.Query("action", "add")
	if action != "add" && action != "remove" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "action must be either 'add' or 'remove'",
		})
	}

	var routerInterface *models.RouterInterface
	var err error

	if action == "add" {
		routerInterface, err = h.Service.AddRouterInterface(id, req)
	} else {
		routerInterface, err = h.Service.RemoveRouterInterface(id, req)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(routerInterface)
}
