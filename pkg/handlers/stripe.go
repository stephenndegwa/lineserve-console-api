package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
	"github.com/stripe/stripe-go/v72"
)

// StripeHandler handles Stripe-related requests
type StripeHandler struct {
	SupabaseClient *client.SupabaseClient
	StripeClient   *client.StripeClient
}

// NewStripeHandler creates a new Stripe handler
func NewStripeHandler(supabaseClient *client.SupabaseClient, stripeClient *client.StripeClient) *StripeHandler {
	return &StripeHandler{
		SupabaseClient: supabaseClient,
		StripeClient:   stripeClient,
	}
}

// CreateCheckoutSession creates a new Stripe checkout session for a VPS invoice
func (h *StripeHandler) CreateCheckoutSession(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Parse request body
	var req models.StripeCheckoutRequest
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
			"error": "You do not have permission to access this invoice",
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

	// Get user's Stripe customer ID or create a new customer
	var customerID string
	user, err := h.SupabaseClient.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get user: %v", err),
		})
	}

	if user.StripeCustomerID == "" {
		// Create a new customer in Stripe
		customer, err := h.StripeClient.CreateCustomer(c.Context(), user.Email, user.Name)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create Stripe customer: %v", err),
			})
		}
		customerID = customer.ID

		// Update user with Stripe customer ID
		err = h.SupabaseClient.UpdateUserStripeCustomerID(userID, customerID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to update user with Stripe customer ID: %v", err),
			})
		}
	} else {
		customerID = user.StripeCustomerID
	}

	// Generate success and cancel URLs
	baseURL := c.BaseURL()
	if baseURL == "" {
		baseURL = "https://lineserve.net" // Default URL if not available from context
	}
	successURL := fmt.Sprintf("%s/payment/success?invoice_id=%s", baseURL, invoice.ID)
	cancelURL := fmt.Sprintf("%s/payment/cancel?invoice_id=%s", baseURL, invoice.ID)

	// Create line items for checkout session
	lineItems := []*stripe.CheckoutSessionLineItemParams{
		{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("usd"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(fmt.Sprintf("VPS Plan %s - %d months", invoice.PlanCode, invoice.PeriodMonths)),
				},
				UnitAmount: stripe.Int64(int64(invoice.Amount * 100)), // Convert to cents
			},
			Quantity: stripe.Int64(1),
		},
	}

	// Create checkout session
	session, err := h.StripeClient.CreateCheckoutSession(c.Context(), customerID, successURL, cancelURL, lineItems)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create checkout session: %v", err),
		})
	}

	// Update invoice with Stripe session ID
	updates := map[string]interface{}{
		"stripe_session_id": session.ID,
	}
	_, err = h.SupabaseClient.UpdateVPSInvoice(invoice.ID, updates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update invoice: %v", err),
		})
	}

	// Return checkout session URL
	return c.JSON(models.StripeCheckoutResponse{
		CheckoutURL: session.URL,
		SessionID:   session.ID,
	})
}

// HandleWebhook handles Stripe webhook events
func (h *StripeHandler) HandleWebhook(c *fiber.Ctx) error {
	// Get webhook secret from environment
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Stripe webhook secret is not configured",
		})
	}

	// Get signature from header
	signature := c.Get("Stripe-Signature")
	if signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing Stripe-Signature header",
		})
	}

	// Read request body
	body := c.Body()

	// Verify webhook signature
	event, err := h.StripeClient.VerifyWebhookSignature(body, signature, webhookSecret)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid webhook signature: %v", err),
		})
	}

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		// Parse the checkout session
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to parse checkout session: %v", err),
			})
		}

		// Find the invoice with this session ID
		invoice, err := h.SupabaseClient.GetVPSInvoiceByStripeSessionID(session.ID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("Invoice not found for session ID %s: %v", session.ID, err),
			})
		}

		// Update invoice status to paid
		now := time.Now()
		invoiceUpdates := map[string]interface{}{
			"status":            "paid",
			"stripe_payment_id": session.PaymentIntent.ID,
			"paid_at":           now,
		}
		_, err = h.SupabaseClient.UpdateVPSInvoice(invoice.ID, invoiceUpdates)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to update invoice: %v", err),
			})
		}

		// Update subscription status to active
		subscriptionUpdates := map[string]interface{}{
			"status":     "active",
			"start_date": now,
		}
		_, err = h.SupabaseClient.UpdateVPSSubscription(invoice.SubscriptionID, subscriptionUpdates)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to update subscription: %v", err),
			})
		}

		// Provision VPS if OpenStack client is available
		// This would typically be handled by the VPS handler after payment confirmation

	case "invoice.paid":
		// Handle subscription renewal payments
		var stripeInvoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &stripeInvoice)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to parse invoice: %v", err),
			})
		}

		// If this is a subscription invoice, update the subscription
		if stripeInvoice.Subscription != nil {
			// Find the subscription with this Stripe subscription ID
			subscription, err := h.SupabaseClient.GetVPSSubscriptionByStripeID(stripeInvoice.Subscription.ID)
			if err != nil {
				// This might be a new subscription, so we don't return an error
				fmt.Printf("Subscription not found for Stripe subscription ID %s: %v\n", stripeInvoice.Subscription.ID, err)
			} else {
				// Update subscription with new period end
				subscriptionUpdates := map[string]interface{}{
					"end_date":         time.Unix(stripeInvoice.PeriodEnd, 0),
					"renewal_due_date": time.Unix(stripeInvoice.PeriodEnd, 0),
				}
				_, err = h.SupabaseClient.UpdateVPSSubscription(subscription.ID, subscriptionUpdates)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": fmt.Sprintf("Failed to update subscription: %v", err),
					})
				}
			}
		}

	case "customer.subscription.deleted":
		// Handle subscription cancellation
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to parse subscription: %v", err),
			})
		}

		// Find the subscription with this Stripe subscription ID
		vpsSubscription, err := h.SupabaseClient.GetVPSSubscriptionByStripeID(subscription.ID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("Subscription not found for Stripe subscription ID %s: %v", subscription.ID, err),
			})
		}

		// Update subscription status to cancelled
		subscriptionUpdates := map[string]interface{}{
			"status": "cancelled",
		}
		_, err = h.SupabaseClient.UpdateVPSSubscription(vpsSubscription.ID, subscriptionUpdates)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to update subscription: %v", err),
			})
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

