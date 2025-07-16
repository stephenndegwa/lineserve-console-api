package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// SubnetHandler handles subnet related requests
type SubnetHandler struct {
	Service *services.SubnetService
}

// NewSubnetHandler creates a new subnet handler
func NewSubnetHandler(client *client.OpenStackClient) *SubnetHandler {
	return &SubnetHandler{
		Service: services.NewSubnetService(client),
	}
}

// ListSubnets handles GET /api/v1/subnets
func (h *SubnetHandler) ListSubnets(c *fiber.Ctx) error {
	subnets, err := h.Service.ListSubnets()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(subnets)
}

// GetSubnet handles GET /api/v1/subnets/:id
func (h *SubnetHandler) GetSubnet(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "subnet ID is required",
		})
	}

	subnet, err := h.Service.GetSubnet(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(subnet)
}

// CreateSubnet handles POST /api/v1/subnets
func (h *SubnetHandler) CreateSubnet(c *fiber.Ctx) error {
	var req models.CreateSubnetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	subnet, err := h.Service.CreateSubnet(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(subnet)
}

// DeleteSubnet handles DELETE /api/v1/subnets/:id
func (h *SubnetHandler) DeleteSubnet(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "subnet ID is required",
		})
	}

	err := h.Service.DeleteSubnet(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
