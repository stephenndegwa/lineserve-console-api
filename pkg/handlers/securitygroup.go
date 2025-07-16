package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// SecurityGroupHandler handles security group related requests
type SecurityGroupHandler struct {
	Service *services.SecurityGroupService
}

// NewSecurityGroupHandler creates a new security group handler
func NewSecurityGroupHandler(client *client.OpenStackClient) *SecurityGroupHandler {
	return &SecurityGroupHandler{
		Service: services.NewSecurityGroupService(client),
	}
}

// ListSecurityGroups handles GET /api/v1/security-groups
func (h *SecurityGroupHandler) ListSecurityGroups(c *fiber.Ctx) error {
	securityGroups, err := h.Service.ListSecurityGroups()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(securityGroups)
}

// GetSecurityGroup handles GET /api/v1/security-groups/:id
func (h *SecurityGroupHandler) GetSecurityGroup(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "security group ID is required",
		})
	}

	securityGroup, err := h.Service.GetSecurityGroup(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(securityGroup)
}

// CreateSecurityGroup handles POST /api/v1/security-groups
func (h *SecurityGroupHandler) CreateSecurityGroup(c *fiber.Ctx) error {
	var req models.CreateSecurityGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	securityGroup, err := h.Service.CreateSecurityGroup(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(securityGroup)
}

// DeleteSecurityGroup handles DELETE /api/v1/security-groups/:id
func (h *SecurityGroupHandler) DeleteSecurityGroup(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "security group ID is required",
		})
	}

	err := h.Service.DeleteSecurityGroup(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ListSecurityGroupRules handles GET /api/v1/security-group-rules
func (h *SecurityGroupHandler) ListSecurityGroupRules(c *fiber.Ctx) error {
	securityGroupRules, err := h.Service.ListSecurityGroupRules()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(securityGroupRules)
}

// CreateSecurityGroupRule handles POST /api/v1/security-group-rules
func (h *SecurityGroupHandler) CreateSecurityGroupRule(c *fiber.Ctx) error {
	var req models.CreateSecurityGroupRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	securityGroupRule, err := h.Service.CreateSecurityGroupRule(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(securityGroupRule)
}

// DeleteSecurityGroupRule handles DELETE /api/v1/security-group-rules/:id
func (h *SecurityGroupHandler) DeleteSecurityGroupRule(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "security group rule ID is required",
		})
	}

	err := h.Service.DeleteSecurityGroupRule(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
