package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lineserve/lineserve-api/pkg/models"
)

// SupabaseClient represents a Supabase client
type SupabaseClient struct {
	ProjectURL string
	APIKey     string
	HTTPClient *http.Client
}

// NewSupabaseClient creates a new Supabase client
func NewSupabaseClient() (*SupabaseClient, error) {
	projectURL := os.Getenv("SUPABASE_URL")
	if projectURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL environment variable not set")
	}

	apiKey := os.Getenv("SUPABASE_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SUPABASE_KEY environment variable not set")
	}

	// Ensure the URL ends with a slash
	if !strings.HasSuffix(projectURL, "/") {
		projectURL += "/"
	}

	return &SupabaseClient{
		ProjectURL: projectURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// GetVPSPlans gets all VPS plans from Supabase
func (c *SupabaseClient) GetVPSPlans() ([]models.VPSPlan, error) {
	req, err := http.NewRequest("GET", c.ProjectURL+"rest/v1/vps_plans?select=*", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var plans []models.VPSPlan
	if err := json.NewDecoder(resp.Body).Decode(&plans); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return plans, nil
}

// GetVPSPlanByCode gets a VPS plan by its code from Supabase
func (c *SupabaseClient) GetVPSPlanByCode(planCode string) (*models.VPSPlan, error) {
	req, err := http.NewRequest("GET", c.ProjectURL+"rest/v1/vps_plans?plan_code=eq."+planCode, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var plans []models.VPSPlan
	if err := json.NewDecoder(resp.Body).Decode(&plans); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(plans) == 0 {
		return nil, fmt.Errorf("plan not found: %s", planCode)
	}

	return &plans[0], nil
}

// CreateVPSSubscription creates a new VPS subscription in Supabase
func (c *SupabaseClient) CreateVPSSubscription(subscription *models.VPSSubscription) (*models.VPSSubscription, error) {
	payload, err := json.Marshal(subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription: %v", err)
	}

	req, err := http.NewRequest("POST", c.ProjectURL+"rest/v1/vps_subscriptions", strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var createdSubscriptions []models.VPSSubscription
	if err := json.NewDecoder(resp.Body).Decode(&createdSubscriptions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(createdSubscriptions) == 0 {
		return nil, fmt.Errorf("no subscription created")
	}

	return &createdSubscriptions[0], nil
}

// GetVPSSubscriptionsByUserID gets all VPS subscriptions for a user from Supabase
func (c *SupabaseClient) GetVPSSubscriptionsByUserID(userID string) ([]models.VPSSubscription, error) {
	req, err := http.NewRequest("GET", c.ProjectURL+"rest/v1/vps_subscriptions?user_id=eq."+userID+"&select=*,plan:plan_id(*)", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var subscriptions []models.VPSSubscription
	if err := json.NewDecoder(resp.Body).Decode(&subscriptions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return subscriptions, nil
}

// GetVPSSubscriptionByID gets a VPS subscription by ID from Supabase
func (c *SupabaseClient) GetVPSSubscriptionByID(id string) (*models.VPSSubscription, error) {
	req, err := http.NewRequest("GET", c.ProjectURL+"rest/v1/vps_subscriptions?id=eq."+id+"&select=*,plan:plan_id(*)", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var subscriptions []models.VPSSubscription
	if err := json.NewDecoder(resp.Body).Decode(&subscriptions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(subscriptions) == 0 {
		return nil, fmt.Errorf("subscription not found: %s", id)
	}

	return &subscriptions[0], nil
}

// UpdateVPSSubscription updates a VPS subscription in Supabase
func (c *SupabaseClient) UpdateVPSSubscription(id string, updates map[string]interface{}) (*models.VPSSubscription, error) {
	payload, err := json.Marshal(updates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updates: %v", err)
	}

	req, err := http.NewRequest("PATCH", c.ProjectURL+"rest/v1/vps_subscriptions?id=eq."+id, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var updatedSubscriptions []models.VPSSubscription
	if err := json.NewDecoder(resp.Body).Decode(&updatedSubscriptions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(updatedSubscriptions) == 0 {
		return nil, fmt.Errorf("no subscription updated")
	}

	return &updatedSubscriptions[0], nil
}

// RunVPSRenewalBilling runs the VPS renewal billing process in Supabase
func (c *SupabaseClient) RunVPSRenewalBilling() ([]models.VPSRenewalResult, error) {
	req, err := http.NewRequest("POST", c.ProjectURL+"rest/v1/rpc/process_vps_renewals", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var results []models.VPSRenewalResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return results, nil
}

// CreateVPSInvoice creates a new VPS invoice in Supabase
func (c *SupabaseClient) CreateVPSInvoice(invoice *models.VPSInvoice) (*models.VPSInvoice, error) {
	payload, err := json.Marshal(invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice: %v", err)
	}

	req, err := http.NewRequest("POST", c.ProjectURL+"rest/v1/vps_invoices", strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var createdInvoices []models.VPSInvoice
	if err := json.NewDecoder(resp.Body).Decode(&createdInvoices); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(createdInvoices) == 0 {
		return nil, fmt.Errorf("no invoice created")
	}

	return &createdInvoices[0], nil
}

// GetVPSInvoiceByID gets a VPS invoice by ID from Supabase
func (c *SupabaseClient) GetVPSInvoiceByID(id string) (*models.VPSInvoice, error) {
	req, err := http.NewRequest("GET", c.ProjectURL+"rest/v1/vps_invoices?id=eq."+id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var invoices []models.VPSInvoice
	if err := json.NewDecoder(resp.Body).Decode(&invoices); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(invoices) == 0 {
		return nil, fmt.Errorf("invoice not found: %s", id)
	}

	return &invoices[0], nil
}

// GetVPSInvoicesByUserID gets all VPS invoices for a user from Supabase
func (c *SupabaseClient) GetVPSInvoicesByUserID(userID string) ([]models.VPSInvoice, error) {
	req, err := http.NewRequest("GET", c.ProjectURL+"rest/v1/vps_invoices?user_id=eq."+userID+"&order=created_at.desc", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var invoices []models.VPSInvoice
	if err := json.NewDecoder(resp.Body).Decode(&invoices); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return invoices, nil
}

// UpdateVPSInvoice updates a VPS invoice in Supabase
func (c *SupabaseClient) UpdateVPSInvoice(id string, updates map[string]interface{}) (*models.VPSInvoice, error) {
	payload, err := json.Marshal(updates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updates: %v", err)
	}

	req, err := http.NewRequest("PATCH", c.ProjectURL+"rest/v1/vps_invoices?id=eq."+id, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var updatedInvoices []models.VPSInvoice
	if err := json.NewDecoder(resp.Body).Decode(&updatedInvoices); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(updatedInvoices) == 0 {
		return nil, fmt.Errorf("no invoice updated")
	}

	return &updatedInvoices[0], nil
}

// GetUserByID gets a user by ID
func (c *SupabaseClient) GetUserByID(userID string) (*models.User, error) {
	url := fmt.Sprintf("%s/rest/v1/users?id=eq.%s&select=*", c.ProjectURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user: status code %d", resp.StatusCode)
	}

	var users []*models.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return users[0], nil
}

// UpdateUserStripeCustomerID updates a user's Stripe customer ID
func (c *SupabaseClient) UpdateUserStripeCustomerID(userID, stripeCustomerID string) error {
	url := fmt.Sprintf("%s/rest/v1/users?id=eq.%s", c.ProjectURL, userID)

	data := map[string]string{
		"stripe_customer_id": stripeCustomerID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to update user: status code %d", resp.StatusCode)
	}

	return nil
}

// GetVPSInvoiceByStripeSessionID gets a VPS invoice by Stripe session ID
func (c *SupabaseClient) GetVPSInvoiceByStripeSessionID(sessionID string) (*models.VPSInvoice, error) {
	url := fmt.Sprintf("%s/rest/v1/vps_invoices?stripe_session_id=eq.%s&select=*", c.ProjectURL, sessionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get invoice: status code %d", resp.StatusCode)
	}

	var invoices []*models.VPSInvoice
	if err := json.NewDecoder(resp.Body).Decode(&invoices); err != nil {
		return nil, err
	}

	if len(invoices) == 0 {
		return nil, fmt.Errorf("invoice not found")
	}

	return invoices[0], nil
}

// GetVPSSubscriptionByStripeID gets a VPS subscription by Stripe subscription ID
func (c *SupabaseClient) GetVPSSubscriptionByStripeID(stripeID string) (*models.VPSSubscription, error) {
	url := fmt.Sprintf("%s/rest/v1/vps_subscriptions?stripe_subscription_id=eq.%s&select=*", c.ProjectURL, stripeID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get subscription: status code %d", resp.StatusCode)
	}

	var subscriptions []*models.VPSSubscription
	if err := json.NewDecoder(resp.Body).Decode(&subscriptions); err != nil {
		return nil, err
	}

	if len(subscriptions) == 0 {
		return nil, fmt.Errorf("subscription not found")
	}

	return subscriptions[0], nil
}

// GetVPSInvoiceByMPesaCheckoutRequestID gets a VPS invoice by M-Pesa checkout request ID
func (c *SupabaseClient) GetVPSInvoiceByMPesaCheckoutRequestID(checkoutRequestID string) (*models.VPSInvoice, error) {
	url := fmt.Sprintf("%s/rest/v1/vps_invoices?mpesa_checkout_request_id=eq.%s&select=*", c.ProjectURL, checkoutRequestID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get VPS invoice: status code %d", resp.StatusCode)
	}

	var invoices []*models.VPSInvoice
	if err := json.NewDecoder(resp.Body).Decode(&invoices); err != nil {
		return nil, err
	}

	if len(invoices) == 0 {
		return nil, fmt.Errorf("VPS invoice not found")
	}

	return invoices[0], nil
}