// CreateSubscription creates a new Stripe subscription
func (h *StripeHandler) CreateSubscription(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Parse request body
	var req models.StripeSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Validate price ID
	if req.PriceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Price ID is required",
		})
	}

	// Get user's Stripe customer ID or create a new customer
	var customerID string
	user, err := h.SupabaseClient.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get user: %v", err),
		})
	}

	if user.StripeCustomerID == "" {
		// Create a new customer in Stripe
		customer, err := h.StripeClient.CreateCustomer(c.Context(), user.Email, user.Name)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create Stripe customer: %v", err),
			})
		}
		customerID = customer.ID

		// Update user with Stripe customer ID
		err = h.SupabaseClient.UpdateUserStripeCustomerID(userID, customerID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to update user with Stripe customer ID: %v", err),
			})
		}
	} else {
		customerID = user.StripeCustomerID
	}

	// Generate success and cancel URLs
	baseURL := c.BaseURL()
	if baseURL == "" {
		baseURL = "https://lineserve.net" // Default URL if not available from context
	}
	successURL := fmt.Sprintf("%s/subscription/success", baseURL)
	cancelURL := fmt.Sprintf("%s/subscription/cancel", baseURL)

	// Create checkout session for subscription
	session, err := h.StripeClient.CreateSubscriptionCheckoutSession(c.Context(), customerID, req.PriceID, successURL, cancelURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create subscription checkout session: %v", err),
		})
	}

	// Return checkout session URL
	return c.JSON(models.StripeCheckoutResponse{
		CheckoutURL: session.URL,
		SessionID:   session.ID,
	})
}

// CancelSubscription cancels a Stripe subscription
func (h *StripeHandler) CancelSubscription(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get subscription ID from URL
	stripeSubscriptionID := c.Params("id")
	if stripeSubscriptionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Stripe subscription ID is required",
		})
	}

	// Parse request body
	var req models.StripeCancelSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Find the subscription with this Stripe subscription ID
	subscription, err := h.SupabaseClient.GetVPSSubscriptionByStripeID(stripeSubscriptionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Subscription not found: %v", err),
		})
	}

	// Check if subscription belongs to user
	if subscription.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You do not have permission to cancel this subscription",
		})
	}

	// Cancel subscription in Stripe
	_, err = h.StripeClient.CancelSubscription(c.Context(), stripeSubscriptionID, req.CancelAtPeriodEnd)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to cancel subscription: %v", err),
		})
	}

	// Update subscription status
	var status string
	if req.CancelAtPeriodEnd {
		status = "cancelling" // Will be cancelled at the end of the period
	} else {
		status = "cancelled" // Cancelled immediately
	}

	subscriptionUpdates := map[string]interface{}{
		"status": status,
	}
	updatedSubscription, err := h.SupabaseClient.UpdateVPSSubscription(subscription.ID, subscriptionUpdates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update subscription: %v", err),
		})
	}

	return c.JSON(models.VPSSubscriptionResponse{
		Subscription: *updatedSubscription,
		Message:      "Subscription cancelled successfully",
	})
}
