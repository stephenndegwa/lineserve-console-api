package services

import (
	"context"

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
			isExternal := false

			// Look for router:external in the network properties
			if network.AdminStateUp {
				isExternal = true
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
	network, err := networks.Get(ctx, s.Client.Network, id).Extract()
	if err != nil {
		return nil, err
	}

	// Check if network has router:external attribute
	isExternal := false

	// Look for router:external in the network properties
	if network.AdminStateUp {
		isExternal = true
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
