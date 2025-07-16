package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/config"
	"github.com/lineserve/lineserve-api/pkg/cron"
	"github.com/lineserve/lineserve-api/pkg/handlers"
	"github.com/lineserve/lineserve-api/pkg/middleware"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create PostgreSQL client
	postgresClient, err := client.NewPostgresClient(cfg.GetPostgresConnectionString())
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL client: %v", err)
	}
	defer postgresClient.Close()

	// Create tables if they don't exist
	if err := postgresClient.CreateTablesIfNotExist(context.Background()); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Add middleware
	app.Use(cors.New())
	app.Use(logger.New())

	// Get JWT secret
	jwtSecret := cfg.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "lineserve-secret-key" // Default secret
	}
	fmt.Printf("Main: Using JWT secret: %s\n", jwtSecret)

	// Serve Swagger documentation
	app.Get("/docs/swagger.json", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	// Serve Swagger UI
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/index.html")
	})

	// Create API group with version
	v1 := app.Group("/v1")

	// Create auth handler
	authHandler := &handlers.AuthHandler{
		PostgresClient: postgresClient,
		JWTSecret:      jwtSecret,
		MemberRoleID:   "c76575246ae343ddb80d0f0f1f2d958b", // Updated member role ID
	}

	// Public routes
	v1.Post("/login", authHandler.Login)
	v1.Post("/register", authHandler.Register)
	v1.Post("/project-token", authHandler.GetProjectToken)

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.JWTMiddleware(jwtSecret))

	// User routes (require authentication but not project scope)
	protected.Get("/projects", func(c *fiber.Ctx) error {
		return authHandler.ListProjects(c)
	})

	// Project-scoped routes
	projectScoped := protected.Group("/")
	projectScoped.Use(middleware.ProjectScopeRequired())

	// Instance routes
	instanceHandler := handlers.NewComputeHandler(jwtSecret)
	projectScoped.Get("/instances", instanceHandler.ListInstances)
	projectScoped.Post("/instances", instanceHandler.CreateInstance)
	projectScoped.Get("/instances/:id", instanceHandler.GetInstance)
	projectScoped.Delete("/instances/:id", instanceHandler.DeleteInstance)
	projectScoped.Put("/instances/:id", instanceHandler.UpdateInstance)
	projectScoped.Post("/instances/:id/action", instanceHandler.PerformInstanceAction)

	// Image routes
	imageHandler := handlers.NewImageHandler(jwtSecret)
	projectScoped.Get("/images", imageHandler.ListImages)
	projectScoped.Get("/images/:id", imageHandler.GetImage)
	projectScoped.Post("/images", imageHandler.CreateImage)
	projectScoped.Delete("/images/:id", imageHandler.DeleteImage)

	// Flavor routes
	projectScoped.Get("/flavors", instanceHandler.ListFlavors)

	// Create OpenStack client for the remaining handlers (optional)
	var openStackClient *client.OpenStackClient
	var networkHandler *handlers.NetworkHandler
	var volumeHandler *handlers.VolumeHandler
	var projectHandler *handlers.ProjectHandler
	var keyPairHandler *handlers.KeyPairHandler
	var floatingIPHandler *handlers.FloatingIPHandler
	var securityGroupHandler *handlers.SecurityGroupHandler
	var subnetHandler *handlers.SubnetHandler
	var routerHandler *handlers.RouterHandler
	var vpsHandler *handlers.VPSHandler
	var supabaseClient *client.SupabaseClient
	var paypalClient *client.PayPalClient
	var paypalHandler *handlers.PayPalHandler

	// Try to create OpenStack client, but don't fail if it doesn't work
	openStackClient, err = client.NewOpenStackClient()
	if err != nil {
		log.Printf("Warning: Failed to create OpenStack client: %v", err)
		log.Println("Some features requiring OpenStack will be unavailable")

		// Create mock handlers that return "not implemented" for OpenStack features
		networkHandler = &handlers.NetworkHandler{}
		volumeHandler = &handlers.VolumeHandler{}
		projectHandler = &handlers.ProjectHandler{}
		keyPairHandler = &handlers.KeyPairHandler{}
		floatingIPHandler = &handlers.FloatingIPHandler{}
		securityGroupHandler = &handlers.SecurityGroupHandler{}
		subnetHandler = &handlers.SubnetHandler{}
		routerHandler = &handlers.RouterHandler{}
	} else {
		// Create real handlers with OpenStack client
		networkHandler = handlers.NewNetworkHandler(openStackClient)
		volumeHandler = handlers.NewVolumeHandler(openStackClient)
		projectHandler = handlers.NewProjectHandler(openStackClient)
		keyPairHandler = handlers.NewKeyPairHandler(openStackClient)
		floatingIPHandler = handlers.NewFloatingIPHandler(openStackClient)
		securityGroupHandler = handlers.NewSecurityGroupHandler(openStackClient)
		subnetHandler = handlers.NewSubnetHandler(openStackClient)
		routerHandler = handlers.NewRouterHandler(openStackClient)
	}

	// Try to create Supabase client
	supabaseClient, err = client.NewSupabaseClient()
	if err != nil {
		log.Printf("Warning: Failed to create Supabase client: %v", err)
		log.Println("VPS features will be unavailable")
		vpsHandler = &handlers.VPSHandler{}
	} else {
		// Create VPS handler with Supabase client
		vpsHandler = handlers.NewVPSHandler(supabaseClient, openStackClient)

		// Start VPS billing cron job
		go cron.StartVPSBillingCron(supabaseClient)
	}

	// Try to create PayPal client
	paypalClient, err = client.NewPayPalClient()
	if err != nil {
		log.Printf("Warning: Failed to create PayPal client: %v", err)
		log.Println("PayPal payment features will be unavailable")
	} else {
		// Create PayPal handler with PayPal client
		paypalHandler = handlers.NewPayPalHandler(paypalClient, supabaseClient, vpsHandler)
	}

	// Initialize Flutterwave client
	flutterwaveClient, err := client.GetFlutterwaveClientFromEnv()
	if err != nil {
		log.Printf("Failed to initialize Flutterwave client: %v", err)
	}

	// Initialize Flutterwave handler
	flutterwaveHandler := handlers.NewFlutterwaveHandler(supabaseClient, flutterwaveClient)

	// Initialize Stripe client
	stripeClient, err := client.GetStripeClientFromEnv()
	if err != nil {
		log.Printf("Failed to initialize Stripe client: %v", err)
	}

	// Initialize Stripe handler
	stripeHandler := handlers.NewStripeHandler(supabaseClient, stripeClient)

	// Initialize M-Pesa client
	mpesaClient, err := client.GetMPesaClientFromEnv()
	if err != nil {
		log.Printf("Failed to initialize M-Pesa client: %v", err)
	}

	// Initialize M-Pesa handler
	mpesaHandler := handlers.NewMPesaHandler(supabaseClient, mpesaClient)

	// Network routes
	projectScoped.Get("/networks", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return networkHandler.ListNetworks(c)
	})
	projectScoped.Post("/networks", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return networkHandler.CreateNetwork(c)
	})
	projectScoped.Get("/networks/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return networkHandler.GetNetwork(c)
	})
	projectScoped.Delete("/networks/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return networkHandler.DeleteNetwork(c)
	})

	// Volume routes
	projectScoped.Get("/volumes", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return volumeHandler.ListVolumes(c)
	})
	projectScoped.Post("/volumes", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return volumeHandler.CreateVolume(c)
	})
	projectScoped.Get("/volumes/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return volumeHandler.GetVolume(c)
	})
	projectScoped.Delete("/volumes/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return volumeHandler.DeleteVolume(c)
	})
	projectScoped.Post("/volumes/:id/attach", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return volumeHandler.AttachVolume(c)
	})
	projectScoped.Post("/volumes/:id/detach", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return volumeHandler.DetachVolume(c)
	})
	projectScoped.Put("/volumes/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return volumeHandler.ResizeVolume(c)
	})
	projectScoped.Get("/volume-types", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return volumeHandler.ListVolumeTypes(c)
	})

	// Project routes
	projectScoped.Get("/projects", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return projectHandler.ListProjects(c)
	})
	projectScoped.Get("/projects/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return projectHandler.GetProject(c)
	})

	// Key pair routes
	projectScoped.Get("/keypairs", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return keyPairHandler.ListKeyPairs(c)
	})
	projectScoped.Post("/keypairs", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return keyPairHandler.CreateKeyPair(c)
	})
	projectScoped.Get("/keypairs/:name", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return keyPairHandler.GetKeyPair(c)
	})
	projectScoped.Delete("/keypairs/:name", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return keyPairHandler.DeleteKeyPair(c)
	})

	// Floating IP routes
	projectScoped.Get("/floating-ips", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return floatingIPHandler.ListFloatingIPs(c)
	})
	projectScoped.Post("/floating-ips", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return floatingIPHandler.CreateFloatingIP(c)
	})
	projectScoped.Get("/floating-ips/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return floatingIPHandler.GetFloatingIP(c)
	})
	projectScoped.Put("/floating-ips/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return floatingIPHandler.UpdateFloatingIP(c)
	})
	projectScoped.Delete("/floating-ips/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return floatingIPHandler.DeleteFloatingIP(c)
	})

	// Security Group routes
	projectScoped.Get("/security-groups", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return securityGroupHandler.ListSecurityGroups(c)
	})
	projectScoped.Get("/security-groups/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return securityGroupHandler.GetSecurityGroup(c)
	})
	projectScoped.Post("/security-groups", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return securityGroupHandler.CreateSecurityGroup(c)
	})
	projectScoped.Delete("/security-groups/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return securityGroupHandler.DeleteSecurityGroup(c)
	})

	// Security Group Rule routes
	projectScoped.Get("/security-group-rules", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return securityGroupHandler.ListSecurityGroupRules(c)
	})
	projectScoped.Post("/security-group-rules", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return securityGroupHandler.CreateSecurityGroupRule(c)
	})
	projectScoped.Delete("/security-group-rules/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return securityGroupHandler.DeleteSecurityGroupRule(c)
	})

	// Subnet routes
	projectScoped.Get("/subnets", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return subnetHandler.ListSubnets(c)
	})
	projectScoped.Get("/subnets/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return subnetHandler.GetSubnet(c)
	})
	projectScoped.Post("/subnets", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return subnetHandler.CreateSubnet(c)
	})
	projectScoped.Delete("/subnets/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return subnetHandler.DeleteSubnet(c)
	})

	// Router routes
	projectScoped.Get("/routers", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return routerHandler.ListRouters(c)
	})
	projectScoped.Get("/routers/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return routerHandler.GetRouter(c)
	})
	projectScoped.Post("/routers", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return routerHandler.CreateRouter(c)
	})
	projectScoped.Delete("/routers/:id", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return routerHandler.DeleteRouter(c)
	})
	projectScoped.Put("/routers/:id/interfaces", func(c *fiber.Ctx) error {
		if openStackClient == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "OpenStack service unavailable",
			})
		}
		return routerHandler.UpdateRouterInterfaces(c)
	})

	// VPS routes
	vpsRoutes := projectScoped.Group("/vps")
	vpsRoutes.Get("/plans", vpsHandler.ListPlans)
	vpsRoutes.Post("/subscribe", vpsHandler.Subscribe)
	vpsRoutes.Get("/subscriptions", vpsHandler.ListSubscriptions)
	vpsRoutes.Post("/subscriptions/:id/cancel", vpsHandler.CancelSubscription)

	// New VPS order and invoice routes
	vpsRoutes.Post("/order", vpsHandler.CreateOrder)
	vpsRoutes.Get("/invoice/:id", vpsHandler.GetInvoice)
	vpsRoutes.Post("/invoice/:id/pay", vpsHandler.PayInvoice)
	vpsRoutes.Get("/invoices", vpsHandler.ListInvoices)

	// PayPal routes
	if paypalClient != nil {
		paypalRoutes := projectScoped.Group("/paypal")
		paypalRoutes.Post("/create-order", paypalHandler.CreateOrder)
		paypalRoutes.Post("/capture-order", paypalHandler.CaptureOrder)
		paypalRoutes.Get("/order/:id", paypalHandler.GetOrderStatus)

		// Webhook endpoint (no authentication required)
		v1.Post("/paypal/webhook", paypalHandler.HandleWebhook)
	}

	// Stripe routes
	if stripeClient != nil {
		stripeRoutes := projectScoped.Group("/stripe")
		stripeRoutes.Post("/checkout", stripeHandler.CreateCheckoutSession)
		stripeRoutes.Post("/subscription", stripeHandler.CreateSubscription)
		stripeRoutes.Post("/subscription/:id/cancel", stripeHandler.CancelSubscription)
		stripeRoutes.Post("/webhook", stripeHandler.HandleWebhook)
	}

	// Flutterwave routes
	v1.Post("/flutterwave/create-payment", flutterwaveHandler.CreatePayment)
	v1.Post("/flutterwave/webhook", flutterwaveHandler.HandleWebhook)
	v1.Get("/flutterwave/verify/:id", flutterwaveHandler.VerifyPayment)
	v1.Get("/flutterwave/status/:tx_ref", flutterwaveHandler.GetPaymentStatus)

	// M-Pesa routes
	if mpesaClient != nil {
		v1.Post("/mpesa/stk-push", mpesaHandler.InitiateSTKPush)
		v1.Post("/mpesa/callback", mpesaHandler.HandleSTKPushCallback)
		v1.Post("/mpesa/check-status", mpesaHandler.CheckSTKPushStatus)
	}

	// Admin routes
	adminRoutes := protected.Group("/admin")
	adminRoutes.Use(middleware.AdminRequired())
	adminRoutes.Post("/vps/billing/run", vpsHandler.RunRenewalBilling)

	// Add a root endpoint that shows API info
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "Lineserve API",
			"version": "1.0.0",
			"docs":    "/docs",
		})
	})

	// Add a test endpoint to check OpenStack client status
	app.Get("/api/v1/openstack-status", func(c *fiber.Ctx) error {
		status := map[string]interface{}{
			"openstack_client": openStackClient != nil,
			"services": map[string]interface{}{
				"compute":  openStackClient != nil && openStackClient.Compute != nil,
				"network":  openStackClient != nil && openStackClient.Network != nil,
				"image":    openStackClient != nil && openStackClient.Image != nil,
				"volume":   openStackClient != nil && openStackClient.Volume != nil,
				"identity": openStackClient != nil && openStackClient.Identity != nil,
			},
		}
		return c.JSON(status)
	})

	// Start server
	port := cfg.APIPort
	if port == "" {
		port = "3070"
	}
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
