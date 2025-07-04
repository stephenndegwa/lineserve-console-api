package handlers

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lineserve/lineserve-api/pkg/client"
	"github.com/lineserve/lineserve-api/pkg/models"
	"github.com/lineserve/lineserve-api/pkg/openstack"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication
type AuthHandler struct {
	JWTSecret      string
	PostgresClient *client.PostgresClient
	MemberRoleID   string
	DomainName     string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(jwtSecret string, postgresClient *client.PostgresClient, memberRoleID, domainName string) *AuthHandler {
	return &AuthHandler{
		JWTSecret:      jwtSecret,
		PostgresClient: postgresClient,
		MemberRoleID:   memberRoleID,
		DomainName:     domainName,
	}
}

// Login handles user login with unscoped authentication
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// Parse request body
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Username and password are required",
		})
	}

	// Use default domain if not provided
	domainName := req.DomainName
	if domainName == "" {
		domainName = h.DomainName
	}

	// Create context
	ctx := c.Context()

	// Authenticate with OpenStack (unscoped)
	provider, userID, err := openstack.AuthenticateUnscoped(ctx, req.Username, req.Password, domainName)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Authentication failed: %v", err),
		})
	}

	// List projects for the user
	openstackProjects, err := openstack.ListUserProjects(ctx, provider, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to list projects: %v", err),
		})
	}

	// Convert OpenStack projects to model projects
	projects := make([]models.Project, len(openstackProjects))
	for i, p := range openstackProjects {
		projects[i] = models.Project{
			ID:       p.ID,
			Name:     p.Name,
			DomainID: p.DomainID,
		}
	}

	// Create JWT token
	jwtToken := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	expiresAt := time.Now().Add(time.Hour * 24)
	claims := jwtToken.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["username"] = req.Username
	claims["domain_name"] = domainName
	claims["exp"] = expiresAt.Unix()

	// Generate encoded token
	encodedToken, err := jwtToken.SignedString([]byte(h.JWTSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to generate token",
		})
	}

	// Return token and projects
	return c.JSON(models.LoginResponse{
		Token:     encodedToken,
		UserID:    userID,
		Projects:  projects,
		ExpiresAt: expiresAt,
	})
}

// GetProjectToken handles getting a project-scoped token
func (h *AuthHandler) GetProjectToken(c *fiber.Ctx) error {
	// Parse request body
	var req models.ProjectScopeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" || req.ProjectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Username, password, and project_id are required",
		})
	}

	// Use default domain if not provided
	domainName := req.DomainName
	if domainName == "" {
		domainName = h.DomainName
	}

	// Create context
	ctx := c.Context()

	// Authenticate with OpenStack (scoped to project)
	provider, err := openstack.AuthenticateScoped(ctx, req.Username, req.Password, domainName, req.ProjectID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Project authentication failed: %v", err),
		})
	}

	// Get auth result to extract expiration time
	authResult, err := openstack.GetAuthResult(ctx, provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to get auth result: %v", err),
		})
	}

	// Extract token from auth result
	tokenObj, err := authResult.ExtractToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to extract token: %v", err),
		})
	}

	// Get the OpenStack token
	openstackToken := provider.Token()

	// Create JWT token
	jwtToken := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	expiresAt := tokenObj.ExpiresAt
	claims := jwtToken.Claims.(jwt.MapClaims)
	claims["username"] = req.Username
	claims["project_id"] = req.ProjectID
	claims["domain_name"] = domainName
	claims["exp"] = expiresAt.Unix()
	claims["openstack_token"] = openstackToken

	// Generate encoded token
	encodedToken, err := jwtToken.SignedString([]byte(h.JWTSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to generate token",
		})
	}

	// Return token
	return c.JSON(models.ProjectScopeResponse{
		Token:     encodedToken,
		ProjectID: req.ProjectID,
		ExpiresAt: expiresAt,
	})
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	ctx := c.Context()

	// Parse request body
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" || req.Email == "" || req.Phone == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Name, email, phone, and password are required",
		})
	}

	// Validate email format
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid email format",
		})
	}

	// Validate phone format (E.164)
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(req.Phone) {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid phone format. Use E.164 format (e.g., +254712345678)",
		})
	}

	// Validate password strength
	if err := validatePasswordStrength(req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	// Check if email already exists
	exists, err := h.PostgresClient.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to check email existence",
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(models.ErrorResponse{
			Error: "Email already registered",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to hash password",
		})
	}

	// Get admin provider for user creation
	adminProvider, err := openstack.GetAdminProvider(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to get admin provider: %v", err),
		})
	}

	// Create OpenStack user
	openstackUser, err := openstack.CreateUser(ctx, adminProvider, req.Email, req.Email, req.Password, "default")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create OpenStack user: %v", err),
		})
	}

	// Create OpenStack project
	projectName := fmt.Sprintf("lineserve-project-%s", strings.Split(uuid.New().String(), "-")[0])
	project, err := openstack.CreateProject(ctx, adminProvider, projectName, fmt.Sprintf("Project for %s", req.Email), "default")
	if err != nil {
		// Cleanup: Delete the OpenStack user if project creation fails
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to create OpenStack project: %v", err),
		})
	}

	// Assign member role to user for the project
	err = openstack.AssignRoleToUserOnProject(ctx, adminProvider, openstackUser.ID, project.ID, h.MemberRoleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to assign role: %v", err),
		})
	}

	// Insert user into database
	user := struct {
		Name            string
		Email           string
		Phone           string
		PasswordHash    string
		OpenstackUserID string
	}{
		Name:            req.Name,
		Email:           req.Email,
		Phone:           req.Phone,
		PasswordHash:    string(hashedPassword),
		OpenstackUserID: openstackUser.ID,
	}

	userID, err := h.PostgresClient.InsertUser(ctx, &user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create user",
		})
	}

	// Associate user with project
	_, err = h.PostgresClient.AssociateUserWithProject(ctx, userID, project.ID, h.MemberRoleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to associate user with project: %v", err),
		})
	}

	// Generate verification token
	verificationToken := uuid.New().String()
	expiresAt := time.Now().Add(time.Hour * 24)

	// Insert verification token
	_, err = h.PostgresClient.InsertEmailVerification(ctx, userID, verificationToken, expiresAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create verification token",
		})
	}

	// Send verification email (async)
	go h.sendVerificationEmail(req.Email, verificationToken)

	// Return success
	return c.Status(fiber.StatusCreated).JSON(models.RegisterResponse{
		ID:        userID,
		Email:     req.Email,
		ProjectID: project.ID,
		Message:   "User registered successfully. Please check your email to verify your account.",
	})
}

