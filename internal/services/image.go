package services

import (
	"context"

	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// ImageService handles operations related to image resources
type ImageService struct {
	Client *client.OpenStackClient
}

// NewImageService creates a new image service
func NewImageService(client *client.OpenStackClient) *ImageService {
	return &ImageService{
		Client: client,
	}
}

// ListImages lists all images
func (s *ImageService) ListImages() ([]models.Image, error) {
	var modelImages []models.Image
	ctx := context.Background()

	// Create a pager
	pager := images.List(s.Client.Image, images.ListOpts{})

	// Extract images from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		imageList, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		for _, image := range imageList {
			// Convert properties to string map
			properties := make(map[string]string)
			for k, v := range image.Properties {
				if strVal, ok := v.(string); ok {
					properties[k] = strVal
				}
			}

			modelImage := models.Image{
				ID:         image.ID,
				Name:       image.Name,
				Status:     string(image.Status),
				Size:       image.SizeBytes,
				Visibility: string(image.Visibility),
				Tags:       image.Tags,
				CreatedAt:  image.CreatedAt,
				UpdatedAt:  image.UpdatedAt,
				Properties: properties,
			}

			modelImages = append(modelImages, modelImage)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelImages, nil
}

// GetImage gets an image by ID
func (s *ImageService) GetImage(id string) (*models.Image, error) {
	ctx := context.Background()

	// Get the image
	image, err := images.Get(ctx, s.Client.Image, id).Extract()
	if err != nil {
		return nil, err
	}

	// Convert properties to string map
	properties := make(map[string]string)
	for k, v := range image.Properties {
		if strVal, ok := v.(string); ok {
			properties[k] = strVal
		}
	}

	// Return the image
	modelImage := &models.Image{
		ID:         image.ID,
		Name:       image.Name,
		Status:     string(image.Status),
		Size:       image.SizeBytes,
		Visibility: string(image.Visibility),
		Tags:       image.Tags,
		CreatedAt:  image.CreatedAt,
		UpdatedAt:  image.UpdatedAt,
		Properties: properties,
	}

	return modelImage, nil
}
