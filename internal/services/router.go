package services

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// RouterService handles operations related to router resources
type RouterService struct {
	Client *client.OpenStackClient
}

// NewRouterService creates a new router service
func NewRouterService(client *client.OpenStackClient) *RouterService {
	return &RouterService{
		Client: client,
	}
}

// ListRouters lists all routers
func (s *RouterService) ListRouters() ([]models.Router, error) {
	// Initialize with empty slice instead of nil
	modelRouters := []models.Router{}
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return modelRouters, fmt.Errorf("network client is nil")
	}

	// Create a pager
	listOpts := routers.ListOpts{}
	pager := routers.List(s.Client.Network, listOpts)

	// Extract routers from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		// Extract routers
		routerList, err := routers.ExtractRouters(page)
		if err != nil {
			return false, err
		}

		// Convert to our model
		for _, router := range routerList {
			// Convert gateway info if it exists
			var gatewayInfo *models.GatewayInfo
			if router.GatewayInfo.NetworkID != "" {
				// Convert external fixed IPs
				externalFixedIPs := make([]models.ExternalFixedIP, len(router.GatewayInfo.ExternalFixedIPs))
				for i, fixedIP := range router.GatewayInfo.ExternalFixedIPs {
					externalFixedIPs[i] = models.ExternalFixedIP{
						SubnetID:  fixedIP.SubnetID,
						IPAddress: fixedIP.IPAddress,
					}
				}

				gatewayInfo = &models.GatewayInfo{
					NetworkID:        router.GatewayInfo.NetworkID,
					EnableSNAT:       router.GatewayInfo.EnableSNAT,
					ExternalFixedIPs: externalFixedIPs,
				}
			}

			// Convert routes
			routes := make([]models.Route, len(router.Routes))
			for i, route := range router.Routes {
				routes[i] = models.Route{
					DestinationCIDR: route.DestinationCIDR,
					NextHop:         route.NextHop,
				}
			}

			modelRouter := models.Router{
				ID:           router.ID,
				Name:         router.Name,
				Status:       router.Status,
				AdminStateUp: router.AdminStateUp,
				GatewayInfo:  gatewayInfo,
				Routes:       routes,
				ProjectID:    router.ProjectID,
				CreatedAt:    router.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:    router.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			modelRouters = append(modelRouters, modelRouter)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelRouters, nil
}

// GetRouter gets a router by ID
func (s *RouterService) GetRouter(id string) (*models.Router, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Get the router
	router, err := routers.Get(ctx, s.Client.Network, id).Extract()
	if err != nil {
		return nil, err
	}

	// Convert gateway info if it exists
	var gatewayInfo *models.GatewayInfo
	if router.GatewayInfo.NetworkID != "" {
		// Convert external fixed IPs
		externalFixedIPs := make([]models.ExternalFixedIP, len(router.GatewayInfo.ExternalFixedIPs))
		for i, fixedIP := range router.GatewayInfo.ExternalFixedIPs {
			externalFixedIPs[i] = models.ExternalFixedIP{
				SubnetID:  fixedIP.SubnetID,
				IPAddress: fixedIP.IPAddress,
			}
		}

		gatewayInfo = &models.GatewayInfo{
			NetworkID:        router.GatewayInfo.NetworkID,
			EnableSNAT:       router.GatewayInfo.EnableSNAT,
			ExternalFixedIPs: externalFixedIPs,
		}
	}

	// Convert routes
	routes := make([]models.Route, len(router.Routes))
	for i, route := range router.Routes {
		routes[i] = models.Route{
			DestinationCIDR: route.DestinationCIDR,
			NextHop:         route.NextHop,
		}
	}

	// Convert to our model
	modelRouter := &models.Router{
		ID:           router.ID,
		Name:         router.Name,
		Status:       router.Status,
		AdminStateUp: router.AdminStateUp,
		GatewayInfo:  gatewayInfo,
		Routes:       routes,
		ProjectID:    router.ProjectID,
		CreatedAt:    router.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    router.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return modelRouter, nil
}

// CreateRouter creates a new router
func (s *RouterService) CreateRouter(req models.CreateRouterRequest) (*models.Router, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Convert gateway info if it exists
	var gatewayInfo *routers.GatewayInfo
	if req.GatewayInfo != nil {
		// Convert external fixed IPs
		externalFixedIPs := make([]routers.ExternalFixedIP, len(req.GatewayInfo.ExternalFixedIPs))
		for i, fixedIP := range req.GatewayInfo.ExternalFixedIPs {
			externalFixedIPs[i] = routers.ExternalFixedIP{
				SubnetID:  fixedIP.SubnetID,
				IPAddress: fixedIP.IPAddress,
			}
		}

		gatewayInfo = &routers.GatewayInfo{
			NetworkID:        req.GatewayInfo.NetworkID,
			EnableSNAT:       req.GatewayInfo.EnableSNAT,
			ExternalFixedIPs: externalFixedIPs,
		}
	}

	// Define router create options (Routes field is not supported in CreateOpts)
	createOpts := routers.CreateOpts{
		Name:         req.Name,
		AdminStateUp: req.AdminStateUp,
		GatewayInfo:  gatewayInfo,
	}

	// Create the router
	router, err := routers.Create(ctx, s.Client.Network, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	// Convert gateway info for response if it exists
	var respGatewayInfo *models.GatewayInfo
	if router.GatewayInfo.NetworkID != "" {
		// Convert external fixed IPs for response
		externalFixedIPs := make([]models.ExternalFixedIP, len(router.GatewayInfo.ExternalFixedIPs))
		for i, fixedIP := range router.GatewayInfo.ExternalFixedIPs {
			externalFixedIPs[i] = models.ExternalFixedIP{
				SubnetID:  fixedIP.SubnetID,
				IPAddress: fixedIP.IPAddress,
			}
		}

		respGatewayInfo = &models.GatewayInfo{
			NetworkID:        router.GatewayInfo.NetworkID,
			EnableSNAT:       router.GatewayInfo.EnableSNAT,
			ExternalFixedIPs: externalFixedIPs,
		}
	}

	// Convert routes for response
	respRoutes := make([]models.Route, len(router.Routes))
	for i, route := range router.Routes {
		respRoutes[i] = models.Route{
			DestinationCIDR: route.DestinationCIDR,
			NextHop:         route.NextHop,
		}
	}

	// Convert to our model
	modelRouter := &models.Router{
		ID:           router.ID,
		Name:         router.Name,
		Status:       router.Status,
		AdminStateUp: router.AdminStateUp,
		GatewayInfo:  respGatewayInfo,
		Routes:       respRoutes,
		ProjectID:    router.ProjectID,
		CreatedAt:    router.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    router.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return modelRouter, nil
}

// DeleteRouter deletes a router
func (s *RouterService) DeleteRouter(id string) error {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return fmt.Errorf("network client is nil")
	}

	// Delete the router
	return routers.Delete(ctx, s.Client.Network, id).ExtractErr()
}

// AddRouterInterface adds an interface to a router
func (s *RouterService) AddRouterInterface(id string, req models.RouterInterfaceRequest) (*models.RouterInterface, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Define add interface options
	addOpts := routers.AddInterfaceOpts{
		SubnetID: req.SubnetID,
		PortID:   req.PortID,
	}

	// Add the interface
	routerInterface, err := routers.AddInterface(ctx, s.Client.Network, id, addOpts).Extract()
	if err != nil {
		return nil, err
	}

	// Convert to our model
	modelRouterInterface := &models.RouterInterface{
		ID:       routerInterface.ID,
		SubnetID: routerInterface.SubnetID,
		PortID:   routerInterface.PortID,
		TenantID: routerInterface.TenantID,
	}

	return modelRouterInterface, nil
}

// RemoveRouterInterface removes an interface from a router
func (s *RouterService) RemoveRouterInterface(id string, req models.RouterInterfaceRequest) (*models.RouterInterface, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Define remove interface options
	removeOpts := routers.RemoveInterfaceOpts{
		SubnetID: req.SubnetID,
		PortID:   req.PortID,
	}

	// Remove the interface
	routerInterface, err := routers.RemoveInterface(ctx, s.Client.Network, id, removeOpts).Extract()
	if err != nil {
		return nil, err
	}

	// Convert to our model
	modelRouterInterface := &models.RouterInterface{
		ID:       routerInterface.ID,
		SubnetID: routerInterface.SubnetID,
		PortID:   routerInterface.PortID,
		TenantID: routerInterface.TenantID,
	}

	return modelRouterInterface, nil
}
