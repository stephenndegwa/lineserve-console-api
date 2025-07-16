package openstack

import (
	"context"
	"fmt"
	"os"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/tokens"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/users"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// Project represents an OpenStack project
type Project struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	DomainID string `json:"domain_id"`
}

// AuthenticateUnscoped authenticates with OpenStack using unscoped auth
// Returns the provider client, user ID, and error if any
func AuthenticateUnscoped(ctx context.Context, username, password, domainName string) (*gophercloud.ProviderClient, string, error) {
	// Create auth options for unscoped authentication
	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: os.Getenv("OS_AUTH_URL"),
		Username:         username,
		Password:         password,
		DomainName:       domainName,
		AllowReauth:      true,
	}

	// Authenticate with OpenStack
	provider, err := openstack.AuthenticatedClient(ctx, authOpts)
	if err != nil {
		return nil, "", fmt.Errorf("failed to authenticate with OpenStack: %w", err)
	}

	// Create identity client to extract user information
	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create identity client: %w", err)
	}

	// Get token information
	tokenResult := tokens.Get(ctx, identityClient, provider.Token())
	user, err := tokenResult.ExtractUser()
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract user from token: %w", err)
	}

	return provider, user.ID, nil
}

// AuthenticateScoped authenticates with OpenStack using project-scoped auth
// Returns the provider client and error if any
func AuthenticateScoped(ctx context.Context, username, password, domainName, projectID string) (*gophercloud.ProviderClient, error) {
	// Create auth options for project-scoped authentication
	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: os.Getenv("OS_AUTH_URL"),
		Username:         username,
		Password:         password,
		DomainName:       domainName,
		AllowReauth:      true,
		Scope: &gophercloud.AuthScope{
			ProjectID: projectID,
		},
	}

	// Authenticate with OpenStack
	provider, err := openstack.AuthenticatedClient(ctx, authOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with OpenStack (scoped): %w", err)
	}

	return provider, nil
}

// AuthenticateWithToken authenticates with an existing token and scopes it to a project
func AuthenticateWithToken(ctx context.Context, tokenID, projectID string) (*gophercloud.ProviderClient, error) {
	// Create auth options for token-based authentication with project scope
	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: os.Getenv("OS_AUTH_URL"),
		TokenID:          tokenID,
		AllowReauth:      false, // Cannot reauth with a token
		Scope: &gophercloud.AuthScope{
			ProjectID: projectID,
		},
	}

	// Authenticate with OpenStack
	provider, err := openstack.AuthenticatedClient(ctx, authOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with token: %w", err)
	}

	return provider, nil
}

// ListUserProjects lists all projects accessible by the user
func ListUserProjects(ctx context.Context, provider *gophercloud.ProviderClient, userID string) ([]Project, error) {
	// Create identity client
	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to create identity client: %w", err)
	}

	// List projects for the user using the correct API approach
	pager := projects.ListAvailable(identityClient)
	var allProjects []projects.Project

	err = pager.EachPage(ctx, func(_ context.Context, page pagination.Page) (bool, error) {
		projectsOnPage, err := projects.ExtractProjects(page)
		if err != nil {
			return false, err
		}
		allProjects = append(allProjects, projectsOnPage...)
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Convert to our Project type
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

// GetAuthToken extracts the token from a provider client
func GetAuthToken(provider *gophercloud.ProviderClient) string {
	if provider == nil {
		return ""
	}
	return provider.Token()
}

// GetAuthResult extracts the auth result from a provider client
func GetAuthResult(ctx context.Context, provider *gophercloud.ProviderClient) (*tokens.GetResult, error) {
	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to create identity client: %w", err)
	}

	result := tokens.Get(ctx, identityClient, provider.Token())
	return &result, nil
}

// GetAdminProvider returns a provider client with admin credentials from environment variables
func GetAdminProvider(ctx context.Context) (*gophercloud.ProviderClient, error) {
	// Get admin credentials from environment variables with fallbacks
	authURL := os.Getenv("OS_AUTH_URL")
	if authURL == "" {
		authURL = "http://102.209.139.152/identity/v3" // fallback
	}

	username := os.Getenv("OS_ADMIN_USERNAME")
	if username == "" {
		username = "admin" // fallback
	}

	password := os.Getenv("OS_ADMIN_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("OS_ADMIN_PASSWORD environment variable must be set")
	}

	domainName := os.Getenv("OS_ADMIN_DOMAIN_NAME")
	if domainName == "" {
		domainName = "Default" // fallback
	}

	projectName := os.Getenv("OS_ADMIN_PROJECT_NAME")
	if projectName == "" {
		projectName = "admin" // fallback
	}

	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: authURL,
		Username:         username,
		Password:         password,
		DomainName:       domainName,
		AllowReauth:      true,
		Scope: &gophercloud.AuthScope{
			ProjectName: projectName,
			DomainName:  domainName,
		},
	}

	// Print debug info (without password)
	fmt.Println("Using admin credentials:")
	fmt.Printf("Auth URL: %s\n", authOpts.IdentityEndpoint)
	fmt.Printf("Username: %s\n", authOpts.Username)
	fmt.Printf("Domain Name: %s\n", authOpts.DomainName)
	fmt.Printf("Project Name: %s\n", authOpts.Scope.ProjectName)

	// Authenticate with OpenStack
	provider, err := openstack.AuthenticatedClient(ctx, authOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with OpenStack as admin: %w", err)
	}

	return provider, nil
}

// CreateUserAccount creates a new OpenStack user account
func CreateUserAccount(ctx context.Context, provider *gophercloud.ProviderClient, name, emailAddress, password, domainName string) (*users.User, error) {
	// Create identity client
	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to create identity client: %w", err)
	}

	// Get domain ID from name if provided
	var domainID string
	if domainName != "" {
		// This would require additional code to look up domain ID by name
		// For simplicity, we'll use the default domain ID if domain name is "default"
		if domainName == "default" || domainName == "Default" {
			domainID = "default"
		}
	}

	// Note: CreateOpts doesn't have an Email field directly
	// We need to use the Extra map instead
	createOpts := users.CreateOpts{
		Name:     name,
		Password: password,
		DomainID: domainID,
		Enabled:  gophercloud.Enabled,
		Extra: map[string]any{
			"email": emailAddress,
		},
	}

	user, err := users.Create(ctx, identityClient, createOpts).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
