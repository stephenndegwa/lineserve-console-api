package services

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// SubnetService handles operations related to subnet resources
type SubnetService struct {
	Client *client.OpenStackClient
}

// NewSubnetService creates a new subnet service
func NewSubnetService(client *client.OpenStackClient) *SubnetService {
	return &SubnetService{
		Client: client,
	}
}

// ListSubnets lists all subnets
func (s *SubnetService) ListSubnets() ([]models.Subnet, error) {
	// Initialize with empty slice instead of nil
	modelSubnets := []models.Subnet{}
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return modelSubnets, fmt.Errorf("network client is nil")
	}

	// Create a pager
	listOpts := subnets.ListOpts{}
	pager := subnets.List(s.Client.Network, listOpts)

	// Extract subnets from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		// Extract subnets
		subnetList, err := subnets.ExtractSubnets(page)
		if err != nil {
			return false, err
		}

		// Convert to our model
		for _, subnet := range subnetList {
			// Convert allocation pools
			allocationPools := make([]models.AllocationPool, len(subnet.AllocationPools))
			for i, pool := range subnet.AllocationPools {
				allocationPools[i] = models.AllocationPool{
					Start: pool.Start,
					End:   pool.End,
				}
			}

			// Convert host routes
			hostRoutes := make([]models.HostRoute, len(subnet.HostRoutes))
			for i, route := range subnet.HostRoutes {
				hostRoutes[i] = models.HostRoute{
					DestinationCIDR: route.DestinationCIDR,
					NextHop:         route.NextHop,
				}
			}

			modelSubnet := models.Subnet{
				ID:              subnet.ID,
				Name:            subnet.Name,
				NetworkID:       subnet.NetworkID,
				CIDR:            subnet.CIDR,
				GatewayIP:       subnet.GatewayIP,
				IPVersion:       subnet.IPVersion,
				EnableDHCP:      subnet.EnableDHCP,
				DNSNameservers:  subnet.DNSNameservers,
				AllocationPools: allocationPools,
				HostRoutes:      hostRoutes,
				ServiceTypes:    subnet.ServiceTypes,
				ProjectID:       subnet.ProjectID,
				CreatedAt:       subnet.CreatedAt,
				UpdatedAt:       subnet.UpdatedAt,
			}
			modelSubnets = append(modelSubnets, modelSubnet)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelSubnets, nil
}

// GetSubnet gets a subnet by ID
func (s *SubnetService) GetSubnet(id string) (*models.Subnet, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Get the subnet
	subnet, err := subnets.Get(ctx, s.Client.Network, id).Extract()
	if err != nil {
		return nil, err
	}

	// Convert allocation pools
	allocationPools := make([]models.AllocationPool, len(subnet.AllocationPools))
	for i, pool := range subnet.AllocationPools {
		allocationPools[i] = models.AllocationPool{
			Start: pool.Start,
			End:   pool.End,
		}
	}

	// Convert host routes
	hostRoutes := make([]models.HostRoute, len(subnet.HostRoutes))
	for i, route := range subnet.HostRoutes {
		hostRoutes[i] = models.HostRoute{
			DestinationCIDR: route.DestinationCIDR,
			NextHop:         route.NextHop,
		}
	}

	// Convert to our model
	modelSubnet := &models.Subnet{
		ID:              subnet.ID,
		Name:            subnet.Name,
		NetworkID:       subnet.NetworkID,
		CIDR:            subnet.CIDR,
		GatewayIP:       subnet.GatewayIP,
		IPVersion:       subnet.IPVersion,
		EnableDHCP:      subnet.EnableDHCP,
		DNSNameservers:  subnet.DNSNameservers,
		AllocationPools: allocationPools,
		HostRoutes:      hostRoutes,
		ServiceTypes:    subnet.ServiceTypes,
		ProjectID:       subnet.ProjectID,
		CreatedAt:       subnet.CreatedAt,
		UpdatedAt:       subnet.UpdatedAt,
	}

	return modelSubnet, nil
}

// CreateSubnet creates a new subnet
func (s *SubnetService) CreateSubnet(req models.CreateSubnetRequest) (*models.Subnet, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Convert allocation pools
	allocationPools := make([]subnets.AllocationPool, len(req.AllocationPools))
	for i, pool := range req.AllocationPools {
		allocationPools[i] = subnets.AllocationPool{
			Start: pool.Start,
			End:   pool.End,
		}
	}

	// Convert host routes
	hostRoutes := make([]subnets.HostRoute, len(req.HostRoutes))
	for i, route := range req.HostRoutes {
		hostRoutes[i] = subnets.HostRoute{
			DestinationCIDR: route.DestinationCIDR,
			NextHop:         route.NextHop,
		}
	}

	// Define subnet create options
	createOpts := subnets.CreateOpts{
		NetworkID:       req.NetworkID,
		Name:            req.Name,
		CIDR:            req.CIDR,
		IPVersion:       req.IPVersion,
		GatewayIP:       req.GatewayIP,
		EnableDHCP:      req.EnableDHCP,
		DNSNameservers:  req.DNSNameservers,
		AllocationPools: allocationPools,
		HostRoutes:      hostRoutes,
		ServiceTypes:    req.ServiceTypes,
	}

	// Create the subnet
	subnet, err := subnets.Create(ctx, s.Client.Network, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	// Convert allocation pools for response
	respAllocationPools := make([]models.AllocationPool, len(subnet.AllocationPools))
	for i, pool := range subnet.AllocationPools {
		respAllocationPools[i] = models.AllocationPool{
			Start: pool.Start,
			End:   pool.End,
		}
	}

	// Convert host routes for response
	respHostRoutes := make([]models.HostRoute, len(subnet.HostRoutes))
	for i, route := range subnet.HostRoutes {
		respHostRoutes[i] = models.HostRoute{
			DestinationCIDR: route.DestinationCIDR,
			NextHop:         route.NextHop,
		}
	}

	// Convert to our model
	modelSubnet := &models.Subnet{
		ID:              subnet.ID,
		Name:            subnet.Name,
		NetworkID:       subnet.NetworkID,
		CIDR:            subnet.CIDR,
		GatewayIP:       subnet.GatewayIP,
		IPVersion:       subnet.IPVersion,
		EnableDHCP:      subnet.EnableDHCP,
		DNSNameservers:  subnet.DNSNameservers,
		AllocationPools: respAllocationPools,
		HostRoutes:      respHostRoutes,
		ServiceTypes:    subnet.ServiceTypes,
		ProjectID:       subnet.ProjectID,
		CreatedAt:       subnet.CreatedAt,
		UpdatedAt:       subnet.UpdatedAt,
	}

	return modelSubnet, nil
}

// DeleteSubnet deletes a subnet
func (s *SubnetService) DeleteSubnet(id string) error {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return fmt.Errorf("network client is nil")
	}

	// Delete the subnet
	return subnets.Delete(ctx, s.Client.Network, id).ExtractErr()
}
