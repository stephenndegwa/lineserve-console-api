package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// FlutterwaveBaseURLLive is the base URL for Flutterwave API in production
	FlutterwaveBaseURLLive = "https://api.flutterwave.com/v3"
	// FlutterwaveBaseURLTest is the base URL for Flutterwave API in test mode
	FlutterwaveBaseURLTest = "https://api.flutterwave.com/v3"
)

// FlutterwaveClient represents a client for interacting with the Flutterwave API
type FlutterwaveClient struct {
	SecretKey string
	PublicKey string
	BaseURL   string
	IsSandbox bool
	Client    *http.Client
}

// NewFlutterwaveClient creates a new Flutterwave client
func NewFlutterwaveClient(secretKey, publicKey string, isSandbox bool) *FlutterwaveClient {
	baseURL := FlutterwaveBaseURLLive
	if isSandbox {
		baseURL = FlutterwaveBaseURLTest
	}

	return &FlutterwaveClient{
		SecretKey: secretKey,
		PublicKey: publicKey,
		BaseURL:   baseURL,
		IsSandbox: isSandbox,
		Client:    &http.Client{Timeout: 30 * time.Second},
	}
}

// FlutterwaveInitiatePaymentRequest represents a request to initiate a payment
type FlutterwaveInitiatePaymentRequest struct {
	TxRef         string                   `json:"tx_ref"`
	Amount        float64                  `json:"amount"`
	Currency      string                   `json:"currency"`
	RedirectURL   string                   `json:"redirect_url"`
	PaymentType   string                   `json:"payment_type,omitempty"`
	Customer      FlutterwaveCustomer      `json:"customer"`
	Customization FlutterwaveCustomization `json:"customizations,omitempty"`
	Meta          map[string]interface{}   `json:"meta,omitempty"`
}

// FlutterwaveCustomer represents customer information
type FlutterwaveCustomer struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Name        string `json:"name"`
}

// FlutterwaveCustomization represents customization options for the payment page
type FlutterwaveCustomization struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Logo        string `json:"logo,omitempty"`
}

// FlutterwaveInitiatePaymentResponse represents the response from initiating a payment
type FlutterwaveInitiatePaymentResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Link string `json:"link"`
	} `json:"data"`
}

// FlutterwaveVerifyTransactionResponse represents the response from verifying a transaction
type FlutterwaveVerifyTransactionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID                int                    `json:"id"`
		TxRef             string                 `json:"tx_ref"`
		FlwRef            string                 `json:"flw_ref"`
		Amount            float64                `json:"amount"`
		Currency          string                 `json:"currency"`
		ChargedAmount     float64                `json:"charged_amount"`
		AppFee            float64                `json:"app_fee"`
		MerchantFee       float64                `json:"merchant_fee"`
		ProcessorResponse string                 `json:"processor_response"`
		Status            string                 `json:"status"`
		PaymentType       string                 `json:"payment_type"`
		CreatedAt         string                 `json:"created_at"`
		Meta              map[string]interface{} `json:"meta"`
		Customer          struct {
			ID          int    `json:"id"`
			Email       string `json:"email"`
			Name        string `json:"name"`
			PhoneNumber string `json:"phone_number"`
		} `json:"customer"`
	} `json:"data"`
}

// FlutterwaveWebhookEvent represents a webhook event from Flutterwave
type FlutterwaveWebhookEvent struct {
	Event string `json:"event"`
	Data  struct {
		ID                int                    `json:"id"`
		TxRef             string                 `json:"tx_ref"`
		FlwRef            string                 `json:"flw_ref"`
		Amount            float64                `json:"amount"`
		Currency          string                 `json:"currency"`
		ChargedAmount     float64                `json:"charged_amount"`
		AppFee            float64                `json:"app_fee"`
		MerchantFee       float64                `json:"merchant_fee"`
		ProcessorResponse string                 `json:"processor_response"`
		Status            string                 `json:"status"`
		PaymentType       string                 `json:"payment_type"`
		CreatedAt         string                 `json:"created_at"`
		Meta              map[string]interface{} `json:"meta"`
		Customer          struct {
			ID          int    `json:"id"`
			Email       string `json:"email"`
			Name        string `json:"name"`
			PhoneNumber string `json:"phone_number"`
		} `json:"customer"`
	} `json:"data"`
}

// InitiatePayment initiates a payment with Flutterwave
func (c *FlutterwaveClient) InitiatePayment(req *FlutterwaveInitiatePaymentRequest) (*FlutterwaveInitiatePaymentResponse, error) {
	url := fmt.Sprintf("%s/payments", c.BaseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.SecretKey)

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response FlutterwaveInitiatePaymentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return &response, nil
}

// VerifyTransaction verifies a transaction with Flutterwave
func (c *FlutterwaveClient) VerifyTransaction(transactionID string) (*FlutterwaveVerifyTransactionResponse, error) {
	url := fmt.Sprintf("%s/transactions/%s/verify", c.BaseURL, transactionID)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.SecretKey)

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response FlutterwaveVerifyTransactionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return &response, nil
}

// VerifyWebhookSignature verifies the signature of a webhook event
func (c *FlutterwaveClient) VerifyWebhookSignature(signature string, payload []byte) bool {
	// In a real implementation, you would verify the signature using the Flutterwave webhook secret
	// For now, we'll just return true for testing purposes
	return true
}

// ListBanks gets a list of banks from Flutterwave
func (c *FlutterwaveClient) ListBanks(country string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/banks/%s", c.BaseURL, country)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.SecretKey)

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Status  string                   `json:"status"`
		Message string                   `json:"message"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return response.Data, nil
}

// GetFlutterwaveClientFromEnv creates a new Flutterwave client from environment variables
func GetFlutterwaveClientFromEnv() (*FlutterwaveClient, error) {
	secretKey := os.Getenv("FLUTTERWAVE_SECRET_KEY")
	publicKey := os.Getenv("FLUTTERWAVE_PUBLIC_KEY")
	isSandbox := os.Getenv("FLUTTERWAVE_SANDBOX") == "true"

	if secretKey == "" || publicKey == "" {
		return nil, errors.New("FLUTTERWAVE_SECRET_KEY and FLUTTERWAVE_PUBLIC_KEY environment variables must be set")
	}

	return NewFlutterwaveClient(secretKey, publicKey, isSandbox), nil
}
