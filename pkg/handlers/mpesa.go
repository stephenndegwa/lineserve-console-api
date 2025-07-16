package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// MPesaHandler handles M-Pesa-related requests
type MPesaHandler struct {
	SupabaseClient *client.SupabaseClient
	MPesaClient    *client.MPesaClient
}

// NewMPesaHandler creates a new M-Pesa handler
func NewMPesaHandler(supabaseClient *client.SupabaseClient, mpesaClient *client.MPesaClient) *MPesaHandler {
	return &MPesaHandler{
		SupabaseClient: supabaseClient,
		MPesaClient:    mpesaClient,
	}
}

// InitiateSTKPush initiates an STK push request to the customer's phone
func (h *MPesaHandler) InitiateSTKPush(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Parse request body
	var req models.MPesaSTKPushRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Validate invoice ID
	if req.InvoiceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invoice ID is required",
		})
	}

	// Get invoice from Supabase
	invoice, err := h.SupabaseClient.GetVPSInvoiceByID(req.InvoiceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Invoice not found: %v", err),
		})
	}

	// Check if invoice belongs to user
	if invoice.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You do not have permission to pay this invoice",
		})
	}

	// Check if invoice is already paid
	if invoice.Status == "paid" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invoice is already paid",
		})
	}

	// Check if invoice is expired
	if invoice.Status == "expired" || invoice.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invoice is expired",
		})
	}

	// Format phone number (remove leading zero if present and add country code)
	phoneNumber := req.PhoneNumber
	if len(phoneNumber) > 0 && phoneNumber[0] == '0' {
		phoneNumber = "254" + phoneNumber[1:]
	} else if len(phoneNumber) > 0 && phoneNumber[0:3] != "254" {
		phoneNumber = "254" + phoneNumber
	}

	// Generate timestamp for M-Pesa
	timestamp := time.Now().Format("20060102150405")

	// Generate password
	shortCode := h.MPesaClient.GetBusinessShortCode()
	passKey := h.MPesaClient.GetPassKey()
	password := h.MPesaClient.GeneratePassword(shortCode, passKey, timestamp)

	// Create STK push request
	stkPushReq := client.STKPushRequest{
		BusinessShortCode: shortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            fmt.Sprintf("%.0f", invoice.Amount), // Convert to string without decimal
		PartyA:            phoneNumber,
		PartyB:            shortCode,
		PhoneNumber:       phoneNumber,
		CallBackURL:       fmt.Sprintf("%s/v1/mpesa/callback", c.BaseURL()),
		AccountReference:  invoice.ID[:8], // Use first 8 chars of invoice ID
		TransactionDesc:   fmt.Sprintf("Payment for VPS plan %s", invoice.PlanCode),
	}

	// Send STK push request
	stkResp, err := h.MPesaClient.STKPush(stkPushReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to initiate STK push: %v", err),
		})
	}

	// Update invoice with M-Pesa checkout request ID
	updates := map[string]interface{}{
		"mpesa_checkout_request_id": stkResp.CheckoutRequestID,
		"payment_method_id":         "mpesa",
	}
	_, err = h.SupabaseClient.UpdateVPSInvoice(invoice.ID, updates)
	if err != nil {
		// Log the error but continue
		fmt.Printf("Failed to update invoice with M-Pesa checkout request ID: %v\n", err)
	}

	// Return response
	return c.JSON(models.MPesaSTKPushResponse{
		MerchantRequestID:   stkResp.MerchantRequestID,
		CheckoutRequestID:   stkResp.CheckoutRequestID,
		ResponseCode:        stkResp.ResponseCode,
		ResponseDescription: stkResp.ResponseDescription,
		CustomerMessage:     stkResp.CustomerMessage,
	})
}

