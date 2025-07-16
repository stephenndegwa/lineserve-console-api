package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// CreateInstance creates a new compute instance in OpenStack
func CreateInstance(client *client.OpenStackClient, projectID, name, flavorID, imageID, networkID string) (*models.Instance, error) {
	if client == nil {
		return nil, fmt.Errorf("OpenStack client is nil")
	}

	// Create instance options
	createOpts := servers.CreateOpts{
		Name:      name,
		FlavorRef: flavorID,
		ImageRef:  imageID,
		Networks: []servers.Network{
			{
				UUID: networkID,
			},
		},
	}

	// Create the instance
	ctx := context.Background()
	server, err := servers.Create(ctx, client.Compute, createOpts, nil).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %v", err)
	}

	// Convert to our model
	instance := &models.Instance{
		ID:      server.ID,
		Name:    server.Name,
		Status:  server.Status,
		Flavor:  flavorID,
		Image:   imageID,
		Created: server.Created,
	}

	return instance, nil
}
