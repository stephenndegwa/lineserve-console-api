package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/middleware"
	"github.com/lineserve/lineserve-api/pkg/models"
)

// VPSHandler handles VPS-related requests
type VPSHandler struct {
	SupabaseClient  *client.SupabaseClient
	OpenStackClient *client.OpenStackClient
}

// NewVPSHandler creates a new VPS handler
func NewVPSHandler(supabaseClient *client.SupabaseClient, openStackClient *client.OpenStackClient) *VPSHandler {
	return &VPSHandler{
		SupabaseClient:  supabaseClient,
		OpenStackClient: openStackClient,
	}
}

// ListPlans lists all available VPS plans
func (h *VPSHandler) ListPlans(c *fiber.Ctx) error {
	plans, err := h.SupabaseClient.GetVPSPlans()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get VPS plans: %v", err),
		})
	}

	return c.JSON(models.VPSPlansResponse{
		Plans: plans,
	})
}

// Subscribe subscribes a user to a VPS plan
func (h *VPSHandler) Subscribe(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Parse request body
	var req models.VPSSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Validate commit period
	validPeriods := map[int]bool{1: true, 3: true, 6: true, 12: true, 24: true}
	if !validPeriods[req.CommitPeriod] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid commit period. Must be 1, 3, 6, 12, or 24 months",
		})
	}

	// Get the plan
	plan, err := h.SupabaseClient.GetVPSPlanByCode(req.PlanCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Plan not found: %v", err),
		})
	}

	// Calculate price based on commit period
	var price float64
	switch req.CommitPeriod {
	case 1:
		price = plan.PriceMonthly
	case 3:
		price = plan.PriceCommit3M
	case 6:
		price = plan.PriceCommit6M
	case 12:
		price = plan.PriceCommit12M
	case 24:
		price = plan.PriceCommit24M
	default:
		price = plan.PriceMonthly
	}

	// If price for the selected period is not set, fall back to monthly
	if price == 0 {
		price = plan.PriceMonthly
	}

	// Calculate dates
	now := time.Now()
	endDate := now.AddDate(0, req.CommitPeriod, 0)
	renewalDueDate := endDate

	// Create subscription
	subscription := &models.VPSSubscription{
		UserID:         userID,
		PlanID:         plan.ID,
		CommitPeriod:   req.CommitPeriod,
		Price:          price,
		StartDate:      now,
		EndDate:        endDate,
		RenewalDueDate: renewalDueDate,
		AutoRenew:      true,
		Status:         "active",
	}

	// Save subscription to Supabase
	createdSubscription, err := h.SupabaseClient.CreateVPSSubscription(subscription)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create subscription: %v", err),
		})
	}

	// Provision OpenStack instance if flavor ID is available
	if plan.OpenStackFlavorID != "" && h.OpenStackClient != nil {
		// Get project ID from context
		projectID := c.Locals("project_id").(string)
		if projectID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Project ID not found",
			})
		}

		// Create instance name
		instanceName := fmt.Sprintf("vps-%s-%s", req.PlanCode, createdSubscription.ID[:8])

		// TODO: Get appropriate image ID and network ID
		imageID := "default-image-id"     // This should be replaced with actual image ID
		networkID := "default-network-id" // This should be replaced with actual network ID

		// Create instance using the compute client directly
		instance, err := h.provisionInstance(projectID, instanceName, plan.OpenStackFlavorID, imageID, networkID)
		if err != nil {
			// Update subscription with error status
			updates := map[string]interface{}{
				"status": "error",
			}
			_, updateErr := h.SupabaseClient.UpdateVPSSubscription(createdSubscription.ID, updates)
			if updateErr != nil {
				// Log the error but continue
				fmt.Printf("Failed to update subscription status: %v\n", updateErr)
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to provision instance: %v", err),
			})
		}

		// Update subscription with instance ID
		updates := map[string]interface{}{
			"instance_id":          instance.ID,
			"openstack_project_id": projectID,
		}
		createdSubscription, err = h.SupabaseClient.UpdateVPSSubscription(createdSubscription.ID, updates)
		if err != nil {
			// Log the error but continue
			fmt.Printf("Failed to update subscription with instance ID: %v\n", err)
		}
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(models.VPSSubscriptionResponse{
		Subscription: *createdSubscription,
		Message:      "Subscription created successfully",
	})
}

