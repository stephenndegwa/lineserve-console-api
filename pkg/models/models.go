package models

import "time"

// LoginRequest represents a login request
type LoginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	DomainName string `json:"domain_name"`
}

// LoginResponse represents the login response body
type LoginResponse struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Projects  []Project `json:"projects"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ProjectScopeRequest represents a request to get a project-scoped token
type ProjectScopeRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	DomainName string `json:"domain_name"`
	ProjectID  string `json:"project_id"`
}

// ProjectScopeResponse represents the response for a project-scoped token
type ProjectScopeResponse struct {
	Token     string    `json:"token"`
	ProjectID string    `json:"project_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// RegisterResponse represents the registration response body
type RegisterResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	ProjectID string `json:"project_id"`
	Message   string `json:"message"`
}

// LineserveCloudUser represents a user in the lineserve_cloud_users table
type LineserveCloudUser struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Email              string    `json:"email"`
	Phone              string    `json:"phone"`
	PasswordHash       string    `json:"password_hash"`
	OpenstackUserID    string    `json:"openstack_user_id"`
	OpenstackProjectID string    `json:"openstack_project_id"`
	CreatedAt          time.Time `json:"created_at"`
	Verified           bool      `json:"verified"`
}

// UserProject represents a mapping between a user and a project
type UserProject struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProjectID string    `json:"project_id"`
	RoleID    string    `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

// EmailVerification represents an email verification record
type EmailVerification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID    string `json:"user_id"`
	ProjectID string `json:"project_id,omitempty"`
	Role      string `json:"role,omitempty"`
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
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
	DomainID    string `json:"domain_id,omitempty"`
}

// ProjectListResponse represents the response for listing projects
type ProjectListResponse struct {
	Projects []Project `json:"projects"`
}

// Role represents an identity role
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
