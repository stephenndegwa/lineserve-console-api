package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/lineserve/lineserve-api/pkg/models"
)

// PayPalClient handles interactions with the PayPal API
type PayPalClient struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	HTTPClient   *http.Client
	IsSandbox    bool
}

// PayPalTokenResponse represents the response from a PayPal OAuth token request
type PayPalTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// PayPalOrder represents a PayPal order
type PayPalOrder struct {
	ID     string       `json:"id"`
	Status string       `json:"status"`
	Links  []PayPalLink `json:"links"`
}

// PayPalLink represents a HATEOAS link in PayPal responses
type PayPalLink struct {
	Href   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

// PayPalCaptureResponse represents the response from a PayPal capture request
type PayPalCaptureResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PurchaseUnits []struct {
		ReferenceID string `json:"reference_id"`
		Payments    struct {
			Captures []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Amount struct {
					Value    string `json:"value"`
					Currency string `json:"currency_code"`
				} `json:"amount"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
	Payer struct {
		Name struct {
			GivenName string `json:"given_name"`
			Surname   string `json:"surname"`
		} `json:"name"`
		Email string `json:"email_address"`
	} `json:"payer,omitempty"`
}

// NewPayPalClient creates a new PayPal client
func NewPayPalClient() (*PayPalClient, error) {
	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	clientSecret := os.Getenv("PAYPAL_CLIENT_SECRET")
	isSandbox := os.Getenv("PAYPAL_SANDBOX") != "false"

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("PayPal client ID or secret not set")
	}

	baseURL := "https://api-m.paypal.com"
	if isSandbox {
		baseURL = "https://api-m.sandbox.paypal.com"
	}

	return &PayPalClient{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		BaseURL:      baseURL,
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
		IsSandbox:    isSandbox,
	}, nil
}

// GetAccessToken gets an OAuth access token from PayPal
func (c *PayPalClient) GetAccessToken() (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(c.ClientID + ":" + c.ClientSecret))

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/oauth2/token", bytes.NewBuffer([]byte("grant_type=client_credentials")))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get access token: %s, status: %d", string(body), resp.StatusCode)
	}

	var tokenResp PayPalTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// CreateOrder creates a PayPal order for a VPS invoice
func (c *PayPalClient) CreateOrder(invoice *models.VPSInvoice, returnURL, cancelURL string) (*models.PayPalCreateOrderResponse, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	// Format the amount with 2 decimal places
	amountStr := fmt.Sprintf("%.2f", invoice.Amount)

	// Create the order request body
	orderRequest := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"reference_id": invoice.ID,
				"description":  fmt.Sprintf("Lineserve VPS Plan: %s (%d months)", invoice.PlanCode, invoice.PeriodMonths),
				"amount": map[string]interface{}{
					"currency_code": invoice.Currency,
					"value":         amountStr,
				},
			},
		},
		"application_context": map[string]interface{}{
			"return_url":  returnURL,
			"cancel_url":  cancelURL,
			"brand_name":  "Lineserve Cloud",
			"user_action": "PAY_NOW",
		},
	}

	orderJSON, err := json.Marshal(orderRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/v2/checkout/orders", bytes.NewBuffer(orderJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create PayPal order: %s, status: %d", string(body), resp.StatusCode)
	}

	var order PayPalOrder
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, err
	}

	// Find the approval URL
	var approvalURL string
	for _, link := range order.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}

	if approvalURL == "" {
		return nil, fmt.Errorf("approval URL not found in PayPal response")
	}

	return &models.PayPalCreateOrderResponse{
		OrderID:     order.ID,
		RedirectURL: approvalURL,
	}, nil
}

// CaptureOrder captures a PayPal order after it has been approved
func (c *PayPalClient) CaptureOrder(orderID string) (*models.PayPalCaptureOrderResponse, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/v2/checkout/orders/"+orderID+"/capture", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to capture PayPal order: %s, status: %d", string(body), resp.StatusCode)
	}

	var captureResp PayPalCaptureResponse
	if err := json.NewDecoder(resp.Body).Decode(&captureResp); err != nil {
		return nil, err
	}

	// Extract the capture ID and other details
	var captureID, paymentAmount, currency string
	if len(captureResp.PurchaseUnits) > 0 && len(captureResp.PurchaseUnits[0].Payments.Captures) > 0 {
		capture := captureResp.PurchaseUnits[0].Payments.Captures[0]
		captureID = capture.ID
		paymentAmount = capture.Amount.Value
		currency = capture.Amount.Currency
	}

	// Get the invoice ID from the reference ID
	invoiceID := ""
	if len(captureResp.PurchaseUnits) > 0 {
		invoiceID = captureResp.PurchaseUnits[0].ReferenceID
	}

	// Extract payer information if available
	payerName := ""
	payerEmail := ""
	if captureResp.Payer.Email != "" {
		payerEmail = captureResp.Payer.Email
		payerName = captureResp.Payer.Name.GivenName + " " + captureResp.Payer.Name.Surname
	}

	return &models.PayPalCaptureOrderResponse{
		OrderID:       orderID,
		Status:        captureResp.Status,
		CaptureID:     captureID,
		InvoiceID:     invoiceID,
		PayerEmail:    payerEmail,
		PayerName:     payerName,
		PaymentAmount: paymentAmount,
		Currency:      currency,
	}, nil
}

// GetOrderDetails gets the details of a PayPal order
func (c *PayPalClient) GetOrderDetails(orderID string) (*PayPalOrder, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", c.BaseURL+"/v2/checkout/orders/"+orderID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get PayPal order details: %s, status: %d", string(body), resp.StatusCode)
	}

	var order PayPalOrder
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, err
	}

	return &order, nil
}
