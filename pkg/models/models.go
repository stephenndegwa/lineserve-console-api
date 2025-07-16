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

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
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

// UpdateInstanceRequest represents a request to update an instance
type UpdateInstanceRequest struct {
	Name       *string `json:"name,omitempty"`
	AccessIPv4 *string `json:"access_ipv4,omitempty"`
	AccessIPv6 *string `json:"access_ipv6,omitempty"`
}

// InstanceActionRequest represents a request to perform an action on an instance
type InstanceActionRequest struct {
	Action string `json:"action" binding:"required"`
	Type   string `json:"type,omitempty"` // For reboot: "SOFT" or "HARD"
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

// CreateImageRequest represents a request to create an image
type CreateImageRequest struct {
	Name            string            `json:"name" binding:"required"`
	ContainerFormat string            `json:"container_format,omitempty"`
	DiskFormat      string            `json:"disk_format,omitempty"`
	MinDisk         int               `json:"min_disk,omitempty"`
	MinRAM          int               `json:"min_ram,omitempty"`
	Protected       bool              `json:"protected,omitempty"`
	Visibility      string            `json:"visibility,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	Properties      map[string]string `json:"properties,omitempty"`
}

// Network represents a network
type Network struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Shared   bool   `json:"shared"`
	External bool   `json:"external"`
}

// CreateNetworkRequest represents a request to create a network
type CreateNetworkRequest struct {
	Name         string `json:"name"`
	Shared       bool   `json:"shared"`
	External     bool   `json:"external"`
	AdminStateUp bool   `json:"admin_state_up"`
}

// VolumeAttachment represents a volume attachment
type VolumeAttachment struct {
	ID         string `json:"id"`
	VolumeID   string `json:"volume_id"`
	InstanceID string `json:"instance_id"`
	Device     string `json:"device"`
	Status     string `json:"status"`
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

// VolumeType represents a volume type
type VolumeType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// KeyPair represents an SSH key pair
type KeyPair struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
	PrivateKey  string `json:"private_key,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Type        string `json:"type,omitempty"`
}

// CreateKeyPairRequest represents a request to create a key pair
type CreateKeyPairRequest struct {
	Name      string `json:"name" binding:"required"`
	PublicKey string `json:"public_key,omitempty"`
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

// FloatingIP represents a floating IP
type FloatingIP struct {
	ID                string `json:"id"`
	FloatingIP        string `json:"floating_ip_address"`
	FloatingNetworkID string `json:"floating_network_id"`
	Status            string `json:"status"`
	PortID            string `json:"port_id,omitempty"`
	FixedIP           string `json:"fixed_ip_address,omitempty"`
	RouterID          string `json:"router_id,omitempty"`
	Description       string `json:"description,omitempty"`
	ProjectID         string `json:"project_id"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// CreateFloatingIPRequest represents a request to create a floating IP
type CreateFloatingIPRequest struct {
	FloatingNetworkID string `json:"floating_network_id,omitempty"`
	Description       string `json:"description,omitempty"`
}

// UpdateFloatingIPRequest represents a request to update a floating IP
type UpdateFloatingIPRequest struct {
	PortID  *string `json:"port_id,omitempty"`
	FixedIP *string `json:"fixed_ip_address,omitempty"`
}

// SecurityGroup represents a security group
type SecurityGroup struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	ProjectID          string              `json:"project_id"`
	CreatedAt          string              `json:"created_at"`
	UpdatedAt          string              `json:"updated_at"`
	SecurityGroupRules []SecurityGroupRule `json:"security_group_rules"`
}

// CreateSecurityGroupRequest represents a request to create a security group
type CreateSecurityGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
}

// SecurityGroupRule represents a security group rule
type SecurityGroupRule struct {
	ID              string `json:"id"`
	Direction       string `json:"direction"`
	EtherType       string `json:"ethertype"`
	Protocol        string `json:"protocol,omitempty"`
	PortRangeMin    *int   `json:"port_range_min,omitempty"`
	PortRangeMax    *int   `json:"port_range_max,omitempty"`
	RemoteIPPrefix  string `json:"remote_ip_prefix,omitempty"`
	RemoteGroupID   string `json:"remote_group_id,omitempty"`
	SecurityGroupID string `json:"security_group_id"`
	ProjectID       string `json:"project_id"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// CreateSecurityGroupRuleRequest represents a request to create a security group rule
type CreateSecurityGroupRuleRequest struct {
	Direction       string `json:"direction" binding:"required"`
	EtherType       string `json:"ethertype" binding:"required"`
	Protocol        string `json:"protocol,omitempty"`
	PortRangeMin    *int   `json:"port_range_min,omitempty"`
	PortRangeMax    *int   `json:"port_range_max,omitempty"`
	RemoteIPPrefix  string `json:"remote_ip_prefix,omitempty"`
	RemoteGroupID   string `json:"remote_group_id,omitempty"`
	SecurityGroupID string `json:"security_group_id" binding:"required"`
}

// Subnet represents a subnet
type Subnet struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	NetworkID       string           `json:"network_id"`
	CIDR            string           `json:"cidr"`
	GatewayIP       *string          `json:"gateway_ip"`
	IPVersion       int              `json:"ip_version"`
	EnableDHCP      bool             `json:"enable_dhcp"`
	DNSNameservers  []string         `json:"dns_nameservers"`
	AllocationPools []AllocationPool `json:"allocation_pools"`
	HostRoutes      []HostRoute      `json:"host_routes"`
	ServiceTypes    []string         `json:"service_types"`
	ProjectID       string           `json:"project_id"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
}

// AllocationPool represents an allocation pool
type AllocationPool struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// HostRoute represents a host route
type HostRoute struct {
	DestinationCIDR string `json:"destination"`
	NextHop         string `json:"nexthop"`
}

// CreateSubnetRequest represents a request to create a subnet
type CreateSubnetRequest struct {
	NetworkID       string           `json:"network_id" binding:"required"`
	Name            string           `json:"name,omitempty"`
	CIDR            string           `json:"cidr" binding:"required"`
	IPVersion       int              `json:"ip_version" binding:"required"`
	GatewayIP       *string          `json:"gateway_ip,omitempty"`
	EnableDHCP      *bool            `json:"enable_dhcp,omitempty"`
	DNSNameservers  []string         `json:"dns_nameservers,omitempty"`
	AllocationPools []AllocationPool `json:"allocation_pools,omitempty"`
	HostRoutes      []HostRoute      `json:"host_routes,omitempty"`
	ServiceTypes    []string         `json:"service_types,omitempty"`
}

// Router represents a router
type Router struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Status       string       `json:"status"`
	AdminStateUp bool         `json:"admin_state_up"`
	GatewayInfo  *GatewayInfo `json:"external_gateway_info,omitempty"`
	Routes       []Route      `json:"routes"`
	ProjectID    string       `json:"project_id"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
}

// GatewayInfo represents gateway information for a router
type GatewayInfo struct {
	NetworkID        string            `json:"network_id"`
	EnableSNAT       *bool             `json:"enable_snat,omitempty"`
	ExternalFixedIPs []ExternalFixedIP `json:"external_fixed_ips,omitempty"`
}

// ExternalFixedIP represents an external fixed IP for a router
type ExternalFixedIP struct {
	SubnetID  string `json:"subnet_id"`
	IPAddress string `json:"ip_address"`
}

// Route represents a route for a router
type Route struct {
	DestinationCIDR string `json:"destination"`
	NextHop         string `json:"nexthop"`
}

// CreateRouterRequest represents a request to create a router
type CreateRouterRequest struct {
	Name         string       `json:"name,omitempty"`
	AdminStateUp *bool        `json:"admin_state_up,omitempty"`
	GatewayInfo  *GatewayInfo `json:"external_gateway_info,omitempty"`
	Routes       []Route      `json:"routes,omitempty"`
}

// RouterInterface represents a router interface
type RouterInterface struct {
	ID        string `json:"id"`
	SubnetID  string `json:"subnet_id"`
	PortID    string `json:"port_id"`
	TenantID  string `json:"tenant_id"`
	ProjectID string `json:"project_id"`
}

// RouterInterfaceRequest represents a request to add/remove a router interface
type RouterInterfaceRequest struct {
	SubnetID string `json:"subnet_id,omitempty"`
	PortID   string `json:"port_id,omitempty"`
}

// VolumeAttachRequest represents a request to attach a volume to an instance
type VolumeAttachRequest struct {
	InstanceID string `json:"instance_id" binding:"required"`
	Device     string `json:"device,omitempty"`
}

// VolumeDetachRequest represents a request to detach a volume from an instance
type VolumeDetachRequest struct {
	AttachmentID string `json:"attachment_id,omitempty"`
}

// VolumeResizeRequest represents a request to resize a volume
type VolumeResizeRequest struct {
	NewSize int `json:"new_size" binding:"required"`
}

// Role represents an identity role
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// VPSPlan represents a VPS plan
type VPSPlan struct {
	ID                string    `json:"id"`
	PlanCode          string    `json:"plan_code"`
	Name              string    `json:"name"`
	VCPU              int       `json:"vcpu"`
	RAMGB             int       `json:"ram_gb"`
	StorageGB         int       `json:"storage_gb"`
	PriceMonthly      float64   `json:"price_monthly"`
	PriceCommit3M     float64   `json:"price_commit_3m,omitempty"`
	PriceCommit6M     float64   `json:"price_commit_6m,omitempty"`
	PriceCommit12M    float64   `json:"price_commit_12m,omitempty"`
	PriceCommit24M    float64   `json:"price_commit_24m,omitempty"`
	IsWindowsAvail    bool      `json:"is_windows_avail"`
	IsBackupAvail     bool      `json:"is_backup_avail"`
	IsPublicIPAvail   bool      `json:"is_public_ip_avail"`
	OpenStackFlavorID string    `json:"openstack_flavor_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// VPSSubscription represents a VPS subscription
type VPSSubscription struct {
	ID                   string    `json:"id,omitempty"`
	UserID               string    `json:"user_id"`
	PlanID               string    `json:"plan_id"`
	CommitPeriod         int       `json:"commit_period"` // in months
	Price                float64   `json:"price"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	RenewalDueDate       time.Time `json:"renewal_due_date"`
	AutoRenew            bool      `json:"auto_renew"`
	Status               string    `json:"status"` // active, grace, expired, cancelled
	InstanceID           string    `json:"instance_id,omitempty"`
	OpenStackProjectID   string    `json:"openstack_project_id,omitempty"`
	StripeSubscriptionID string    `json:"stripe_subscription_id,omitempty"`
	PaymentID            string    `json:"payment_id,omitempty"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
	UpdatedAt            time.Time `json:"updated_at,omitempty"`
	Plan                 *VPSPlan  `json:"plan,omitempty"` // Embedded plan details
}

// VPSSubscriptionRequest represents a request to subscribe to a VPS plan
type VPSSubscriptionRequest struct {
	PlanCode     string `json:"plan_code" binding:"required"`
	CommitPeriod int    `json:"commit_period" binding:"required"` // in months
}

// VPSSubscriptionResponse represents the response for a VPS subscription request
type VPSSubscriptionResponse struct {
	Subscription VPSSubscription `json:"subscription"`
	Message      string          `json:"message"`
}

// VPSPlansResponse represents the response for listing VPS plans
type VPSPlansResponse struct {
	Plans []VPSPlan `json:"plans"`
}

// VPSSubscriptionsResponse represents the response for listing VPS subscriptions
type VPSSubscriptionsResponse struct {
	Subscriptions []VPSSubscription `json:"subscriptions"`
}

// VPSCancelRequest represents a request to cancel a VPS subscription
type VPSCancelRequest struct {
	AutoRenew bool `json:"auto_renew"`
}

// VPSRenewalResult represents the result of a VPS renewal operation
type VPSRenewalResult struct {
	SubscriptionID string  `json:"subscription_id"`
	UserID         string  `json:"user_id"`
	PlanCode       string  `json:"plan_code"`
	Price          float64 `json:"price"`
	Status         string  `json:"status"`
	RenewalResult  string  `json:"renewal_result"`
}

// VPSInvoice represents a VPS invoice
type VPSInvoice struct {
	ID                     string     `json:"id,omitempty"`
	UserID                 string     `json:"user_id"`
	SubscriptionID         string     `json:"subscription_id,omitempty"`
	PlanCode               string     `json:"plan_code"`
	PeriodMonths           int        `json:"period_months"`
	Amount                 float64    `json:"amount"`
	Currency               string     `json:"currency"`
	Status                 string     `json:"status"`
	PaymentMethodID        string     `json:"payment_method_id,omitempty"`
	PaymentIntentID        string     `json:"payment_intent_id,omitempty"`
	TxRef                  string     `json:"tx_ref,omitempty"`
	StripeSessionID        string     `json:"stripe_session_id,omitempty"`
	StripePaymentID        string     `json:"stripe_payment_id,omitempty"`
	MPesaCheckoutRequestID string     `json:"mpesa_checkout_request_id,omitempty"`
	MPesaReceiptNo         string     `json:"mpesa_receipt_no,omitempty"`
	MPesaPhoneNumber       string     `json:"mpesa_phone_number,omitempty"`
	CreatedAt              time.Time  `json:"created_at,omitempty"`
	ExpiresAt              time.Time  `json:"expires_at"`
	PaidAt                 *time.Time `json:"paid_at,omitempty"`
	UpdatedAt              time.Time  `json:"updated_at,omitempty"`
}

// VPSOrderRequest represents a request to order a VPS
type VPSOrderRequest struct {
	PlanCode        string `json:"plan_code" binding:"required"`
	CommitPeriod    int    `json:"commit_period" binding:"required"` // in months
	PaymentMethodID string `json:"payment_method_id,omitempty"`
}

// VPSOrderResponse represents the response for a VPS order request
type VPSOrderResponse struct {
	SubscriptionID string  `json:"subscription_id"`
	InvoiceID      string  `json:"invoice_id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	PaymentURL     string  `json:"payment_url"`
}

// VPSInvoicePayRequest represents a request to pay a VPS invoice
type VPSInvoicePayRequest struct {
	PaymentMethodID string `json:"payment_method_id" binding:"required"`
	PaymentMethod   string `json:"payment_method,omitempty"` // "card", "paypal"
	PayPalOrderID   string `json:"paypal_order_id,omitempty"`
}

// VPSInvoicePayResponse represents the response for a VPS invoice payment
type VPSInvoicePayResponse struct {
	Status         string `json:"status"`
	SubscriptionID string `json:"subscription_id"`
	InstanceID     string `json:"instance_id,omitempty"`
	RedirectURL    string `json:"redirect_url,omitempty"` // For PayPal redirect flow
}

// PayPalCreateOrderRequest represents a request to create a PayPal order
type PayPalCreateOrderRequest struct {
	InvoiceID string `json:"invoice_id" binding:"required"`
	ReturnURL string `json:"return_url" binding:"required"`
	CancelURL string `json:"cancel_url" binding:"required"`
}

// PayPalCreateOrderResponse represents the response from creating a PayPal order
type PayPalCreateOrderResponse struct {
	OrderID     string `json:"order_id"`
	RedirectURL string `json:"redirect_url"`
}

// PayPalCaptureOrderRequest represents a request to capture a PayPal order
type PayPalCaptureOrderRequest struct {
	OrderID string `json:"order_id" binding:"required"`
}

// PayPalCaptureOrderResponse represents the response from capturing a PayPal order
type PayPalCaptureOrderResponse struct {
	OrderID       string `json:"order_id"`
	Status        string `json:"status"`
	CaptureID     string `json:"capture_id"`
	InvoiceID     string `json:"invoice_id"`
	PayerEmail    string `json:"payer_email,omitempty"`
	PayerName     string `json:"payer_name,omitempty"`
	PaymentAmount string `json:"payment_amount"`
	Currency      string `json:"currency"`
}

// PayPalWebhookEvent represents a PayPal webhook event
type PayPalWebhookEvent struct {
	ID              string                 `json:"id"`
	EventType       string                 `json:"event_type"`
	ResourceType    string                 `json:"resource_type"`
	Summary         string                 `json:"summary"`
	Resource        map[string]interface{} `json:"resource"`
	EventVersion    string                 `json:"event_version"`
	ResourceVersion string                 `json:"resource_version"`
	CreateTime      time.Time              `json:"create_time"`
}

// FlutterwavePaymentRequest represents a request to create a payment using Flutterwave
type FlutterwavePaymentRequest struct {
	InvoiceID   string `json:"invoice_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// FlutterwavePaymentResponse represents a response from creating a payment using Flutterwave
type FlutterwavePaymentResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	PaymentLink string `json:"payment_link"`
	TxRef       string `json:"tx_ref"`
}

// User represents a user in the system
type User struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	StripeCustomerID string    `json:"stripe_customer_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// StripeCheckoutRequest represents a request to create a checkout session
type StripeCheckoutRequest struct {
	InvoiceID string `json:"invoice_id"`
}

// StripeCheckoutResponse represents a response from creating a checkout session
type StripeCheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
}

// StripeSubscriptionRequest represents a request to create a subscription
type StripeSubscriptionRequest struct {
	PriceID string `json:"price_id"`
}

// StripeCancelSubscriptionRequest represents a request to cancel a subscription
type StripeCancelSubscriptionRequest struct {
	CancelAtPeriodEnd bool `json:"cancel_at_period_end"`
}

// MPesaSTKPushRequest represents a request to initiate an STK push
type MPesaSTKPushRequest struct {
	InvoiceID   string `json:"invoice_id"`
	PhoneNumber string `json:"phone_number"`
}

// MPesaSTKPushResponse represents a response from initiating an STK push
type MPesaSTKPushResponse struct {
	MerchantRequestID   string `json:"merchant_request_id"`
	CheckoutRequestID   string `json:"checkout_request_id"`
	ResponseCode        string `json:"response_code"`
	ResponseDescription string `json:"response_description"`
	CustomerMessage     string `json:"customer_message"`
}

// MPesaSTKPushStatusRequest represents a request to check the status of an STK push
type MPesaSTKPushStatusRequest struct {
	CheckoutRequestID string `json:"checkout_request_id"`
}

// MPesaSTKPushStatusResponse represents a response from checking the status of an STK push
type MPesaSTKPushStatusResponse struct {
	ResponseCode        string `json:"response_code"`
	ResponseDescription string `json:"response_description"`
	MerchantRequestID   string `json:"merchant_request_id"`
	CheckoutRequestID   string `json:"checkout_request_id"`
	CustomerMessage     string `json:"customer_message"`
}
