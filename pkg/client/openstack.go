package client

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	osconfig "github.com/gophercloud/gophercloud/v2/openstack/config"
	"github.com/gophercloud/gophercloud/v2/openstack/config/clouds"
	appconfig "github.com/lineserve/lineserve-api/pkg/config"
)

// OpenStackClient represents an OpenStack client
type OpenStackClient struct {
	Provider *gophercloud.ProviderClient
	Compute  *gophercloud.ServiceClient
	Image    *gophercloud.ServiceClient
	Network  *gophercloud.ServiceClient
	Volume   *gophercloud.ServiceClient
	Identity *gophercloud.ServiceClient
}

// AuthResponse represents the OpenStack authentication response
type AuthResponse struct {
	Token struct {
		Methods   []string      `json:"methods"`
		ExpiresAt string        `json:"expires_at"`
		IssuedAt  string        `json:"issued_at"`
		AuditIDs  []string      `json:"audit_ids"`
		Catalog   []interface{} `json:"catalog"`
		Project   interface{}   `json:"project"`
		User      interface{}   `json:"user"`
		Roles     []interface{} `json:"roles"`
		IsDomain  bool          `json:"is_domain"`
	} `json:"token"`
}

// NewOpenStackClient creates a new OpenStack client
func NewOpenStackClient() (*OpenStackClient, error) {
	// Create a context
	ctx := context.Background()

	// Try to load from clouds.yaml first
	authOptions, endpointOptions, tlsConfig, err := clouds.Parse()
	if err != nil {
		// If clouds.yaml not found, try environment variables
		authOptions := gophercloud.AuthOptions{
			IdentityEndpoint: os.Getenv("OS_AUTH_URL"),
			Username:         os.Getenv("OS_USERNAME"),
			Password:         os.Getenv("OS_PASSWORD"),
			TenantID:         os.Getenv("OS_PROJECT_ID"),
			TenantName:       os.Getenv("OS_PROJECT_NAME"),
			AllowReauth:      true,
			// Use explicit scope structure
			Scope: &gophercloud.AuthScope{
				ProjectName: os.Getenv("OS_PROJECT_NAME"),
				DomainName:  os.Getenv("OS_PROJECT_DOMAIN_NAME"),
			},
		}

		// Set domain information for v3 authentication
		if os.Getenv("OS_USER_DOMAIN_NAME") != "" {
			authOptions.DomainName = os.Getenv("OS_USER_DOMAIN_NAME")
		}

		// Print debug information
		fmt.Printf("Auth URL: %s\n", authOptions.IdentityEndpoint)
		fmt.Printf("Username: %s\n", authOptions.Username)
		fmt.Printf("Project ID: %s\n", authOptions.TenantID)
		fmt.Printf("Project Name: %s\n", authOptions.TenantName)
		fmt.Printf("Domain Name: %s\n", authOptions.DomainName)
		if authOptions.Scope != nil {
			fmt.Printf("Scope Project Name: %s\n", authOptions.Scope.ProjectName)
			fmt.Printf("Scope Domain Name: %s\n", authOptions.Scope.DomainName)
		}

		// Validate required fields
		if authOptions.IdentityEndpoint == "" {
			return nil, errors.New("OS_AUTH_URL is required")
		}
		if authOptions.Username == "" {
			return nil, errors.New("OS_USERNAME is required")
		}
		if authOptions.Password == "" {
			return nil, errors.New("OS_PASSWORD is required")
		}

		// Create provider client
		provider, err := openstack.AuthenticatedClient(ctx, authOptions)
		if err != nil {
			return nil, err
		}

		// Create service clients
		endpointOpts := gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		}

		// Create compute client
		compute, err := openstack.NewComputeV2(provider, endpointOpts)
		if err != nil {
			return nil, err
		}

		// Create image client
		image, err := openstack.NewImageV2(provider, endpointOpts)
		if err != nil {
			return nil, err
		}

		// Create network client
		network, err := openstack.NewNetworkV2(provider, endpointOpts)
		if err != nil {
			return nil, err
		}

		// Create volume client
		volume, err := openstack.NewBlockStorageV3(provider, endpointOpts)
		if err != nil {
			return nil, err
		}

		// Create identity client
		identity, err := openstack.NewIdentityV3(provider, endpointOpts)
		if err != nil {
			return nil, err
		}

		// Return the client
		return &OpenStackClient{
			Provider: provider,
			Compute:  compute,
			Image:    image,
			Network:  network,
			Volume:   volume,
			Identity: identity,
		}, nil
	} else {
		// Use clouds.yaml configuration
		provider, err := osconfig.NewProviderClient(ctx, authOptions, osconfig.WithTLSConfig(tlsConfig))
		if err != nil {
			return nil, err
		}

		// Create compute client
		compute, err := openstack.NewComputeV2(provider, endpointOptions)
		if err != nil {
			return nil, err
		}

		// Create image client
		image, err := openstack.NewImageV2(provider, endpointOptions)
		if err != nil {
			return nil, err
		}

		// Create network client
		network, err := openstack.NewNetworkV2(provider, endpointOptions)
		if err != nil {
			return nil, err
		}

		// Create volume client
		volume, err := openstack.NewBlockStorageV3(provider, endpointOptions)
		if err != nil {
			return nil, err
		}

		// Create identity client
		identity, err := openstack.NewIdentityV3(provider, endpointOptions)
		if err != nil {
			return nil, err
		}

		// Return the client
		return &OpenStackClient{
			Provider: provider,
			Compute:  compute,
			Image:    image,
			Network:  network,
			Volume:   volume,
			Identity: identity,
		}, nil
	}
}

// Authenticate verifies if the provided credentials are valid
func Authenticate(cfg *appconfig.Config) error {
	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: cfg.OSAuthURL,
		Username:         cfg.OSUsername,
		Password:         cfg.OSPassword,
		AllowReauth:      true,
		Scope: &gophercloud.AuthScope{
			ProjectName: cfg.OSProjectName,
			DomainName:  cfg.OSProjectDomainName,
		},
		DomainName: cfg.OSUserDomainName,
	}

	_, err := openstack.AuthenticatedClient(context.Background(), authOpts)
	return err
}
