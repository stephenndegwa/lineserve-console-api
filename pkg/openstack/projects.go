package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
)

// CreateProject creates a new project (admin only)
func CreateProject(ctx context.Context, provider *gophercloud.ProviderClient, name, description, domainID string) (*Project, error) {
	identityClient, err := NewIdentityClient(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity client: %w", err)
	}

	createOpts := projects.CreateOpts{
		Name:        name,
		Description: description,
		DomainID:    domainID,
	}

	project, err := projects.Create(ctx, identityClient, createOpts).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &Project{
		ID:       project.ID,
		Name:     project.Name,
		DomainID: project.DomainID,
	}, nil
}

// GetProject gets a project by ID
func GetProject(ctx context.Context, provider *gophercloud.ProviderClient, projectID string) (*Project, error) {
	identityClient, err := NewIdentityClient(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity client: %w", err)
	}

	project, err := projects.Get(ctx, identityClient, projectID).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &Project{
		ID:       project.ID,
		Name:     project.Name,
		DomainID: project.DomainID,
	}, nil
}

// ListAllProjects lists all projects in the system (admin only)
func ListAllProjects(ctx context.Context, provider *gophercloud.ProviderClient) ([]Project, error) {
	identityClient, err := NewIdentityClient(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity client: %w", err)
	}

	allPages, err := projects.List(identityClient, projects.ListOpts{}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	allProjects, err := projects.ExtractProjects(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract projects: %w", err)
	}

	result := make([]Project, len(allProjects))
	for i, p := range allProjects {
		result[i] = Project{
			ID:       p.ID,
			Name:     p.Name,
			DomainID: p.DomainID,
		}
	}

	return result, nil
}

// UpdateProject updates a project (admin only)
func UpdateProject(ctx context.Context, provider *gophercloud.ProviderClient, projectID, name, description string) (*Project, error) {
	identityClient, err := NewIdentityClient(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity client: %w", err)
	}

	// In Gophercloud v2.7.0, Name is string but Description is *string
	descPtr := description

	updateOpts := projects.UpdateOpts{
		Name:        name,
		Description: &descPtr,
	}

	project, err := projects.Update(ctx, identityClient, projectID, updateOpts).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &Project{
		ID:       project.ID,
		Name:     project.Name,
		DomainID: project.DomainID,
	}, nil
}

// DeleteProject deletes a project (admin only)
func DeleteProject(ctx context.Context, provider *gophercloud.ProviderClient, projectID string) error {
	identityClient, err := NewIdentityClient(provider)
	if err != nil {
		return fmt.Errorf("failed to create identity client: %w", err)
	}

	err = projects.Delete(ctx, identityClient, projectID).ExtractErr()
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}
