package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// FlutterwaveHandler handles Flutterwave-related requests
type FlutterwaveHandler struct {
	SupabaseClient    *client.SupabaseClient
	FlutterwaveClient *client.FlutterwaveClient
}

// NewFlutterwaveHandler creates a new Flutterwave handler
func NewFlutterwaveHandler(supabaseClient *client.SupabaseClient, flutterwaveClient *client.FlutterwaveClient) *FlutterwaveHandler {
	return &FlutterwaveHandler{
		SupabaseClient:    supabaseClient,
		FlutterwaveClient: flutterwaveClient,
	}
}

// CreatePayment creates a new payment using Flutterwave
func (h *FlutterwaveHandler) CreatePayment(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Parse request body
	var req models.FlutterwavePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Validate request
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

	// Generate a unique transaction reference
	txRef := fmt.Sprintf("LSFW-%s-%s", invoice.ID[:8], uuid.New().String()[:8])

	// Create Flutterwave payment request
	baseURL := c.BaseURL()
	if baseURL == "" {
		baseURL = "https://lineserve.net" // Default URL if not available from context
	}

	redirectURL := fmt.Sprintf("%s/payment/flutterwave/callback", baseURL)

	flutterwaveReq := &client.FlutterwaveInitiatePaymentRequest{
		TxRef:       txRef,
		Amount:      invoice.Amount,
		Currency:    invoice.Currency,
		RedirectURL: redirectURL,
		Customer: client.FlutterwaveCustomer{
			Email:       req.Email,
			Name:        req.Name,
			PhoneNumber: req.PhoneNumber,
		},
		Customization: client.FlutterwaveCustomization{
			Title:       "LineServe VPS Payment",
			Description: fmt.Sprintf("Payment for VPS Plan: %s", invoice.PlanCode),
			Logo:        "https://lineserve.net/logo.png",
		},
		Meta: map[string]interface{}{
			"invoice_id": invoice.ID,
			"user_id":    userID,
		},
	}

	// Initiate payment with Flutterwave
	response, err := h.FlutterwaveClient.InitiatePayment(flutterwaveReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to initiate payment: %v", err),
		})
	}

	// Update invoice with transaction reference
	updates := map[string]interface{}{
		"payment_method": "flutterwave",
		"tx_ref":         txRef,
	}
	_, err = h.SupabaseClient.UpdateVPSInvoice(invoice.ID, updates)
	if err != nil {
		// Log the error but continue
		fmt.Printf("Failed to update invoice with transaction reference: %v\n", err)
	}

	// Return payment link
	return c.JSON(models.FlutterwavePaymentResponse{
		Status:      "success",
		Message:     "Payment initiated successfully",
		PaymentLink: response.Data.Link,
		TxRef:       txRef,
	})
}

// HandleWebhook handles Flutterwave webhook events
func (h *FlutterwaveHandler) HandleWebhook(c *fiber.Ctx) error {
	// Get signature from header
	signature := c.Get("verif-hash")
	if signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing signature",
		})
	}

	// Read request body
	body := c.Body()

	// Verify webhook signature
	if !h.FlutterwaveClient.VerifyWebhookSignature(signature, body) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid signature",
		})
	}

	// Parse webhook event
	var event client.FlutterwaveWebhookEvent
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid webhook payload: %v", err),
		})
	}

	// Process webhook event
	if event.Event == "charge.completed" && event.Data.Status == "successful" {
		// Extract invoice ID from meta data
		metaInvoiceID, ok := event.Data.Meta["invoice_id"].(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing invoice ID in meta data",
			})
		}

		// Get invoice from Supabase
		invoice, err := h.SupabaseClient.GetVPSInvoiceByID(metaInvoiceID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("Invoice not found: %v", err),
			})
		}

		// Update invoice status to paid
		now := time.Now()
		invoiceUpdates := map[string]interface{}{
			"status":            "paid",
			"payment_intent_id": fmt.Sprintf("fw_%d", event.Data.ID),
			"paid_at":           now,
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
	}

	return c.SendStatus(fiber.StatusOK)
}

// VerifyPayment verifies a Flutterwave payment
func (h *FlutterwaveHandler) VerifyPayment(c *fiber.Ctx) error {
	// Get transaction ID from URL
	transactionID := c.Params("id")
	if transactionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Transaction ID is required",
		})
	}

	// Verify transaction with Flutterwave
	response, err := h.FlutterwaveClient.VerifyTransaction(transactionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to verify transaction: %v", err),
		})
	}

	// Check if transaction was successful
	if response.Data.Status != "successful" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Transaction was not successful",
			"status": response.Data.Status,
		})
	}

	// Extract invoice ID from meta data
	metaInvoiceID, ok := response.Data.Meta["invoice_id"].(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing invoice ID in meta data",
		})
	}

	// Get invoice from Supabase
	invoice, err := h.SupabaseClient.GetVPSInvoiceByID(metaInvoiceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Invoice not found: %v", err),
		})
	}

	// Update invoice status to paid if not already
	if invoice.Status != "paid" {
		now := time.Now()
		invoiceUpdates := map[string]interface{}{
			"status":            "paid",
			"payment_intent_id": fmt.Sprintf("fw_%d", response.Data.ID),
			"paid_at":           now,
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
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Payment verified successfully",
		"data": fiber.Map{
			"invoice_id":      invoice.ID,
			"subscription_id": invoice.SubscriptionID,
			"amount":          response.Data.Amount,
			"currency":        response.Data.Currency,
			"status":          response.Data.Status,
		},
	})
}

// GetPaymentStatus gets the status of a Flutterwave payment by transaction reference
func (h *FlutterwaveHandler) GetPaymentStatus(c *fiber.Ctx) error {
	// Get transaction reference from URL
	txRef := c.Params("tx_ref")
	if txRef == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Transaction reference is required",
		})
	}

	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get all invoices for the user
	invoices, err := h.SupabaseClient.GetVPSInvoicesByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get invoices: %v", err),
		})
	}

	// Find the invoice with the matching transaction reference
	var invoice *models.VPSInvoice
	for _, inv := range invoices {
		if inv.TxRef == txRef {
			invoice = &inv
			break
		}
	}

	if invoice == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Invoice not found for the given transaction reference",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Payment status retrieved successfully",
		"data": fiber.Map{
			"invoice_id":      invoice.ID,
			"subscription_id": invoice.SubscriptionID,
			"amount":          invoice.Amount,
			"currency":        invoice.Currency,
			"status":          invoice.Status,
			"paid_at":         invoice.PaidAt,
		},
	})
}