// HandleSTKPushCallback handles the callback from M-Pesa after STK push
func (h *MPesaHandler) HandleSTKPushCallback(c *fiber.Ctx) error {
	// Parse callback body
	var callback client.STKPushCallback
	if err := c.BodyParser(&callback); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid callback body: %v", err),
		})
	}

	// Extract result code and checkout request ID
	resultCode := callback.Body.StkCallback.ResultCode
	checkoutRequestID := callback.Body.StkCallback.CheckoutRequestID

	// Get invoice from Supabase by checkout request ID
	invoice, err := h.SupabaseClient.GetVPSInvoiceByMPesaCheckoutRequestID(checkoutRequestID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Invoice not found: %v", err),
		})
	}

	// Check if payment was successful
	if resultCode == 0 {
		// Extract payment details from callback
		var mpesaReceiptNumber string
		var phoneNumber string
		var amount float64

		for _, item := range callback.Body.StkCallback.CallbackMetadata.Item {
			switch item.Name {
			case "MpesaReceiptNumber":
				mpesaReceiptNumber = item.Value.(string)
			case "PhoneNumber":
				phoneNumber = fmt.Sprintf("%v", item.Value)
			case "Amount":
				switch v := item.Value.(type) {
				case float64:
					amount = v
				case string:
					fmt.Sscanf(v, "%f", &amount)
				}
			}
		}

		// Update invoice status to paid
		now := time.Now()
		invoiceUpdates := map[string]interface{}{
			"status":             "paid",
			"mpesa_receipt_no":   mpesaReceiptNumber,
			"mpesa_phone_number": phoneNumber,
			"paid_at":            now,
		}
		_, err = h.SupabaseClient.UpdateVPSInvoice(invoice.ID, invoiceUpdates)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to update invoice: %v", err),
			})
		}

		// Get subscription
		subscription, err := h.SupabaseClient.GetVPSSubscriptionByID(invoice.SubscriptionID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to get subscription: %v", err),
			})
		}

		// Update subscription status to active
		subscriptionUpdates := map[string]interface{}{
			"status":     "active",
			"start_date": now,
		}
		_, err = h.SupabaseClient.UpdateVPSSubscription(subscription.ID, subscriptionUpdates)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to update subscription: %v", err),
			})
		}

		// Return success
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"ResultCode": 0,
			"ResultDesc": "Accepted",
		})
	} else {
		// Update invoice status to failed
		updates := map[string]interface{}{
			"status": "failed",
		}
		_, err = h.SupabaseClient.UpdateVPSInvoice(invoice.ID, updates)
		if err != nil {
			// Log the error but continue
			fmt.Printf("Failed to update invoice status: %v\n", err)
		}

		// Return failure
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"ResultCode": 1,
			"ResultDesc": "Rejected",
		})
	}
}

// CheckSTKPushStatus checks the status of an STK push transaction
func (h *MPesaHandler) CheckSTKPushStatus(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Parse request body
	var req models.MPesaSTKPushStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Validate checkout request ID
	if req.CheckoutRequestID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Checkout request ID is required",
		})
	}

	// Get invoice from Supabase by checkout request ID
	invoice, err := h.SupabaseClient.GetVPSInvoiceByMPesaCheckoutRequestID(req.CheckoutRequestID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Invoice not found: %v", err),
		})
	}

	// Check if invoice belongs to user
	if invoice.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You do not have permission to check this transaction",
		})
	}

	// Generate timestamp for M-Pesa
	timestamp := time.Now().Format("20060102150405")

	// Generate password
	shortCode := h.MPesaClient.GetBusinessShortCode()
	passKey := h.MPesaClient.GetPassKey()
	password := h.MPesaClient.GeneratePassword(shortCode, passKey, timestamp)

	// Query STK push status
	statusResp, err := h.MPesaClient.QuerySTKPushStatus(shortCode, password, timestamp, req.CheckoutRequestID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to check STK push status: %v", err),
		})
	}

	// Return response
	return c.JSON(models.MPesaSTKPushStatusResponse{
		ResponseCode:        statusResp.ResponseCode,
		ResponseDescription: statusResp.ResponseDescription,
		MerchantRequestID:   statusResp.MerchantRequestID,
		CheckoutRequestID:   statusResp.CheckoutRequestID,
		CustomerMessage:     statusResp.CustomerMessage,
	})
}

// RegisterRoutes registers the M-Pesa routes
func (h *MPesaHandler) RegisterRoutes(app *fiber.App) {
	mpesa := app.Group("/v1/mpesa")
	mpesa.Post("/stk-push", h.InitiateSTKPush)
	mpesa.Post("/callback", h.HandleSTKPushCallback)
	mpesa.Post("/check-status", h.CheckSTKPushStatus)
}
