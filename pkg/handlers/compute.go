package handlers

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/lineserve/lineserve-api/pkg/models"
	"github.com/lineserve/lineserve-api/pkg/openstack"
)

// ComputeHandler handles compute related endpoints
type ComputeHandler struct {
	JWTSecret string
}

// NewComputeHandler creates a new compute handler
func NewComputeHandler(jwtSecret string) *ComputeHandler {
	return &ComputeHandler{
		JWTSecret: jwtSecret,
	}
}

// getProviderFromToken extracts the project-scoped provider from the JWT token
func (h *ComputeHandler) getProviderFromToken(c *fiber.Ctx) (*gophercloud.ProviderClient, error) {
	// Get token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return nil, fmt.Errorf("invalid or missing Authorization header")
	}

	// Parse token
	tokenString := authHeader[7:]
	fmt.Printf("Parsing JWT token: %s\n", tokenString)

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

	fmt.Printf("JWT Claims: %+v\n", claims)

	// Extract username, password, domain, and project ID
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return nil, fmt.Errorf("username not found in token")
	}

	projectID, ok := claims["project_id"].(string)
	if !ok || projectID == "" {
		return nil, fmt.Errorf("project_id not found in token")
	}

	domainName, ok := claims["domain_name"].(string)
	if !ok || domainName == "" {
		domainName = "Default" // Use default domain if not specified
	}

	fmt.Printf("Token info - Username: %s, ProjectID: %s, DomainName: %s\n", username, projectID, domainName)

	// Get provider from context (if available)
	if provider, ok := c.Locals("provider").(*gophercloud.ProviderClient); ok && provider != nil {
		fmt.Println("Using provider from context")
		return provider, nil
	}

	// Try to get OpenStack token from JWT claims
	openstackToken, ok := claims["openstack_token"].(string)
	if ok && openstackToken != "" {
		fmt.Printf("Found OpenStack token in JWT claims: %s\n", openstackToken[:20]+"...")
		// Use the OpenStack token to authenticate
		ctx := c.Context()
		provider, err := openstack.AuthenticateWithToken(ctx, openstackToken, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with OpenStack token: %v", err)
		}
		return provider, nil
	} else {
		fmt.Println("No OpenStack token found in JWT claims")
	}

	// If no OpenStack token in claims, try header
	tokenID := c.Get("X-Auth-Token")
	if tokenID != "" {
		fmt.Printf("Found X-Auth-Token header: %s\n", tokenID[:20]+"...")
		ctx := c.Context()
		provider, err := openstack.AuthenticateWithToken(ctx, tokenID, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with X-Auth-Token: %v", err)
		}
		return provider, nil
	} else {
		fmt.Println("No X-Auth-Token header found")
	}

	// Last resort: try to authenticate with username (will likely fail without password)
	fmt.Println("Trying to authenticate with username (no password)")
	ctx := c.Context()
	provider, err := openstack.AuthenticateScoped(ctx, username, "", domainName, projectID)
	if err != nil {
		return nil, fmt.Errorf("no authenticated provider found: %v", err)
	}

	return provider, nil
}

// ListInstances lists all instances in the project
func (h *ComputeHandler) ListInstances(c *fiber.Ctx) error {
	// Get provider from token
	provider, err := h.getProviderFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Authentication error: %v", err),
		})
	}

	// Create compute client
	computeClient, err := openstack.NewComputeClient(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create compute client: %v", err),
		})
	}

	// List servers
	ctx := context.Background()
	allPages, err := servers.List(computeClient, servers.ListOpts{}).AllPages(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to list instances: %v", err),
		})
	}

	allServers, err := servers.ExtractServers(allPages)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to extract instances: %v", err),
		})
	}

	// Convert to our model
	instances := make([]models.Instance, len(allServers))
	for i, server := range allServers {
		// Convert addresses
		addresses := make(map[string][]models.Address)
		for networkName, networkAddresses := range server.Addresses {
			addrs := []models.Address{}
			for _, addr := range networkAddresses.([]interface{}) {
				addrMap := addr.(map[string]interface{})
				address := models.Address{
					Type:    addrMap["OS-EXT-IPS:type"].(string),
					Address: addrMap["addr"].(string),
				}
				addrs = append(addrs, address)
			}
			addresses[networkName] = addrs
		}

		// Convert metadata to map[string]interface{} if needed
		metadata := make(map[string]interface{})
		for k, v := range server.Metadata {
			metadata[k] = v
		}

		instances[i] = models.Instance{
			ID:        server.ID,
			Name:      server.Name,
			Status:    server.Status,
			Flavor:    server.Flavor["id"].(string),
			Image:     server.Image["id"].(string),
			Addresses: addresses,
			Created:   server.Created,
			Metadata:  metadata,
		}
	}

	// Return instances
	return c.JSON(instances)
}

