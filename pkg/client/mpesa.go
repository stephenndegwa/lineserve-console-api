package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// MPesaBaseURLLive is the base URL for M-Pesa API in production
	MPesaBaseURLLive = "https://api.safaricom.co.ke"
	// MPesaBaseURLSandbox is the base URL for M-Pesa API in sandbox
	MPesaBaseURLSandbox = "https://sandbox.safaricom.co.ke"
)

// MPesaClient represents a client for interacting with the M-Pesa API
type MPesaClient struct {
	ConsumerKey       string
	ConsumerSecret    string
	BaseURL           string
	IsSandbox         bool
	Client            *http.Client
	AccessToken       string
	TokenExpiry       time.Time
	BusinessShortCode string
	PassKey           string
}

// MPesaResponse represents a generic M-Pesa API response
type MPesaResponse struct {
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	MerchantRequestID   string `json:"MerchantRequestID,omitempty"`
	CheckoutRequestID   string `json:"CheckoutRequestID,omitempty"`
	CustomerMessage     string `json:"CustomerMessage,omitempty"`
}

// STKPushRequest represents an STK push request
type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            string `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

// STKPushResponse represents an STK push response
type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

// STKPushCallback represents an STK push callback
type STKPushCallback struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value,omitempty"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// TransactionStatusRequest represents a transaction status request
type TransactionStatusRequest struct {
	Initiator          string `json:"Initiator"`
	SecurityCredential string `json:"SecurityCredential"`
	CommandID          string `json:"CommandID"`
	TransactionID      string `json:"TransactionID"`
	PartyA             string `json:"PartyA"`
	IdentifierType     string `json:"IdentifierType"`
	ResultURL          string `json:"ResultURL"`
	QueueTimeOutURL    string `json:"QueueTimeOutURL"`
	Remarks            string `json:"Remarks"`
	Occasion           string `json:"Occasion"`
}

// TransactionStatusResponse represents a transaction status response
type TransactionStatusResponse struct {
	OriginatorConversationID string `json:"OriginatorConversationID"`
	ConversationID           string `json:"ConversationID"`
	ResponseCode             string `json:"ResponseCode"`
	ResponseDescription      string `json:"ResponseDescription"`
}

// C2BRegisterURLRequest represents a C2B register URL request
type C2BRegisterURLRequest struct {
	ShortCode       string `json:"ShortCode"`
	ResponseType    string `json:"ResponseType"`
	ConfirmationURL string `json:"ConfirmationURL"`
	ValidationURL   string `json:"ValidationURL"`
}

// C2BRegisterURLResponse represents a C2B register URL response
type C2BRegisterURLResponse struct {
	OriginatorConversationID string `json:"OriginatorConversationID"`
	ConversationID           string `json:"ConversationID"`
	ResponseDescription      string `json:"ResponseDescription"`
}

// C2BSimulateRequest represents a C2B simulate request
type C2BSimulateRequest struct {
	ShortCode     string `json:"ShortCode"`
	CommandID     string `json:"CommandID"`
	Amount        string `json:"Amount"`
	Msisdn        string `json:"Msisdn"`
	BillRefNumber string `json:"BillRefNumber"`
}

// C2BSimulateResponse represents a C2B simulate response
type C2BSimulateResponse struct {
	OriginatorConversationID string `json:"OriginatorConversationID"`
	ConversationID           string `json:"ConversationID"`
	ResponseDescription      string `json:"ResponseDescription"`
}

// NewMPesaClient creates a new M-Pesa client
func NewMPesaClient(consumerKey, consumerSecret, businessShortCode, passKey string, isSandbox bool) *MPesaClient {
	baseURL := MPesaBaseURLLive
	if isSandbox {
		baseURL = MPesaBaseURLSandbox
	}

	return &MPesaClient{
		ConsumerKey:       consumerKey,
		ConsumerSecret:    consumerSecret,
		BusinessShortCode: businessShortCode,
		PassKey:           passKey,
		BaseURL:           baseURL,
		IsSandbox:         isSandbox,
		Client:            &http.Client{Timeout: 30 * time.Second},
	}
}

// GetMPesaClientFromEnv creates a new M-Pesa client from environment variables
func GetMPesaClientFromEnv() (*MPesaClient, error) {
	consumerKey := os.Getenv("MPESA_CONSUMER_KEY")
	consumerSecret := os.Getenv("MPESA_CONSUMER_SECRET")
	businessShortCode := os.Getenv("MPESA_BUSINESS_SHORTCODE")
	passKey := os.Getenv("MPESA_PASS_KEY")
	sandboxStr := os.Getenv("MPESA_SANDBOX")

	if consumerKey == "" || consumerSecret == "" || businessShortCode == "" || passKey == "" {
		return nil, errors.New("MPESA_CONSUMER_KEY, MPESA_CONSUMER_SECRET, MPESA_BUSINESS_SHORTCODE, and MPESA_PASS_KEY must be set")
	}

	isSandbox := sandboxStr == "true"
	return NewMPesaClient(consumerKey, consumerSecret, businessShortCode, passKey, isSandbox), nil
}

// Authenticate authenticates with the M-Pesa API and gets an access token
func (c *MPesaClient) Authenticate() error {
	// Check if we already have a valid token
	if c.AccessToken != "" && time.Now().Before(c.TokenExpiry) {
		return nil
	}

	url := fmt.Sprintf("%s/oauth/v1/generate?grant_type=client_credentials", c.BaseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.ConsumerKey + ":" + c.ConsumerSecret))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: %s", string(body))
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	c.AccessToken = authResp.AccessToken
	expiresIn := 3600 // Default to 1 hour if parsing fails
	if authResp.ExpiresIn != "" {
		fmt.Sscanf(authResp.ExpiresIn, "%d", &expiresIn)
	}
	c.TokenExpiry = time.Now().Add(time.Duration(expiresIn-60) * time.Second) // Subtract 60 seconds for safety

	return nil
}

// GeneratePassword generates the password for STK push
func (c *MPesaClient) GeneratePassword(shortCode, passkey, timestamp string) string {
	str := shortCode + passkey + timestamp
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// STKPush initiates an STK push request
func (c *MPesaClient) STKPush(req STKPushRequest) (*STKPushResponse, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/mpesa/stkpush/v1/processrequest", c.BaseURL)
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add("Authorization", "Bearer "+c.AccessToken)
	httpReq.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("STK push failed: %s", string(body))
	}

	var stkResp STKPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&stkResp); err != nil {
		return nil, err
	}

	return &stkResp, nil
}

// QuerySTKPushStatus queries the status of an STK push transaction
func (c *MPesaClient) QuerySTKPushStatus(businessShortCode, password, timestamp, checkoutRequestID string) (*MPesaResponse, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/mpesa/stkpushquery/v1/query", c.BaseURL)
	reqBody := map[string]string{
		"BusinessShortCode": businessShortCode,
		"Password":          password,
		"Timestamp":         timestamp,
		"CheckoutRequestID": checkoutRequestID,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add("Authorization", "Bearer "+c.AccessToken)
	httpReq.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("STK push query failed: %s", string(body))
	}

	var queryResp MPesaResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, err
	}

	return &queryResp, nil
}

// RegisterC2BURL registers the C2B URLs
func (c *MPesaClient) RegisterC2BURL(req C2BRegisterURLRequest) (*C2BRegisterURLResponse, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/mpesa/c2b/v1/registerurl", c.BaseURL)
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add("Authorization", "Bearer "+c.AccessToken)
	httpReq.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("C2B URL registration failed: %s", string(body))
	}

	var registerResp C2BRegisterURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		return nil, err
	}

	return &registerResp, nil
}

// SimulateC2B simulates a C2B transaction (only works in sandbox)
func (c *MPesaClient) SimulateC2B(req C2BSimulateRequest) (*C2BSimulateResponse, error) {
	if !c.IsSandbox {
		return nil, errors.New("C2B simulation is only available in sandbox environment")
	}

	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/mpesa/c2b/v1/simulate", c.BaseURL)
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add("Authorization", "Bearer "+c.AccessToken)
	httpReq.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("C2B simulation failed: %s", string(body))
	}

	var simulateResp C2BSimulateResponse
	if err := json.NewDecoder(resp.Body).Decode(&simulateResp); err != nil {
		return nil, err
	}

	return &simulateResp, nil
}

// TransactionStatus checks the status of a transaction
func (c *MPesaClient) TransactionStatus(req TransactionStatusRequest) (*TransactionStatusResponse, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/mpesa/transactionstatus/v1/query", c.BaseURL)
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add("Authorization", "Bearer "+c.AccessToken)
	httpReq.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("transaction status check failed: %s", string(body))
	}

	var statusResp TransactionStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, err
	}

	return &statusResp, nil
}

// GetBusinessShortCode returns the business short code
func (c *MPesaClient) GetBusinessShortCode() string {
	return c.BusinessShortCode
}

// GetPassKey returns the pass key
func (c *MPesaClient) GetPassKey() string {
	return c.PassKey
}
