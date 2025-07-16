package services

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumetypes"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/volumeattach"
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

	// Check if Volume client is nil
	if s.Client == nil || s.Client.Volume == nil {
		return modelVolumes, fmt.Errorf("volume client is nil")
	}

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
					ID:         attachment.AttachmentID,
					VolumeID:   volume.ID,
					InstanceID: attachment.ServerID,
					Device:     attachment.Device,
					Status:     "attached", // OpenStack doesn't provide status in attachment, so we assume it's attached
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

	// Check if Volume client is nil
	if s.Client == nil || s.Client.Volume == nil {
		return nil, fmt.Errorf("volume client is nil")
	}

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

	// Check if Volume client is nil
	if s.Client == nil || s.Client.Volume == nil {
		return nil, fmt.Errorf("volume client is nil")
	}

	// Get the volume
	volume, err := volumes.Get(ctx, s.Client.Volume, id).Extract()
	if err != nil {
		return nil, err
	}

	// Convert attachments
	var modelAttachments []models.VolumeAttachment
	for _, attachment := range volume.Attachments {
		modelAttachment := models.VolumeAttachment{
			ID:         attachment.AttachmentID,
			VolumeID:   volume.ID,
			InstanceID: attachment.ServerID,
			Device:     attachment.Device,
			Status:     "attached", // OpenStack doesn't provide status in attachment, so we assume it's attached
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

// DeleteVolume deletes a volume by ID
func (s *VolumeService) DeleteVolume(id string) error {
	ctx := context.Background()

	// Check if Volume client is nil
	if s.Client == nil || s.Client.Volume == nil {
		return fmt.Errorf("volume client is nil")
	}

	// Delete the volume
	return volumes.Delete(ctx, s.Client.Volume, id, volumes.DeleteOpts{}).ExtractErr()
}

// ListVolumeTypes lists all volume types
func (s *VolumeService) ListVolumeTypes() ([]models.VolumeType, error) {
	var modelVolumeTypes []models.VolumeType
	ctx := context.Background()

	// Check if Volume client is nil
	if s.Client == nil || s.Client.Volume == nil {
		return modelVolumeTypes, fmt.Errorf("volume client is nil")
	}

	// Create a pager
	pager := volumetypes.List(s.Client.Volume, volumetypes.ListOpts{})

	// Extract volume types from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		volumeTypeList, err := volumetypes.ExtractVolumeTypes(page)
		if err != nil {
			return false, err
		}

		for _, volumeType := range volumeTypeList {
			modelVolumeType := models.VolumeType{
				ID:          volumeType.ID,
				Name:        volumeType.Name,
				Description: volumeType.Description,
				IsPublic:    volumeType.IsPublic,
			}

			modelVolumeTypes = append(modelVolumeTypes, modelVolumeType)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelVolumeTypes, nil
}

// AttachVolume attaches a volume to an instance
func (s *VolumeService) AttachVolume(volumeID string, req models.VolumeAttachRequest) (*models.VolumeAttachment, error) {
	ctx := context.Background()

	// Check if Compute client is nil
	if s.Client == nil || s.Client.Compute == nil {
		return nil, fmt.Errorf("compute client is nil")
	}

	// Define attach options
	attachOpts := volumeattach.CreateOpts{
		VolumeID: volumeID,
		Device:   req.Device,
	}

	// Attach the volume
	attachment, err := volumeattach.Create(ctx, s.Client.Compute, req.InstanceID, attachOpts).Extract()
	if err != nil {
		return nil, err
	}

	// Return the attachment
	modelAttachment := &models.VolumeAttachment{
		ID:         attachment.ID,
		VolumeID:   attachment.VolumeID,
		InstanceID: attachment.ServerID,
		Device:     attachment.Device,
		Status:     "attaching", // Initial status is attaching
	}

	return modelAttachment, nil
}

// DetachVolume detaches a volume from an instance
func (s *VolumeService) DetachVolume(volumeID string, req models.VolumeDetachRequest) error {
	ctx := context.Background()

	// Check if Compute client is nil
	if s.Client == nil || s.Client.Compute == nil {
		return fmt.Errorf("compute client is nil")
	}

	// First, get the volume to find the server ID
	volume, err := s.GetVolume(volumeID)
	if err != nil {
		return err
	}

	// Find the attachment
	var serverID, attachmentID string
	if req.AttachmentID != "" {
		// If attachment ID is provided, find it in the attachments
		for _, attachment := range volume.Attachments {
			if attachment.ID == req.AttachmentID {
				serverID = attachment.InstanceID
				attachmentID = attachment.ID
				break
			}
		}
	} else if len(volume.Attachments) > 0 {
		// If no attachment ID is provided, use the first attachment
		serverID = volume.Attachments[0].InstanceID
		attachmentID = volume.Attachments[0].ID
	}

	if serverID == "" {
		return fmt.Errorf("volume is not attached to any instance")
	}

	// Detach the volume
	return volumeattach.Delete(ctx, s.Client.Compute, serverID, attachmentID).ExtractErr()
}

// ResizeVolume resizes a volume to a new size
func (s *VolumeService) ResizeVolume(volumeID string, req models.VolumeResizeRequest) error {
	ctx := context.Background()

	// Check if Volume client is nil
	if s.Client == nil || s.Client.Volume == nil {
		return fmt.Errorf("volume client is nil")
	}

	// Define resize options
	resizeOpts := volumes.ExtendSizeOpts{
		NewSize: req.NewSize,
	}

	// Resize the volume
	return volumes.ExtendSize(ctx, s.Client.Volume, volumeID, resizeOpts).ExtractErr()
}
