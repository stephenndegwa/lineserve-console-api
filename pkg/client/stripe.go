package client

import (
	"context"
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/refund"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/webhook"
)

// StripeClient represents a client for interacting with the Stripe API
type StripeClient struct {
	SecretKey string
	PublicKey string
}

// GetStripeClientFromEnv creates a new Stripe client from environment variables
func GetStripeClientFromEnv() (*StripeClient, error) {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	publicKey := os.Getenv("STRIPE_PUBLIC_KEY")

	if secretKey == "" || publicKey == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY and STRIPE_PUBLIC_KEY must be set")
	}

	return &StripeClient{
		SecretKey: secretKey,
		PublicKey: publicKey,
	}, nil
}

// NewStripeClient creates a new Stripe client
func NewStripeClient(secretKey, publicKey string) *StripeClient {
	return &StripeClient{
		SecretKey: secretKey,
		PublicKey: publicKey,
	}
}

// Initialize sets up the Stripe API key
func (c *StripeClient) Initialize() {
	stripe.Key = c.SecretKey
}

// CreateCustomer creates a new customer in Stripe
func (c *StripeClient) CreateCustomer(ctx context.Context, email, name string) (*stripe.Customer, error) {
	c.Initialize()

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}

	return customer.New(params)
}

// AttachPaymentMethod attaches a payment method to a customer
func (c *StripeClient) AttachPaymentMethod(ctx context.Context, paymentMethodID, customerID string) error {
	c.Initialize()

	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}

	_, err := paymentmethod.Attach(paymentMethodID, params)
	return err
}

// CreatePaymentIntent creates a payment intent
func (c *StripeClient) CreatePaymentIntent(ctx context.Context, amount int64, currency, customerID, paymentMethodID, description string) (*stripe.PaymentIntent, error) {
	c.Initialize()

	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(amount),
		Currency:           stripe.String(currency),
		Customer:           stripe.String(customerID),
		PaymentMethod:      stripe.String(paymentMethodID),
		Description:        stripe.String(description),
		Confirm:            stripe.Bool(true),
		ConfirmationMethod: stripe.String(string(stripe.PaymentIntentConfirmationMethodAutomatic)),
		SetupFutureUsage:   stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
	}

	return paymentintent.New(params)
}

// ConfirmPaymentIntent confirms a payment intent
func (c *StripeClient) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string) (*stripe.PaymentIntent, error) {
	c.Initialize()

	params := &stripe.PaymentIntentConfirmParams{}
	return paymentintent.Confirm(paymentIntentID, params)
}

// GetPaymentIntent retrieves a payment intent by ID
func (c *StripeClient) GetPaymentIntent(ctx context.Context, paymentIntentID string) (*stripe.PaymentIntent, error) {
	c.Initialize()
	return paymentintent.Get(paymentIntentID, nil)
}

// CreateRefund creates a refund for a payment intent
func (c *StripeClient) CreateRefund(ctx context.Context, paymentIntentID string, amount int64) (*stripe.Refund, error) {
	c.Initialize()

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}

	if amount > 0 {
		params.Amount = stripe.Int64(amount)
	}

	return refund.New(params)
}

// CreateCheckoutSession creates a checkout session for a one-time payment
func (c *StripeClient) CreateCheckoutSession(ctx context.Context, customerID, successURL, cancelURL string, lineItems []*stripe.CheckoutSessionLineItemParams) (*stripe.CheckoutSession, error) {
	c.Initialize()

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(customerID),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems:  lineItems,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
	}

	return session.New(params)
}

// CreateSubscriptionCheckoutSession creates a checkout session for a subscription
func (c *StripeClient) CreateSubscriptionCheckoutSession(ctx context.Context, customerID, priceID, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	c.Initialize()

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(customerID),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
	}

	return session.New(params)
}

// CancelSubscription cancels a subscription
func (c *StripeClient) CancelSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool) (*stripe.Subscription, error) {
	c.Initialize()

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(cancelAtPeriodEnd),
	}

	return sub.Update(subscriptionID, params)
}

// GetSubscription retrieves a subscription by ID
func (c *StripeClient) GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
	c.Initialize()
	return sub.Get(subscriptionID, nil)
}

// VerifyWebhookSignature verifies the signature of a webhook event
func (c *StripeClient) VerifyWebhookSignature(payload []byte, signature, webhookSecret string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, signature, webhookSecret)
}
