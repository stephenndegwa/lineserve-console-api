package services

import (
	"context"

	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/keypairs"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// KeyPairService handles operations related to key pair resources
type KeyPairService struct {
	Client *client.OpenStackClient
}

// NewKeyPairService creates a new key pair service
func NewKeyPairService(client *client.OpenStackClient) *KeyPairService {
	return &KeyPairService{
		Client: client,
	}
}

// ListKeyPairs lists all key pairs for the current user
func (s *KeyPairService) ListKeyPairs() ([]models.KeyPair, error) {
	var modelKeyPairs []models.KeyPair
	ctx := context.Background()

	// Create a pager
	pager := keypairs.List(s.Client.Compute, nil)

	// Extract key pairs from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		keyPairList, err := keypairs.ExtractKeyPairs(page)
		if err != nil {
			return false, err
		}

		for _, keyPair := range keyPairList {
			modelKeyPair := models.KeyPair{
				Name:        keyPair.Name,
				Fingerprint: keyPair.Fingerprint,
				PublicKey:   keyPair.PublicKey,
				UserID:      keyPair.UserID,
				Type:        keyPair.Type,
			}

			modelKeyPairs = append(modelKeyPairs, modelKeyPair)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelKeyPairs, nil
}

// GetKeyPair gets a key pair by name
func (s *KeyPairService) GetKeyPair(name string) (*models.KeyPair, error) {
	ctx := context.Background()

	// Get the key pair
	keyPair, err := keypairs.Get(ctx, s.Client.Compute, name, nil).Extract()
	if err != nil {
		return nil, err
	}

	// Return the key pair
	modelKeyPair := &models.KeyPair{
		Name:        keyPair.Name,
		Fingerprint: keyPair.Fingerprint,
		PublicKey:   keyPair.PublicKey,
		UserID:      keyPair.UserID,
		Type:        keyPair.Type,
	}

	return modelKeyPair, nil
}

// CreateKeyPair creates a new key pair
func (s *KeyPairService) CreateKeyPair(req models.CreateKeyPairRequest) (*models.KeyPair, error) {
	ctx := context.Background()

	// Define key pair create options
	createOpts := keypairs.CreateOpts{
		Name:      req.Name,
		PublicKey: req.PublicKey,
	}

	// Create the key pair
	keyPair, err := keypairs.Create(ctx, s.Client.Compute, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	// Return the key pair
	modelKeyPair := &models.KeyPair{
		Name:        keyPair.Name,
		Fingerprint: keyPair.Fingerprint,
		PublicKey:   keyPair.PublicKey,
		PrivateKey:  keyPair.PrivateKey,
		UserID:      keyPair.UserID,
		Type:        keyPair.Type,
	}

	return modelKeyPair, nil
}

// DeleteKeyPair deletes a key pair by name
func (s *KeyPairService) DeleteKeyPair(name string) error {
	ctx := context.Background()

	// Delete the key pair
	return keypairs.Delete(ctx, s.Client.Compute, name, nil).ExtractErr()
}
