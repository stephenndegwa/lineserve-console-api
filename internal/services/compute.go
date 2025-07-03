package services

import (
	"context"

	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// ComputeService handles operations related to compute resources
type ComputeService struct {
	Client *client.OpenStackClient
}

// NewComputeService creates a new compute service
func NewComputeService(client *client.OpenStackClient) *ComputeService {
	return &ComputeService{
		Client: client,
	}
}

// ListInstances lists all instances
func (s *ComputeService) ListInstances() ([]models.Instance, error) {
	var modelInstances []models.Instance
	ctx := context.Background()

	// Create a pager
	pager := servers.List(s.Client.Compute, servers.ListOpts{})

	// Extract instances from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}

		for _, server := range serverList {
			// Convert addresses safely
			addresses := make(map[string][]models.Address)
			if server.Addresses != nil {
				for network, addrList := range server.Addresses {
					var modelAddresses []models.Address
					if addrList != nil {
						addrArray, ok := addrList.([]interface{})
						if ok {
							for _, addr := range addrArray {
								if addr != nil {
									addrMap, ok := addr.(map[string]interface{})
									if ok {
										// Safely extract type and address
										addrType := "fixed" // default
										if typeVal, ok := addrMap["OS-EXT-IPS:type"]; ok && typeVal != nil {
											if typeStr, ok := typeVal.(string); ok {
												addrType = typeStr
											}
										}

										addrValue := ""
										if addrVal, ok := addrMap["addr"]; ok && addrVal != nil {
											if addrStr, ok := addrVal.(string); ok {
												addrValue = addrStr
											}
										}

										if addrValue != "" {
											modelAddr := models.Address{
												Type:    addrType,
												Address: addrValue,
											}
											modelAddresses = append(modelAddresses, modelAddr)
										}
									}
								}
							}
						}
					}
					if len(modelAddresses) > 0 {
						addresses[network] = modelAddresses
					}
				}
			}

			// Handle potentially nil flavor and image
			var flavorID, imageID string
			if server.Flavor != nil {
				if id, ok := server.Flavor["id"]; ok && id != nil {
					flavorID = id.(string)
				}
			}
			if server.Image != nil {
				if id, ok := server.Image["id"]; ok && id != nil {
					imageID = id.(string)
				}
			}

			modelInstance := models.Instance{
				ID:        server.ID,
				Name:      server.Name,
				Status:    server.Status,
				Created:   server.Created,
				Flavor:    flavorID,
				Image:     imageID,
				Addresses: addresses,
				Metadata:  convertMetadata(server.Metadata),
			}

			modelInstances = append(modelInstances, modelInstance)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelInstances, nil
}

// CreateInstance creates a new instance
func (s *ComputeService) CreateInstance(req models.CreateInstanceRequest) (*models.Instance, error) {
	ctx := context.Background()

	// Define server create options
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

	// Create the server
	server, err := servers.Create(ctx, s.Client.Compute, createOpts, nil).Extract()
	if err != nil {
		return nil, err
	}

	// Return the instance
	modelInstance := &models.Instance{
		ID:        server.ID,
		Name:      server.Name,
		Status:    server.Status,
		Created:   server.Created,
		Flavor:    req.FlavorID,
		Image:     req.ImageID,
		Addresses: make(map[string][]models.Address),
		Metadata:  make(map[string]interface{}),
	}

	return modelInstance, nil
}

// GetInstance gets an instance by ID
func (s *ComputeService) GetInstance(id string) (*models.Instance, error) {
	ctx := context.Background()

	// Get the server
	server, err := servers.Get(ctx, s.Client.Compute, id).Extract()
	if err != nil {
		return nil, err
	}

	// Convert addresses safely
	addresses := make(map[string][]models.Address)
	if server.Addresses != nil {
		for network, addrList := range server.Addresses {
			var modelAddresses []models.Address
			if addrList != nil {
				addrArray, ok := addrList.([]interface{})
				if ok {
					for _, addr := range addrArray {
						if addr != nil {
							addrMap, ok := addr.(map[string]interface{})
							if ok {
								// Safely extract type and address
								addrType := "fixed" // default
								if typeVal, ok := addrMap["OS-EXT-IPS:type"]; ok && typeVal != nil {
									if typeStr, ok := typeVal.(string); ok {
										addrType = typeStr
									}
								}

								addrValue := ""
								if addrVal, ok := addrMap["addr"]; ok && addrVal != nil {
									if addrStr, ok := addrVal.(string); ok {
										addrValue = addrStr
									}
								}

								if addrValue != "" {
									modelAddr := models.Address{
										Type:    addrType,
										Address: addrValue,
									}
									modelAddresses = append(modelAddresses, modelAddr)
								}
							}
						}
					}
				}
			}
			if len(modelAddresses) > 0 {
				addresses[network] = modelAddresses
			}
		}
	}

	// Handle potentially nil flavor and image
	var flavorID, imageID string
	if server.Flavor != nil {
		if id, ok := server.Flavor["id"]; ok && id != nil {
			flavorID = id.(string)
		}
	}
	if server.Image != nil {
		if id, ok := server.Image["id"]; ok && id != nil {
			imageID = id.(string)
		}
	}

	// Return the instance
	modelInstance := &models.Instance{
		ID:        server.ID,
		Name:      server.Name,
		Status:    server.Status,
		Created:   server.Created,
		Flavor:    flavorID,
		Image:     imageID,
		Addresses: addresses,
		Metadata:  convertMetadata(server.Metadata),
	}

	return modelInstance, nil
}

// ListFlavors lists all flavors
func (s *ComputeService) ListFlavors() ([]models.Flavor, error) {
	var modelFlavors []models.Flavor
	ctx := context.Background()

	// Create a pager
	pager := flavors.ListDetail(s.Client.Compute, flavors.ListOpts{})

	// Extract flavors from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}

		for _, flavor := range flavorList {
			modelFlavor := models.Flavor{
				ID:    flavor.ID,
				Name:  flavor.Name,
				RAM:   flavor.RAM,
				VCPUs: flavor.VCPUs,
				Disk:  flavor.Disk,
			}

			modelFlavors = append(modelFlavors, modelFlavor)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelFlavors, nil
}

// Helper function to convert metadata
func convertMetadata(metadata map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	if metadata == nil {
		return result
	}

	for k, v := range metadata {
		result[k] = v
	}
	return result
}
