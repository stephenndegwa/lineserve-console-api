package services

import (
	"context"
	"fmt"

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
	// Initialize with empty slice instead of nil
	modelNetworks := []models.Network{}
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		fmt.Println("ERROR: Network client is nil")
		return modelNetworks, fmt.Errorf("network client is nil")
	}

	// Create a pager
	listOpts := networks.ListOpts{}
	pager := networks.List(s.Client.Network, listOpts)

	// Extract networks from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		// Extract basic network info
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			fmt.Printf("ERROR extracting networks: %v\n", err)
			return false, err
		}

		fmt.Printf("Found %d networks in page\n", len(networkList))

		// Extract networks with external extension info
		var networkWithExtList []struct {
			networks.Network
			external.NetworkExternalExt
		}

		err = networks.ExtractNetworksInto(page, &networkWithExtList)
		if err != nil {
			fmt.Printf("ERROR extracting external network info: %v\n", err)
			// If we can't extract external info, continue with basic info
			for _, network := range networkList {
				modelNetwork := models.Network{
					ID:       network.ID,
					Name:     network.Name,
					Status:   network.Status,
					Shared:   network.Shared,
					External: false, // Default to false if we can't determine
				}
				modelNetworks = append(modelNetworks, modelNetwork)
			}
		} else {
			fmt.Printf("Successfully extracted %d networks with external info\n", len(networkWithExtList))
			// Use the extracted external info
			for _, network := range networkWithExtList {
				modelNetwork := models.Network{
					ID:       network.ID,
					Name:     network.Name,
					Status:   network.Status,
					Shared:   network.Shared,
					External: network.External,
				}
				modelNetworks = append(modelNetworks, modelNetwork)
			}
		}

		return true, nil
	})

	if err != nil {
		fmt.Printf("ERROR in pager.EachPage: %v\n", err)
		return nil, err
	}

	fmt.Printf("Returning %d networks\n", len(modelNetworks))
	return modelNetworks, nil
}

// GetNetwork gets a network by ID
func (s *NetworkService) GetNetwork(id string) (*models.Network, error) {
	ctx := context.Background()

	// Get the network
	r := networks.Get(ctx, s.Client.Network, id)

	// Create a struct to hold both network and external extension info
	var networkWithExt struct {
		networks.Network
		external.NetworkExternalExt
	}

	// Extract into the combined struct
	err := r.ExtractInto(&networkWithExt)
	if err != nil {
		return nil, err
	}

	// Return the network
	modelNetwork := &models.Network{
		ID:       networkWithExt.ID,
		Name:     networkWithExt.Name,
		Status:   networkWithExt.Status,
		Shared:   networkWithExt.Shared,
		External: networkWithExt.External,
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
		// Convert bool to *bool for the external extension
		externalBool := req.External
		createOptsBuilder = external.CreateOptsExt{
			CreateOptsBuilder: createOptsBuilder,
			External:          &externalBool,
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
