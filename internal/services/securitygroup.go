package services

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// SecurityGroupService handles operations related to security group resources
type SecurityGroupService struct {
	Client *client.OpenStackClient
}

// NewSecurityGroupService creates a new security group service
func NewSecurityGroupService(client *client.OpenStackClient) *SecurityGroupService {
	return &SecurityGroupService{
		Client: client,
	}
}

// Helper function to convert int to *int
func intPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// ListSecurityGroups lists all security groups
func (s *SecurityGroupService) ListSecurityGroups() ([]models.SecurityGroup, error) {
	// Initialize with empty slice instead of nil
	modelSecurityGroups := []models.SecurityGroup{}
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return modelSecurityGroups, fmt.Errorf("network client is nil")
	}

	// Create a pager
	listOpts := groups.ListOpts{}
	pager := groups.List(s.Client.Network, listOpts)

	// Extract security groups from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		// Extract security groups
		secGroups, err := groups.ExtractGroups(page)
		if err != nil {
			return false, err
		}

		// Convert to our model
		for _, secGroup := range secGroups {
			// Convert security group rules
			sgRules := make([]models.SecurityGroupRule, len(secGroup.Rules))
			for i, rule := range secGroup.Rules {
				sgRules[i] = models.SecurityGroupRule{
					ID:              rule.ID,
					Direction:       rule.Direction,
					EtherType:       rule.EtherType,
					Protocol:        rule.Protocol,
					PortRangeMin:    intPtr(rule.PortRangeMin),
					PortRangeMax:    intPtr(rule.PortRangeMax),
					RemoteIPPrefix:  rule.RemoteIPPrefix,
					RemoteGroupID:   rule.RemoteGroupID,
					SecurityGroupID: rule.SecGroupID, // Note: Gophercloud uses SecGroupID
					ProjectID:       rule.ProjectID,
					CreatedAt:       rule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					UpdatedAt:       rule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				}
			}

			modelSecGroup := models.SecurityGroup{
				ID:                 secGroup.ID,
				Name:               secGroup.Name,
				Description:        secGroup.Description,
				ProjectID:          secGroup.ProjectID,
				CreatedAt:          secGroup.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:          secGroup.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				SecurityGroupRules: sgRules,
			}
			modelSecurityGroups = append(modelSecurityGroups, modelSecGroup)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelSecurityGroups, nil
}