// CreateInstance creates a new instance
func (h *ComputeHandler) CreateInstance(c *fiber.Ctx) error {
	// Parse request body
	var req models.CreateInstanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Create instance using internal method
	instance, err := h.createInstanceInternal(c, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create instance: %v", err),
		})
	}

	// Return instance
	return c.Status(fiber.StatusCreated).JSON(instance)
}

// createInstanceInternal is an internal method that creates a new instance
// This can be called by other handlers
func (h *ComputeHandler) createInstanceInternal(c *fiber.Ctx, req models.CreateInstanceRequest) (*models.Instance, error) {
	// Get provider from token
	provider, err := h.getProviderFromToken(c)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %v", err)
	}

	// Validate required fields
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.FlavorID == "" {
		return nil, fmt.Errorf("flavorID is required")
	}
	if req.ImageID == "" {
		return nil, fmt.Errorf("imageID is required")
	}
	if req.NetworkID == "" {
		return nil, fmt.Errorf("networkID is required")
	}

	// Create compute client
	computeClient, err := openstack.NewComputeClient(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %v", err)
	}

	// Create server
	createOpts := servers.CreateOpts{
		Name:      req.Name,
		FlavorRef: req.FlavorID,
		ImageRef:  req.ImageID,
		Networks: []servers.Network{
			{
				UUID: req.NetworkID,
			},
		},
	}

	// Add key name if provided
	var serverCreateOpts servers.CreateOptsBuilder = createOpts
	if req.KeyName != "" {
		type createOptsWithKeyName struct {
			servers.CreateOpts
			KeyName string `json:"key_name,omitempty"`
		}
		serverCreateOpts = createOptsWithKeyName{
			CreateOpts: createOpts,
			KeyName:    req.KeyName,
		}
	}

	// Create server
	ctx := context.Background()
	server, err := servers.Create(ctx, computeClient, serverCreateOpts, nil).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %v", err)
	}

	// Convert to our model
	instance := &models.Instance{
		ID:      server.ID,
		Name:    server.Name,
		Status:  server.Status,
		Flavor:  req.FlavorID,
		Image:   req.ImageID,
		Created: server.Created,
	}

	return instance, nil
}

// GetInstance gets an instance by ID
func (h *ComputeHandler) GetInstance(c *fiber.Ctx) error {
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
			Error: "Instance ID is required",
		})
	}

	// Create compute client
	computeClient, err := openstack.NewComputeClient(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create compute client: %v", err),
		})
	}

	// Get server
	ctx := context.Background()
	server, err := servers.Get(ctx, computeClient, id).Extract()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to get instance: %v", err),
		})
	}

	// Convert addresses
	addresses := make(map[string][]models.Address)
	for networkName, networkAddresses := range server.Addresses {
		addrs := []models.Address{}
		for _, addr := range networkAddresses.([]interface{}) {
			addrMap := addr.(map[string]interface{})
			address := models.Address{
				Type:    addrMap["OS-EXT-IPS:type"].(string),
				Address: addrMap["addr"].(string),
			}
			addrs = append(addrs, address)
		}
		addresses[networkName] = addrs
	}

	// Convert metadata to map[string]interface{} if needed
	metadata := make(map[string]interface{})
	for k, v := range server.Metadata {
		metadata[k] = v
	}

	// Convert to our model
	instance := models.Instance{
		ID:        server.ID,
		Name:      server.Name,
		Status:    server.Status,
		Flavor:    server.Flavor["id"].(string),
		Image:     server.Image["id"].(string),
		Addresses: addresses,
		Created:   server.Created,
		Metadata:  metadata,
	}

	// Return instance
	return c.JSON(instance)
}

