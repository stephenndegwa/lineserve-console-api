package openstack

import (
	"os"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
)

// EndpointOpts returns the standard endpoint options with region from environment
func EndpointOpts() gophercloud.EndpointOpts {
	return gophercloud.EndpointOpts{
		Region:       os.Getenv("OS_REGION_NAME"),
		Availability: gophercloud.AvailabilityPublic,
	}
}

// NewComputeClient creates a new compute service client
func NewComputeClient(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewComputeV2(provider, EndpointOpts())
}

// NewNetworkClient creates a new network service client
func NewNetworkClient(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewNetworkV2(provider, EndpointOpts())
}

// NewBlockStorageClient creates a new block storage service client
func NewBlockStorageClient(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewBlockStorageV3(provider, EndpointOpts())
}

// NewIdentityClient creates a new identity service client
func NewIdentityClient(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewIdentityV3(provider, EndpointOpts())
}

// NewImageClient creates a new image service client
func NewImageClient(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewImageV2(provider, EndpointOpts())
}

// NewObjectStorageClient creates a new object storage service client
func NewObjectStorageClient(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewObjectStorageV1(provider, EndpointOpts())
}

// NewLoadBalancerClient creates a new load balancer service client
func NewLoadBalancerClient(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewLoadBalancerV2(provider, EndpointOpts())
}