// provisionInstance provisions a new compute instance in OpenStack
func (h *VPSHandler) provisionInstance(projectID, name, flavorID, imageID, networkID string) (*models.Instance, error) {
	// Create a new compute handler
	computeHandler := NewComputeHandler("")

	// Create instance request
	req := models.CreateInstanceRequest{
		Name:      name,
		FlavorID:  flavorID,
		ImageID:   imageID,
		NetworkID: networkID,
	}

	// Create a mock fiber context with project ID
	// In a real implementation, this would be done differently
	ctx := fiber.New().AcquireCtx(nil)
	ctx.Locals("project_id", projectID)

	// Call the compute handler to create the instance
	instance, err := computeHandler.createInstanceInternal(ctx, req)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// ListSubscriptions lists all VPS subscriptions for the authenticated user
func (h *VPSHandler) ListSubscriptions(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get subscriptions from Supabase
	subscriptions, err := h.SupabaseClient.GetVPSSubscriptionsByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get subscriptions: %v", err),
		})
	}

	return c.JSON(models.VPSSubscriptionsResponse{
		Subscriptions: subscriptions,
	})
}

// CancelSubscription cancels auto-renewal for a VPS subscription
func (h *VPSHandler) CancelSubscription(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get subscription ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subscription ID is required",
		})
	}

	// Parse request body
	var req models.VPSCancelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Get subscription from Supabase
	subscription, err := h.SupabaseClient.GetVPSSubscriptionByID(id)
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

	// Update subscription
	updates := map[string]interface{}{
		"auto_renew": req.AutoRenew,
	}
	updatedSubscription, err := h.SupabaseClient.UpdateVPSSubscription(id, updates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update subscription: %v", err),
		})
	}

	// Return response
	return c.JSON(models.VPSSubscriptionResponse{
		Subscription: *updatedSubscription,
		Message:      "Subscription updated successfully",
	})
}

// RunRenewalBilling runs the VPS renewal billing process
func (h *VPSHandler) RunRenewalBilling(c *fiber.Ctx) error {
	// Check if user is admin
	isAdmin := middleware.IsAdmin(c)
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only administrators can run billing processes",
		})
	}

	// Run renewal billing
	results, err := h.SupabaseClient.RunVPSRenewalBilling()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to run renewal billing: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Renewal billing completed successfully",
		"results": results,
	})
}