// validatePasswordStrength validates password strength
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// sendVerificationEmail sends a verification email
func (h *AuthHandler) sendVerificationEmail(email, token string) {
	// This is a placeholder for sending verification emails
	// In a real implementation, you would use an email service
	fmt.Printf("Sending verification email to %s with token %s\n", email, token)
}

// ListProjects lists all projects for the authenticated user
func (h *AuthHandler) ListProjects(c *fiber.Ctx) error {
	// Get user claims from JWT
	claims := c.Locals("user").(jwt.MapClaims)
	userID, _ := claims["user_id"].(string)
	username, _ := claims["username"].(string)
	domainName, _ := claims["domain_name"].(string)

	if userID == "" || username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid token: missing user information",
		})
	}

	// Create context
	ctx := c.Context()

	// Try to get user's projects from database first
	dbProjects, err := h.PostgresClient.GetUserProjects(ctx, userID)
	if err == nil && len(dbProjects) > 0 {
		// Convert to model projects
		projects := make([]models.Project, len(dbProjects))
		for i, p := range dbProjects {
			projects[i] = models.Project{
				ID:       p.ID,
				Name:     p.Name,
				DomainID: p.DomainID,
			}
		}

		// Return projects from database
		return c.JSON(models.ProjectListResponse{
			Projects: projects,
		})
	}

	// If no projects found in database or error occurred, try OpenStack
	// Re-authenticate with OpenStack to get fresh token
	provider, _, err := openstack.AuthenticateUnscoped(ctx, username, "", domainName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to authenticate with OpenStack: %v", err),
		})
	}

	// List projects for the user
	openstackProjects, err := openstack.ListUserProjects(ctx, provider, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Failed to list projects: %v", err),
		})
	}

	// Save projects to database for future use
	for _, p := range openstackProjects {
		// Save project
		err := h.PostgresClient.SaveProject(ctx, struct {
			ID          string
			Name        string
			Description string
			DomainID    string
			Enabled     bool
		}{
			ID:       p.ID,
			Name:     p.Name,
			DomainID: p.DomainID,
			Enabled:  true,
		})
		if err != nil {
			// Log error but continue
			fmt.Printf("Failed to save project %s: %v\n", p.ID, err)
		}

		// Associate user with project if not already associated
		_, err = h.PostgresClient.AssociateUserWithProject(ctx, userID, p.ID, h.MemberRoleID)
		if err != nil {
			// Log error but continue
			fmt.Printf("Failed to associate user %s with project %s: %v\n", userID, p.ID, err)
		}
	}

	// Convert OpenStack projects to model projects
	projects := make([]models.Project, len(openstackProjects))
	for i, p := range openstackProjects {
		projects[i] = models.Project{
			ID:       p.ID,
			Name:     p.Name,
			DomainID: p.DomainID,
		}
	}

	return c.JSON(models.ProjectListResponse{
		Projects: projects,
	})
}
