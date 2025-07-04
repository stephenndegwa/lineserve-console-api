package handlers

import (
	"fmt"

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
	fmt.Println("ListNetworks handler called")

	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Network == nil {
		fmt.Println("ERROR: OpenStack network service unavailable")
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack network service unavailable",
		})
	}

	// Create network service
	networkService := services.NewNetworkService(h.Client)

	// Get networks
	networks, err := networkService.ListNetworks()
	if err != nil {
		fmt.Printf("ERROR in networkService.ListNetworks: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to list networks: " + err.Error(),
		})
	}

	fmt.Printf("Handler received %d networks\n", len(networks))

	// Ensure we return an empty array instead of null
	if networks == nil {
		networks = []models.Network{}
	}

	// Return networks
	return c.JSON(networks)
}

// CreateNetwork handles creating a new network
func (h *NetworkHandler) CreateNetwork(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Network == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack network service unavailable",
		})
	}

	// Parse request body
	var req models.CreateNetworkRequest
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

	// Create network service
	networkService := services.NewNetworkService(h.Client)

	// Create network
	network, err := networkService.CreateNetwork(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create network: " + err.Error(),
		})
	}

	// Return network
	return c.Status(fiber.StatusCreated).JSON(network)
}

// GetNetwork handles getting a network by ID
func (h *NetworkHandler) GetNetwork(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Network == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack network service unavailable",
		})
	}

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

// DeleteNetwork handles deleting a network by ID
func (h *NetworkHandler) DeleteNetwork(c *fiber.Ctx) error {
	// Check if OpenStack client is available
	if h.Client == nil || h.Client.Network == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ErrorResponse{
			Error: "OpenStack network service unavailable",
		})
	}

	// Get network ID
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Network ID is required",
		})
	}

	// Create network service
	networkService := services.NewNetworkService(h.Client)

	// Delete network
	err := networkService.DeleteNetwork(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to delete network: " + err.Error(),
		})
	}

	// Return success
	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: "Network deleted successfully",
	})
}