// CreateOrder creates a new VPS order and invoice
func (h *VPSHandler) CreateOrder(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Parse request body
	var req models.VPSOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Validate commit period
	validPeriods := map[int]bool{1: true, 3: true, 6: true, 12: true, 24: true}
	if !validPeriods[req.CommitPeriod] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid commit period. Must be 1, 3, 6, 12, or 24 months",
		})
	}

	// Get the plan
	plan, err := h.SupabaseClient.GetVPSPlanByCode(req.PlanCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Plan not found: %v", err),
		})
	}

	// Calculate price based on commit period
	var price float64
	switch req.CommitPeriod {
	case 1:
		price = plan.PriceMonthly
	case 3:
		price = plan.PriceCommit3M
	case 6:
		price = plan.PriceCommit6M
	case 12:
		price = plan.PriceCommit12M
	case 24:
		price = plan.PriceCommit24M
	default:
		price = plan.PriceMonthly
	}

	// If price for the selected period is not set, fall back to monthly
	if price == 0 {
		price = plan.PriceMonthly
	}

	// Calculate dates
	now := time.Now()
	endDate := now.AddDate(0, req.CommitPeriod, 0)
	renewalDueDate := endDate
	invoiceExpiresAt := now.Add(24 * time.Hour) // Invoice expires in 24 hours

	// Create subscription with pending status
	subscription := &models.VPSSubscription{
		UserID:         userID,
		PlanID:         plan.ID,
		CommitPeriod:   req.CommitPeriod,
		Price:          price,
		StartDate:      now,
		EndDate:        endDate,
		RenewalDueDate: renewalDueDate,
		AutoRenew:      true,
		Status:         "pending", // Start as pending until payment is confirmed
	}

	// Save subscription to Supabase
	createdSubscription, err := h.SupabaseClient.CreateVPSSubscription(subscription)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create subscription: %v", err),
		})
	}

	// Create invoice
	invoice := &models.VPSInvoice{
		UserID:          userID,
		SubscriptionID:  createdSubscription.ID,
		PlanCode:        req.PlanCode,
		PeriodMonths:    req.CommitPeriod,
		Amount:          price,
		Currency:        "USD",
		Status:          "unpaid",
		PaymentMethodID: req.PaymentMethodID,
		CreatedAt:       now,
		ExpiresAt:       invoiceExpiresAt,
	}

	// Save invoice to Supabase
	createdInvoice, err := h.SupabaseClient.CreateVPSInvoice(invoice)
	if err != nil {
		// If invoice creation fails, delete the subscription
		// This is a simplistic approach to rollback - in production, use transactions
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create invoice: %v", err),
		})
	}

	// Generate payment URL
	baseURL := c.BaseURL()
	if baseURL == "" {
		baseURL = "https://lineserve.net" // Default URL if not available from context
	}
	paymentURL := fmt.Sprintf("%s/payment/invoice/%s", baseURL, createdInvoice.ID)

	// Return response
	return c.Status(fiber.StatusCreated).JSON(models.VPSOrderResponse{
		SubscriptionID: createdSubscription.ID,
		InvoiceID:      createdInvoice.ID,
		Amount:         price,
		Currency:       "USD",
		PaymentURL:     paymentURL,
	})
}

// GetInvoice gets a VPS invoice by ID
func (h *VPSHandler) GetInvoice(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get invoice ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invoice ID is required",
		})
	}

	// Get invoice from Supabase
	invoice, err := h.SupabaseClient.GetVPSInvoiceByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Invoice not found: %v", err),
		})
	}

	// Check if invoice belongs to user
	if invoice.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You do not have permission to view this invoice",
		})
	}

	// Return invoice
	return c.JSON(invoice)
}

// PayInvoice pays a VPS invoice and provisions the VPS
func (h *VPSHandler) PayInvoice(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get invoice ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invoice ID is required",
		})
	}

	// Parse request body
	var req models.VPSInvoicePayRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Get invoice from Supabase
	invoice, err := h.SupabaseClient.GetVPSInvoiceByID(id)
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

	// Process payment (this would integrate with a payment gateway in a real implementation)
	// For this example, we'll just simulate a successful payment
	paymentSuccessful := true
	paymentIntentID := "pi_" + uuid.New().String()

	if !paymentSuccessful {
		// Update invoice status to failed
		updates := map[string]interface{}{
			"status":            "failed",
			"payment_method_id": req.PaymentMethodID,
		}
		_, err := h.SupabaseClient.UpdateVPSInvoice(id, updates)
		if err != nil {
			// Log the error but continue
			fmt.Printf("Failed to update invoice status: %v\n", err)
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Payment failed",
		})
	}

	// Update invoice status to paid
	now := time.Now()
	invoiceUpdates := map[string]interface{}{
		"status":            "paid",
		"payment_method_id": req.PaymentMethodID,
		"payment_intent_id": paymentIntentID,
		"paid_at":           now,
	}
	_, err = h.SupabaseClient.UpdateVPSInvoice(id, invoiceUpdates)
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
	if h.OpenStackClient != nil && subscription.Plan != nil && subscription.Plan.OpenStackFlavorID != "" {
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
		instance, err := h.provisionInstance(projectID, instanceName, subscription.Plan.OpenStackFlavorID, imageID, networkID)
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

// ListInvoices lists all invoices for the authenticated user
func (h *VPSHandler) ListInvoices(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get invoices from Supabase
	invoices, err := h.SupabaseClient.GetVPSInvoicesByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get invoices: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"invoices": invoices,
	})
}
