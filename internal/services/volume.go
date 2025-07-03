package services

import (
	"context"

	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// VolumeService handles operations related to volume resources
type VolumeService struct {
	Client *client.OpenStackClient
}

// NewVolumeService creates a new volume service
func NewVolumeService(client *client.OpenStackClient) *VolumeService {
	return &VolumeService{
		Client: client,
	}
}

// ListVolumes lists all volumes
func (s *VolumeService) ListVolumes() ([]models.Volume, error) {
	var modelVolumes []models.Volume
	ctx := context.Background()

	// Create a pager
	pager := volumes.List(s.Client.Volume, volumes.ListOpts{})

	// Extract volumes from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		volumeList, err := volumes.ExtractVolumes(page)
		if err != nil {
			return false, err
		}

		for _, volume := range volumeList {
			// Convert attachments
			var modelAttachments []models.VolumeAttachment
			for _, attachment := range volume.Attachments {
				modelAttachment := models.VolumeAttachment{
					ServerID:     attachment.ServerID,
					AttachmentID: attachment.AttachmentID,
					DeviceName:   attachment.Device,
				}
				modelAttachments = append(modelAttachments, modelAttachment)
			}

			modelVolume := models.Volume{
				ID:               volume.ID,
				Name:             volume.Name,
				Status:           volume.Status,
				Size:             volume.Size,
				VolumeType:       volume.VolumeType,
				AvailabilityZone: volume.AvailabilityZone,
				CreatedAt:        volume.CreatedAt,
				Attachments:      modelAttachments,
			}

			modelVolumes = append(modelVolumes, modelVolume)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelVolumes, nil
}

// CreateVolume creates a new volume
func (s *VolumeService) CreateVolume(req models.CreateVolumeRequest) (*models.Volume, error) {
	ctx := context.Background()

	// Define volume create options
	createOpts := volumes.CreateOpts{
		Name:             req.Name,
		Size:             req.Size,
		VolumeType:       req.VolumeType,
		AvailabilityZone: req.AvailabilityZone,
	}

	// Create the volume
	volume, err := volumes.Create(ctx, s.Client.Volume, createOpts, nil).Extract()
	if err != nil {
		return nil, err
	}

	// Return the volume
	modelVolume := &models.Volume{
		ID:               volume.ID,
		Name:             volume.Name,
		Status:           volume.Status,
		Size:             volume.Size,
		VolumeType:       volume.VolumeType,
		AvailabilityZone: volume.AvailabilityZone,
		CreatedAt:        volume.CreatedAt,
		Attachments:      []models.VolumeAttachment{},
	}

	return modelVolume, nil
}

// GetVolume gets a volume by ID
func (s *VolumeService) GetVolume(id string) (*models.Volume, error) {
	ctx := context.Background()

	// Get the volume
	volume, err := volumes.Get(ctx, s.Client.Volume, id).Extract()
	if err != nil {
		return nil, err
	}

	// Convert attachments
	var modelAttachments []models.VolumeAttachment
	for _, attachment := range volume.Attachments {
		modelAttachment := models.VolumeAttachment{
			ServerID:     attachment.ServerID,
			AttachmentID: attachment.AttachmentID,
			DeviceName:   attachment.Device,
		}
		modelAttachments = append(modelAttachments, modelAttachment)
	}

	// Return the volume
	modelVolume := &models.Volume{
		ID:               volume.ID,
		Name:             volume.Name,
		Status:           volume.Status,
		Size:             volume.Size,
		VolumeType:       volume.VolumeType,
		AvailabilityZone: volume.AvailabilityZone,
		CreatedAt:        volume.CreatedAt,
		Attachments:      modelAttachments,
	}

	return modelVolume, nil
}