// DeleteInstance deletes an instance by ID
func (h *ComputeHandler) DeleteInstance(c *fiber.Ctx) error {
	// Get provider from token
	provider, err := h.getProviderFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Authentication error: %v", err),
		})
	}

	// Get instance ID from URL
	instanceID := c.Params("id")
	if instanceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Instance ID is required",
		})
	}

	// Create compute client
	computeClient, err := openstack.NewComputeClient(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create compute client: %v", err),
		})
	}

	// Delete server
	ctx := context.Background()
	err = servers.Delete(ctx, computeClient, instanceID).ExtractErr()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to delete instance: %v", err),
		})
	}

	// Return success
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// UpdateInstance updates an instance by ID
func (h *ComputeHandler) UpdateInstance(c *fiber.Ctx) error {
	// Get provider from token
	provider, err := h.getProviderFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Authentication error: %v", err),
		})
	}

	// Get instance ID from URL
	instanceID := c.Params("id")
	if instanceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Instance ID is required",
		})
	}

	// Parse request body
	var req models.UpdateInstanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate that at least one field is provided
	if req.Name == nil && req.AccessIPv4 == nil && req.AccessIPv6 == nil && req.Hostname == nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "At least one field to update is required",
		})
	}

	// Create compute client
	computeClient, err := openstack.NewComputeClient(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create compute client: %v", err),
		})
	}

	// Create update options
	updateOpts := servers.UpdateOpts{
		Name:       req.Name,
		AccessIPv4: req.AccessIPv4,
		AccessIPv6: req.AccessIPv6,
		Hostname:   req.Hostname,
	}

	// Update server
	ctx := context.Background()
	server, err := servers.Update(ctx, computeClient, instanceID, updateOpts).Extract()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to update instance: %v", err),
		})
	}

	// Convert to our model
	addresses := make(map[string][]models.Address)
	for networkName, networkAddresses := range server.Addresses {
		addrs := []models.Address{}
		for _, addr := range networkAddresses.([]interface{}) {
			addrMap := addr.(map[string]interface{})
			address := models.Address{
				Type:    addrMap["OS-EXT-IPS:type"].(string),
				Address: addrMap["addr"].(string),
			}
			addrs = append(addrs, address)
		}
		addresses[networkName] = addrs
	}

	// Convert metadata to map[string]interface{} if needed
	metadata := make(map[string]interface{})
	for k, v := range server.Metadata {
		metadata[k] = v
	}

	instance := models.Instance{
		ID:        server.ID,
		Name:      server.Name,
		Status:    server.Status,
		Flavor:    server.Flavor["id"].(string),
		Image:     server.Image["id"].(string),
		Addresses: addresses,
		Created:   server.Created,
		Metadata:  metadata,
	}

	// Return updated instance
	return c.JSON(instance)
}

// PerformInstanceAction performs an action on an instance
func (h *ComputeHandler) PerformInstanceAction(c *fiber.Ctx) error {
	// Get provider from token
	provider, err := h.getProviderFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Authentication error: %v", err),
		})
	}

	// Get instance ID from URL
	instanceID := c.Params("id")
	if instanceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Instance ID is required",
		})
	}

	// Parse request body
	var req models.InstanceActionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate action
	if req.Action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Action is required",
		})
	}

	// Create compute client
	computeClient, err := openstack.NewComputeClient(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create compute client: %v", err),
		})
	}

	// Perform action
	ctx := context.Background()
	var actionErr error

	switch req.Action {
	case "start":
		actionErr = servers.Start(ctx, computeClient, instanceID).ExtractErr()
	case "stop":
		actionErr = servers.Stop(ctx, computeClient, instanceID).ExtractErr()
	case "reboot":
		rebootType := servers.SoftReboot
		if req.Type == "HARD" {
			rebootType = servers.HardReboot
		}
		actionErr = servers.Reboot(ctx, computeClient, instanceID, servers.RebootOpts{
			Type: rebootType,
		}).ExtractErr()
	default:
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Unsupported action: %s", req.Action),
		})
	}

	if actionErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to perform action %s: %v", req.Action, actionErr),
		})
	}

	// Return success
	return c.JSON(models.SuccessResponse{
		Message: fmt.Sprintf("Action %s performed successfully", req.Action),
	})
}

// ListFlavors lists all flavors in the project
func (h *ComputeHandler) ListFlavors(c *fiber.Ctx) error {
	// Get provider from token
	provider, err := h.getProviderFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Authentication error: %v", err),
		})
	}

	// Create compute client
	computeClient, err := openstack.NewComputeClient(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create compute client: %v", err),
		})
	}

	// List flavors
	ctx := context.Background()
	allPages, err := flavors.ListDetail(computeClient, flavors.ListOpts{}).AllPages(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to list flavors: %v", err),
		})
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to extract flavors: %v", err),
		})
	}

	// Convert to our model
	result := make([]models.Flavor, len(allFlavors))
	for i, flavor := range allFlavors {
		result[i] = models.Flavor{
			ID:    flavor.ID,
			Name:  flavor.Name,
			RAM:   flavor.RAM,
			VCPUs: flavor.VCPUs,
			Disk:  flavor.Disk,
		}
	}

	// Return flavors
	return c.JSON(result)
}
