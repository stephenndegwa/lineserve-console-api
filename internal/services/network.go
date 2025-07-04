package services

import (
	"context"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// NetworkService handles operations related to network resources
type NetworkService struct {
	Client *client.OpenStackClient
}

// NewNetworkService creates a new network service
func NewNetworkService(client *client.OpenStackClient) *NetworkService {
	return &NetworkService{
		Client: client,
	}
}

// ListNetworks lists all networks
func (s *NetworkService) ListNetworks() ([]models.Network, error) {
	var modelNetworks []models.Network
	ctx := context.Background()

	// Create a pager
	pager := networks.List(s.Client.Network, networks.ListOpts{})

	// Extract networks from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}

		for _, network := range networkList {
			// Check if network has router:external attribute
			var externalNetwork external.NetworkExternalExt
			err := networks.ExtractInto(page, &externalNetwork)

			// Default to false if we can't determine
			isExternal := false
			if err == nil {
				isExternal = externalNetwork.External
			}

			modelNetwork := models.Network{
				ID:       network.ID,
				Name:     network.Name,
				Status:   network.Status,
				Shared:   network.Shared,
				External: isExternal,
			}

			modelNetworks = append(modelNetworks, modelNetwork)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelNetworks, nil
}

// GetNetwork gets a network by ID
func (s *NetworkService) GetNetwork(id string) (*models.Network, error) {
	ctx := context.Background()

	// Get the network
	r := networks.Get(ctx, s.Client.Network, id)

	// Extract the basic network information
	network, err := r.Extract()
	if err != nil {
		return nil, err
	}

	// Try to extract the external attribute
	var externalNetwork external.NetworkExternalExt
	err = r.ExtractInto(&externalNetwork)

	// Default to false if we can't determine
	isExternal := false
	if err == nil {
		isExternal = externalNetwork.External
	}

	// Return the network
	modelNetwork := &models.Network{
		ID:       network.ID,
		Name:     network.Name,
		Status:   network.Status,
		Shared:   network.Shared,
		External: isExternal,
	}

	return modelNetwork, nil
}

// CreateNetwork creates a new network
func (s *NetworkService) CreateNetwork(req models.CreateNetworkRequest) (*models.Network, error) {
	ctx := context.Background()

	// Define network create options
	createOpts := networks.CreateOpts{
		Name:         req.Name,
		AdminStateUp: &req.AdminStateUp,
		Shared:       &req.Shared,
	}

	// Add external network extension if requested
	var createOptsBuilder networks.CreateOptsBuilder = createOpts
	if req.External {
		createOptsBuilder = external.CreateOptsExt{
			CreateOptsBuilder: createOpts,
			External:          true,
		}
	}

	// Create the network
	network, err := networks.Create(ctx, s.Client.Network, createOptsBuilder).Extract()
	if err != nil {
		return nil, err
	}

	// Return the network
	modelNetwork := &models.Network{
		ID:       network.ID,
		Name:     network.Name,
		Status:   network.Status,
		Shared:   network.Shared,
		External: req.External, // Use the requested value since it may not be immediately reflected
	}

	return modelNetwork, nil
}

// DeleteNetwork deletes a network by ID
func (s *NetworkService) DeleteNetwork(id string) error {
	ctx := context.Background()

	// Delete the network
	return networks.Delete(ctx, s.Client.Network, id).ExtractErr()
}
