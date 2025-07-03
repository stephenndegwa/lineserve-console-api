package models

import "time"

// LoginRequest represents a login request
type LoginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	AuthURL     string `json:"auth_url"`
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	DomainName  string `json:"domain_name"`
	RegionName  string `json:"region_name"`
}

// LoginResponse represents the login response body
type LoginResponse struct {
	Token string `json:"token"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Address represents an instance network address
type Address struct {
	Type    string `json:"type"`
	Address string `json:"addr"`
}

// Instance represents a compute instance
type Instance struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Flavor    string                 `json:"flavor"`
	Image     string                 `json:"image"`
	Addresses map[string][]Address   `json:"addresses"`
	Created   time.Time              `json:"created"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CreateInstanceRequest represents a request to create an instance
type CreateInstanceRequest struct {
	Name      string `json:"name"`
	FlavorID  string `json:"flavor_id"`
	ImageID   string `json:"image_id"`
	NetworkID string `json:"network_id"`
	KeyName   string `json:"key_name,omitempty"`
}

// Flavor represents a compute flavor
type Flavor struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RAM      int    `json:"ram"`
	Disk     int    `json:"disk"`
	VCPUs    int    `json:"vcpus"`
	IsPublic bool   `json:"is_public"`
}

// Image represents an image
type Image struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Status     string            `json:"status"`
	Size       int64             `json:"size"`
	Visibility string            `json:"visibility"`
	Tags       []string          `json:"tags"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Properties map[string]string `json:"properties,omitempty"`
}

// Network represents a network
type Network struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Shared   bool   `json:"shared"`
	External bool   `json:"external"`
}

// VolumeAttachment represents a volume attachment
type VolumeAttachment struct {
	ServerID     string `json:"server_id"`
	AttachmentID string `json:"attachment_id"`
	DeviceName   string `json:"device"`
}

// Volume represents a volume
type Volume struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Status           string             `json:"status"`
	Size             int                `json:"size"`
	VolumeType       string             `json:"volume_type"`
	AvailabilityZone string             `json:"availability_zone"`
	CreatedAt        time.Time          `json:"created_at"`
	Attachments      []VolumeAttachment `json:"attachments"`
}

// CreateVolumeRequest represents a request to create a volume
type CreateVolumeRequest struct {
	Name             string `json:"name"`
	Size             int    `json:"size"`
	VolumeType       string `json:"volume_type,omitempty"`
	AvailabilityZone string `json:"availability_zone,omitempty"`
}

// Project represents an identity project
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	DomainID    string `json:"domain_id"`
}