// GetSecurityGroup gets a security group by ID
func (s *SecurityGroupService) GetSecurityGroup(id string) (*models.SecurityGroup, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Get the security group
	secGroup, err := groups.Get(ctx, s.Client.Network, id).Extract()
	if err != nil {
		return nil, err
	}

	// Convert security group rules
	sgRules := make([]models.SecurityGroupRule, len(secGroup.Rules))
	for i, rule := range secGroup.Rules {
		sgRules[i] = models.SecurityGroupRule{
			ID:              rule.ID,
			Direction:       rule.Direction,
			EtherType:       rule.EtherType,
			Protocol:        rule.Protocol,
			PortRangeMin:    intPtr(rule.PortRangeMin),
			PortRangeMax:    intPtr(rule.PortRangeMax),
			RemoteIPPrefix:  rule.RemoteIPPrefix,
			RemoteGroupID:   rule.RemoteGroupID,
			SecurityGroupID: rule.SecGroupID, // Note: Gophercloud uses SecGroupID
			ProjectID:       rule.ProjectID,
			CreatedAt:       rule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       rule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Convert to our model
	modelSecGroup := &models.SecurityGroup{
		ID:                 secGroup.ID,
		Name:               secGroup.Name,
		Description:        secGroup.Description,
		ProjectID:          secGroup.ProjectID,
		CreatedAt:          secGroup.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          secGroup.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		SecurityGroupRules: sgRules,
	}

	return modelSecGroup, nil
}

// CreateSecurityGroup creates a new security group
func (s *SecurityGroupService) CreateSecurityGroup(req models.CreateSecurityGroupRequest) (*models.SecurityGroup, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Define security group create options
	createOpts := groups.CreateOpts{
		Name:        req.Name,
		Description: req.Description,
	}

	// Create the security group
	secGroup, err := groups.Create(ctx, s.Client.Network, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	// Convert security group rules
	sgRules := make([]models.SecurityGroupRule, len(secGroup.Rules))
	for i, rule := range secGroup.Rules {
		sgRules[i] = models.SecurityGroupRule{
			ID:              rule.ID,
			Direction:       rule.Direction,
			EtherType:       rule.EtherType,
			Protocol:        rule.Protocol,
			PortRangeMin:    intPtr(rule.PortRangeMin),
			PortRangeMax:    intPtr(rule.PortRangeMax),
			RemoteIPPrefix:  rule.RemoteIPPrefix,
			RemoteGroupID:   rule.RemoteGroupID,
			SecurityGroupID: rule.SecGroupID, // Note: Gophercloud uses SecGroupID
			ProjectID:       rule.ProjectID,
			CreatedAt:       rule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       rule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Convert to our model
	modelSecGroup := &models.SecurityGroup{
		ID:                 secGroup.ID,
		Name:               secGroup.Name,
		Description:        secGroup.Description,
		ProjectID:          secGroup.ProjectID,
		CreatedAt:          secGroup.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          secGroup.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		SecurityGroupRules: sgRules,
	}

	return modelSecGroup, nil
}

// DeleteSecurityGroup deletes a security group
func (s *SecurityGroupService) DeleteSecurityGroup(id string) error {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return fmt.Errorf("network client is nil")
	}

	// Delete the security group
	return groups.Delete(ctx, s.Client.Network, id).ExtractErr()
}

// ListSecurityGroupRules lists all security group rules
func (s *SecurityGroupService) ListSecurityGroupRules() ([]models.SecurityGroupRule, error) {
	// Initialize with empty slice instead of nil
	modelSecurityGroupRules := []models.SecurityGroupRule{}
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return modelSecurityGroupRules, fmt.Errorf("network client is nil")
	}

	// Create a pager
	listOpts := rules.ListOpts{}
	pager := rules.List(s.Client.Network, listOpts)

	// Extract security group rules from pages
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		// Extract security group rules
		secGroupRules, err := rules.ExtractRules(page)
		if err != nil {
			return false, err
		}

		// Convert to our model
		for _, rule := range secGroupRules {
			modelRule := models.SecurityGroupRule{
				ID:              rule.ID,
				Direction:       rule.Direction,
				EtherType:       rule.EtherType,
				Protocol:        rule.Protocol,
				PortRangeMin:    intPtr(rule.PortRangeMin),
				PortRangeMax:    intPtr(rule.PortRangeMax),
				RemoteIPPrefix:  rule.RemoteIPPrefix,
				RemoteGroupID:   rule.RemoteGroupID,
				SecurityGroupID: rule.SecGroupID, // Note: Gophercloud uses SecGroupID
				ProjectID:       rule.ProjectID,
				CreatedAt:       rule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:       rule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			modelSecurityGroupRules = append(modelSecurityGroupRules, modelRule)
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return modelSecurityGroupRules, nil
}

// Helper function to convert *int to int for CreateOpts
func intFromPtr(ptr *int) int {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// CreateSecurityGroupRule creates a new security group rule
func (s *SecurityGroupService) CreateSecurityGroupRule(req models.CreateSecurityGroupRuleRequest) (*models.SecurityGroupRule, error) {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return nil, fmt.Errorf("network client is nil")
	}

	// Define security group rule create options
	createOpts := rules.CreateOpts{
		Direction:      rules.RuleDirection(req.Direction),
		EtherType:      rules.RuleEtherType(req.EtherType),
		Protocol:       rules.RuleProtocol(req.Protocol),
		PortRangeMin:   intFromPtr(req.PortRangeMin),
		PortRangeMax:   intFromPtr(req.PortRangeMax),
		RemoteIPPrefix: req.RemoteIPPrefix,
		RemoteGroupID:  req.RemoteGroupID,
		SecGroupID:     req.SecurityGroupID,
	}

	// Create the security group rule
	rule, err := rules.Create(ctx, s.Client.Network, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	// Convert to our model
	modelRule := &models.SecurityGroupRule{
		ID:              rule.ID,
		Direction:       rule.Direction,
		EtherType:       rule.EtherType,
		Protocol:        rule.Protocol,
		PortRangeMin:    intPtr(rule.PortRangeMin),
		PortRangeMax:    intPtr(rule.PortRangeMax),
		RemoteIPPrefix:  rule.RemoteIPPrefix,
		RemoteGroupID:   rule.RemoteGroupID,
		SecurityGroupID: rule.SecGroupID, // Note: Gophercloud uses SecGroupID
		ProjectID:       rule.ProjectID,
		CreatedAt:       rule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       rule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return modelRule, nil
}

// DeleteSecurityGroupRule deletes a security group rule
func (s *SecurityGroupService) DeleteSecurityGroupRule(id string) error {
	ctx := context.Background()

	// Check if Network client is nil
	if s.Client == nil || s.Client.Network == nil {
		return fmt.Errorf("network client is nil")
	}

	// Delete the security group rule
	return rules.Delete(ctx, s.Client.Network, id).ExtractErr()
}
