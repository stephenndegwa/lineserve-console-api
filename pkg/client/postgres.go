package client

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresClient represents a PostgreSQL client
type PostgresClient struct {
	DB *sql.DB
}

// NewPostgresClient creates a new PostgreSQL client
func NewPostgresClient(connStr string) (*PostgresClient, error) {
	// Connect to PostgreSQL
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}

	return &PostgresClient{
		DB: db,
	}, nil
}

// Close closes the database connection
func (c *PostgresClient) Close() error {
	return c.DB.Close()
}

// CreateTablesIfNotExist creates the necessary tables if they don't exist
func (c *PostgresClient) CreateTablesIfNotExist(ctx context.Context) error {
	// Create users table
	_, err := c.DB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS lineserve_cloud_users (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			phone TEXT,
			password_hash TEXT NOT NULL,
			openstack_user_id TEXT,
			created_at TIMESTAMP NOT NULL,
			verified BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Create projects table
	_, err = c.DB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS lineserve_cloud_projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			domain_id TEXT,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create projects table: %v", err)
	}

	// Create user_projects table for multi-tenant support
	_, err = c.DB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS lineserve_cloud_user_projects (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES lineserve_cloud_users(id),
			project_id TEXT NOT NULL,
			role_id TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create user_projects table: %v", err)
	}

	// Create email verifications table
	_, err = c.DB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS lineserve_cloud_email_verifications (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES lineserve_cloud_users(id),
			token TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create email verifications table: %v", err)
	}

	return nil
}

// CheckEmailExists checks if an email already exists in the database
func (c *PostgresClient) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := c.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM lineserve_cloud_users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if email exists: %v", err)
	}
	return exists, nil
}

// InsertUser inserts a new user into the database
func (c *PostgresClient) InsertUser(ctx context.Context, user *struct {
	Name            string
	Email           string
	Phone           string
	PasswordHash    string
	OpenstackUserID string
}) (string, error) {
	userID := uuid.New().String()
	createdAt := time.Now()

	_, err := c.DB.ExecContext(ctx, `
		INSERT INTO lineserve_cloud_users (id, name, email, phone, password_hash, openstack_user_id, created_at, verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, userID, user.Name, user.Email, user.Phone, user.PasswordHash, user.OpenstackUserID, createdAt, false)

	if err != nil {
		return "", fmt.Errorf("failed to insert user: %v", err)
	}

	return userID, nil
}

// AssociateUserWithProject associates a user with a project
func (c *PostgresClient) AssociateUserWithProject(ctx context.Context, userID, projectID, roleID string) (string, error) {
	id := uuid.New().String()
	createdAt := time.Now()

	_, err := c.DB.ExecContext(ctx, `
		INSERT INTO lineserve_cloud_user_projects (id, user_id, project_id, role_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, id, userID, projectID, roleID, createdAt)

	if err != nil {
		return "", fmt.Errorf("failed to associate user with project: %v", err)
	}

	return id, nil
}

// GetUserProjects gets all projects associated with a user
func (c *PostgresClient) GetUserProjects(ctx context.Context, userID string) ([]struct {
	ID       string
	Name     string
	DomainID string
}, error) {
	rows, err := c.DB.QueryContext(ctx, `
		SELECT up.project_id, COALESCE(p.name, 'Unknown Project'), COALESCE(p.domain_id, '')
		FROM lineserve_cloud_user_projects up
		LEFT JOIN lineserve_cloud_projects p ON up.project_id = p.id
		WHERE up.user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %v", err)
	}
	defer rows.Close()

	var projects []struct {
		ID       string
		Name     string
		DomainID string
	}

	for rows.Next() {
		var project struct {
			ID       string
			Name     string
			DomainID string
		}
		if err := rows.Scan(&project.ID, &project.Name, &project.DomainID); err != nil {
			return nil, fmt.Errorf("failed to scan user project: %v", err)
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user projects: %v", err)
	}

	return projects, nil
}

// InsertEmailVerification inserts a new email verification record
func (c *PostgresClient) InsertEmailVerification(ctx context.Context, userID, token string, expiresAt time.Time) (string, error) {
	verificationID := uuid.New().String()
	createdAt := time.Now()

	_, err := c.DB.ExecContext(ctx, `
		INSERT INTO lineserve_cloud_email_verifications (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, verificationID, userID, token, expiresAt, createdAt)

	if err != nil {
		return "", fmt.Errorf("failed to insert email verification: %v", err)
	}

	return verificationID, nil
}

// GetUserByID gets a user by ID
func (c *PostgresClient) GetUserByID(ctx context.Context, id string) (*struct {
	ID              string
	Name            string
	Email           string
	Phone           string
	OpenstackUserID string
	Verified        bool
}, error) {
	user := struct {
		ID              string
		Name            string
		Email           string
		Phone           string
		OpenstackUserID string
		Verified        bool
	}{}

	err := c.DB.QueryRowContext(ctx, `
		SELECT id, name, email, phone, openstack_user_id, verified
		FROM lineserve_cloud_users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.OpenstackUserID,
		&user.Verified,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

// GetUserByEmail gets a user by email
func (c *PostgresClient) GetUserByEmail(ctx context.Context, email string) (*struct {
	ID              string
	Name            string
	Email           string
	Phone           string
	PasswordHash    string
	OpenstackUserID string
	Verified        bool
}, error) {
	user := struct {
		ID              string
		Name            string
		Email           string
		Phone           string
		PasswordHash    string
		OpenstackUserID string
		Verified        bool
	}{}

	err := c.DB.QueryRowContext(ctx, `
		SELECT id, name, email, phone, password_hash, openstack_user_id, verified
		FROM lineserve_cloud_users
		WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.OpenstackUserID,
		&user.Verified,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

// SaveProject saves a project to the database
func (c *PostgresClient) SaveProject(ctx context.Context, project struct {
	ID          string
	Name        string
	Description string
	DomainID    string
	Enabled     bool
}) error {
	createdAt := time.Now()

	// Use upsert (insert or update)
	_, err := c.DB.ExecContext(ctx, `
		INSERT INTO lineserve_cloud_projects (id, name, description, domain_id, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			domain_id = EXCLUDED.domain_id,
			enabled = EXCLUDED.enabled
	`, project.ID, project.Name, project.Description, project.DomainID, project.Enabled, createdAt)

	if err != nil {
		return fmt.Errorf("failed to save project: %v", err)
	}

	return nil
}
