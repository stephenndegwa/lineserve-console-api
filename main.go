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
		MemberRoleID:   "93f6b134e78644d69817b8061205f339", // Updated member role ID
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

	// Image routes
	imageHandler := handlers.NewImageHandler(jwtSecret)
	projectScoped.Get("/images", imageHandler.ListImages)
	projectScoped.Get("/images/:id", imageHandler.GetImage)

	// Flavor routes
	projectScoped.Get("/flavors", instanceHandler.ListFlavors)

	// Create OpenStack client for the remaining handlers (optional)
	var openStackClient *client.OpenStackClient
	var networkHandler *handlers.NetworkHandler
	var volumeHandler *handlers.VolumeHandler
	var projectHandler *handlers.ProjectHandler
	var keyPairHandler *handlers.KeyPairHandler

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
	} else {
		// Create real handlers with OpenStack client
		networkHandler = handlers.NewNetworkHandler(openStackClient)
		volumeHandler = handlers.NewVolumeHandler(openStackClient)
		projectHandler = handlers.NewProjectHandler(openStackClient)
		keyPairHandler = handlers.NewKeyPairHandler(openStackClient)
	}

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
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
