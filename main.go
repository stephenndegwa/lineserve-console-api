package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/handlers"
	"github.com/lineserve/lineserve-api/pkg/middleware"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Create OpenStack client
	openStackClient, err := client.NewOpenStackClient()
	if err != nil {
		log.Fatalf("Failed to create OpenStack client: %v", err)
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
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "lineserve-secret-key" // Default secret
	}
	fmt.Printf("Main: Using JWT secret: %s\n", jwtSecret)

	// Create API group with version
	v1 := app.Group("/v1")

	// Public routes
	v1.Post("/login", handlers.NewAuthHandler(jwtSecret).Login)

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.JWTMiddleware(jwtSecret))

	// Instance routes
	instanceHandler := handlers.NewComputeHandler(openStackClient)
	protected.Get("/instances", instanceHandler.ListInstances)
	protected.Post("/instances", instanceHandler.CreateInstance)
	protected.Get("/instances/:id", instanceHandler.GetInstance)

	// Image routes
	imageHandler := handlers.NewImageHandler(openStackClient)
	protected.Get("/images", imageHandler.ListImages)
	protected.Get("/images/:id", imageHandler.GetImage)

	// Flavor routes
	protected.Get("/flavors", instanceHandler.ListFlavors)

	// Network routes
	networkHandler := handlers.NewNetworkHandler(openStackClient)
	protected.Get("/networks", networkHandler.ListNetworks)
	protected.Get("/networks/:id", networkHandler.GetNetwork)

	// Volume routes
	volumeHandler := handlers.NewVolumeHandler(openStackClient)
	protected.Get("/volumes", volumeHandler.ListVolumes)
	protected.Post("/volumes", volumeHandler.CreateVolume)
	protected.Get("/volumes/:id", volumeHandler.GetVolume)

	// Project routes
	projectHandler := handlers.NewProjectHandler(openStackClient)
	protected.Get("/projects", projectHandler.ListProjects)
	protected.Get("/projects/:id", projectHandler.GetProject)

	// Add a root endpoint that shows API info
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "Lineserve API",
			"version": "1.0.0",
			"docs":    "/v1",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
