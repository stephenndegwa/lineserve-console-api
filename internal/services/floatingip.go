package services

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// FloatingIPService handles operations related to floating IP resources
type FloatingIPService struct {
	Client *client.OpenStackClient
}

// NewFloatingIPService creates a new floating IP service
func NewFloatingIPService(client *client.OpenStackClient) *FloatingIPService {
	return &FloatingIPService{
		Client: client,
	}
}

// ListFloatingIPs lists all floating IPs
func (s *FloatingIPService) ListFloatingIPs() ([]models.FloatingIP, error) {
	// Initialize with empty slice instead of nil
	modelFloatingIPs := []models.FloatingIP{}
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return modelFloatingIPs, fmt.Errorf("network client is nil")
	}

	// Create a pager
	listOpts := floatingips.ListOpts{}
	pager := floatingips.List(s.Client.Network, listOpts)

	// Extract floating IPs from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		// Extract floating IPs
		fips, err := floatingips.ExtractFloatingIPs(page)
		if err != nil {
			return false, err
		}

		// Convert to our model
		for _, fip := range fips {
			modelFIP := models.FloatingIP{
				ID:                fip.ID,
				FloatingIP:        fip.FloatingIP,
				FloatingNetworkID: fip.FloatingNetworkID,
				Status:            fip.Status,
				PortID:            fip.PortID,
				FixedIP:           fip.FixedIP,
				RouterID:          fip.RouterID,
				Description:       fip.Description,
				ProjectID:         fip.ProjectID,
				CreatedAt:         fip.CreatedAt,
				UpdatedAt:         fip.UpdatedAt,
			}
			modelFloatingIPs = append(modelFloatingIPs, modelFIP)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelFloatingIPs, nil
}

// GetFloatingIP gets a floating IP by ID
func (s *FloatingIPService) GetFloatingIP(id string) (*models.FloatingIP, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Get the floating IP
	fip, err := floatingips.Get(ctx, s.Client.Network, id).Extract()
	if err != nil {
		return nil, err
	}

	// Convert to our model
	modelFIP := &models.FloatingIP{
		ID:                fip.ID,
		FloatingIP:        fip.FloatingIP,
		FloatingNetworkID: fip.FloatingNetworkID,
		Status:            fip.Status,
		PortID:            fip.PortID,
		FixedIP:           fip.FixedIP,
		RouterID:          fip.RouterID,
		Description:       fip.Description,
		ProjectID:         fip.ProjectID,
		CreatedAt:         fip.CreatedAt,
		UpdatedAt:         fip.UpdatedAt,
	}

	return modelFIP, nil
}

// CreateFloatingIP creates a new floating IP
func (s *FloatingIPService) CreateFloatingIP(req models.CreateFloatingIPRequest) (*models.FloatingIP, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Define floating IP create options
	createOpts := floatingips.CreateOpts{
		FloatingNetworkID: req.FloatingNetworkID,
		Description:       req.Description,
	}

	// Create the floating IP
	fip, err := floatingips.Create(ctx, s.Client.Network, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	// Convert to our model
	modelFIP := &models.FloatingIP{
		ID:                fip.ID,
		FloatingIP:        fip.FloatingIP,
		FloatingNetworkID: fip.FloatingNetworkID,
		Status:            fip.Status,
		PortID:            fip.PortID,
		FixedIP:           fip.FixedIP,
		RouterID:          fip.RouterID,
		Description:       fip.Description,
		ProjectID:         fip.ProjectID,
		CreatedAt:         fip.CreatedAt,
		UpdatedAt:         fip.UpdatedAt,
	}

	return modelFIP, nil
}

// UpdateFloatingIP updates a floating IP
func (s *FloatingIPService) UpdateFloatingIP(id string, req models.UpdateFloatingIPRequest) (*models.FloatingIP, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Define floating IP update options
	updateOpts := floatingips.UpdateOpts{
		PortID:  req.PortID,
		FixedIP: req.FixedIP,
	}

	// Update the floating IP
	fip, err := floatingips.Update(ctx, s.Client.Network, id, updateOpts).Extract()
	if err != nil {
		return nil, err
	}

	// Convert to our model
	modelFIP := &models.FloatingIP{
		ID:                fip.ID,
		FloatingIP:        fip.FloatingIP,
		FloatingNetworkID: fip.FloatingNetworkID,
		Status:            fip.Status,
		PortID:            fip.PortID,
		FixedIP:           fip.FixedIP,
		RouterID:          fip.RouterID,
		Description:       fip.Description,
		ProjectID:         fip.ProjectID,
		CreatedAt:         fip.CreatedAt,
		UpdatedAt:         fip.UpdatedAt,
	}

	return modelFIP, nil
}

// DeleteFloatingIP deletes a floating IP
func (s *FloatingIPService) DeleteFloatingIP(id string) error {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return fmt.Errorf("network client is nil")
	}

	// Delete the floating IP
	return floatingips.Delete(ctx, s.Client.Network, id).ExtractErr()
}
