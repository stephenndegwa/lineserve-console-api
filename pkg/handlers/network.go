package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/internal/services"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// NetworkHandler handles network related endpoints
type NetworkHandler struct {
	Client *client.OpenStackClient
}

// NewNetworkHandler creates a new network handler
func NewNetworkHandler(client *client.OpenStackClient) *NetworkHandler {
	return &NetworkHandler{
		Client: client,
	}
}

// ListNetworks handles listing all networks
func (h *NetworkHandler) ListNetworks(c *fiber.Ctx) error {
	// Create network service
	networkService := services.NewNetworkService(h.Client)

	// Get networks
	networks, err := networkService.ListNetworks()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to list networks: " + err.Error(),
		})
	}

	// Return networks
	return c.JSON(networks)
}

// GetNetwork handles getting a network by ID
func (h *NetworkHandler) GetNetwork(c *fiber.Ctx) error {
	// Get network ID
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Network ID is required",
		})
	}

	// Create network service
	networkService := services.NewNetworkService(h.Client)

	// Get network
	network, err := networkService.GetNetwork(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to get network: " + err.Error(),
		})
	}

	// Return network
	return c.JSON(network)
}
