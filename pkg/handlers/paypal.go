package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// PayPalHandler handles PayPal-related requests
type PayPalHandler struct {
	PayPalClient   *client.PayPalClient
	SupabaseClient *client.SupabaseClient
	VPSHandler     *VPSHandler
}

// NewPayPalHandler creates a new PayPal handler
func NewPayPalHandler(paypalClient *client.PayPalClient, supabaseClient *client.SupabaseClient, vpsHandler *VPSHandler) *PayPalHandler {
	return &PayPalHandler{
		PayPalClient:   paypalClient,
		SupabaseClient: supabaseClient,
		VPSHandler:     vpsHandler,
	}
}

// CreateOrder creates a PayPal order for a VPS invoice
func (h *PayPalHandler) CreateOrder(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Parse request body
	var req models.PayPalCreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
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

	// Create PayPal order
	orderResp, err := h.PayPalClient.CreateOrder(invoice, req.ReturnURL, req.CancelURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create PayPal order: %v", err),
		})
	}

	// Update invoice with PayPal order ID
	updates := map[string]interface{}{
		"payment_method_id": "paypal",
		"payment_intent_id": orderResp.OrderID,
	}
	_, err = h.SupabaseClient.UpdateVPSInvoice(invoice.ID, updates)
	if err != nil {
		// Log the error but continue
		fmt.Printf("Failed to update invoice with PayPal order ID: %v\n", err)
	}

	return c.JSON(orderResp)
}

// CaptureOrder captures a PayPal order after it has been approved
func (h *PayPalHandler) CaptureOrder(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Parse request body
	var req models.PayPalCaptureOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Capture the order
	captureResp, err := h.PayPalClient.CaptureOrder(req.OrderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to capture PayPal order: %v", err),
		})
	}

	// Check if the capture was successful
	if captureResp.Status != "COMPLETED" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("PayPal order capture failed with status: %s", captureResp.Status),
		})
	}

	// Get the invoice
	invoice, err := h.SupabaseClient.GetVPSInvoiceByID(captureResp.InvoiceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get invoice: %v", err),
		})
	}

	// Check if invoice belongs to user
	if invoice.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You do not have permission to pay this invoice",
		})
	}

	// Update invoice status to paid
	now := time.Now()
	invoiceUpdates := map[string]interface{}{
		"status":            "paid",
		"payment_method_id": "paypal",
		"payment_intent_id": captureResp.OrderID,
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
	updatedSubscription, err := h.SupabaseClient.UpdateVPSSubscription(subscription.ID, subscriptionUpdates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update subscription: %v", err),
		})
	}

	// Provision VPS if OpenStack client is available
	var instanceID string
	if h.VPSHandler.OpenStackClient != nil && subscription.Plan != nil && subscription.Plan.OpenStackFlavorID != "" {
		// Get project ID from context
		projectID := c.Locals("project_id").(string)
		if projectID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Project ID not found",
			})
		}

		// Create instance name
		instanceName := fmt.Sprintf("vps-%s-%s", subscription.Plan.PlanCode, subscription.ID[:8])

		// TODO: Get appropriate image ID and network ID
		imageID := "default-image-id"     // This should be replaced with actual image ID
		networkID := "default-network-id" // This should be replaced with actual network ID

		// Create instance
		instance, err := h.VPSHandler.provisionInstance(projectID, instanceName, subscription.Plan.OpenStackFlavorID, imageID, networkID)
		if err != nil {
			// Update subscription with error status
			errorUpdates := map[string]interface{}{
				"status": "provisioning_failed",
			}
			_, updateErr := h.SupabaseClient.UpdateVPSSubscription(subscription.ID, errorUpdates)
			if updateErr != nil {
				// Log the error but continue
				fmt.Printf("Failed to update subscription status: %v\n", updateErr)
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to provision instance: %v", err),
			})
		}

		// Update subscription with instance ID
		instanceUpdates := map[string]interface{}{
			"instance_id":          instance.ID,
			"openstack_project_id": projectID,
		}
		updatedSubscription, err = h.SupabaseClient.UpdateVPSSubscription(subscription.ID, instanceUpdates)
		if err != nil {
			// Log the error but continue
			fmt.Printf("Failed to update subscription with instance ID: %v\n", err)
		}

		instanceID = instance.ID
	}

	// Return response
	return c.JSON(models.VPSInvoicePayResponse{
		Status:         "success",
		SubscriptionID: updatedSubscription.ID,
		InstanceID:     instanceID,
	})
}

// HandleWebhook handles PayPal webhook events
func (h *PayPalHandler) HandleWebhook(c *fiber.Ctx) error {
	// Parse webhook event
	var event models.PayPalWebhookEvent
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid webhook event: %v", err),
		})
	}

	// Handle different event types
	switch event.EventType {
	case "PAYMENT.CAPTURE.COMPLETED":
		// Extract order ID from resource
		orderID, ok := event.Resource["id"].(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid order ID in webhook event",
			})
		}

		// Get order details
		order, err := h.PayPalClient.GetOrderDetails(orderID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to get order details: %v", err),
			})
		}

		// Process the payment (similar to CaptureOrder but without user authentication)
		// This is a backup in case the client-side capture fails
		fmt.Printf("Received webhook for completed payment for order: %s with status: %s\n", orderID, order.Status)

		// For now, just acknowledge the webhook
		return c.SendStatus(fiber.StatusOK)

	case "PAYMENT.CAPTURE.DENIED":
		// Handle denied payment
		fmt.Printf("Received webhook for denied payment: %v\n", event.Resource)
		return c.SendStatus(fiber.StatusOK)

	default:
		// Log unhandled event type
		fmt.Printf("Unhandled PayPal webhook event type: %s\n", event.EventType)
		return c.SendStatus(fiber.StatusOK)
	}
}

// GetOrderStatus gets the status of a PayPal order
func (h *PayPalHandler) GetOrderStatus(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get order ID from URL
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	// Get order details from PayPal
	order, err := h.PayPalClient.GetOrderDetails(orderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get order details: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"order_id": order.ID,
		"status":   order.Status,
	})
}
