package services

import (
	"context"

	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// IdentityService handles operations related to identity resources
type IdentityService struct {
	Client *client.OpenStackClient
}

// NewIdentityService creates a new identity service
func NewIdentityService(client *client.OpenStackClient) *IdentityService {
	return &IdentityService{
		Client: client,
	}
}

// ListProjects lists all projects
func (s *IdentityService) ListProjects() ([]models.Project, error) {
	var modelProjects []models.Project
	ctx := context.Background()

	// Create a pager
	pager := projects.List(s.Client.Identity, projects.ListOpts{})

	// Extract projects from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		projectList, err := projects.ExtractProjects(page)
		if err != nil {
			return false, err
		}

		for _, project := range projectList {
			modelProject := models.Project{
				ID:          project.ID,
				Name:        project.Name,
				Description: project.Description,
				Enabled:     project.Enabled,
				DomainID:    project.DomainID,
			}

			modelProjects = append(modelProjects, modelProject)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelProjects, nil
}

// GetProject gets a project by ID
func (s *IdentityService) GetProject(id string) (*models.Project, error) {
	ctx := context.Background()

	// Get the project
	project, err := projects.Get(ctx, s.Client.Identity, id).Extract()
	if err != nil {
		return nil, err
	}

	// Return the project
	modelProject := &models.Project{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Enabled:     project.Enabled,
		DomainID:    project.DomainID,
	}

	return modelProject, nil
}
