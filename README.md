# Lineserve Cloud API

This is the backend API for Lineserve Cloud, providing user registration, authentication, and cloud resource management through OpenStack using Gophercloud v2.7.0.

## Features

- User registration with OpenStack integration
- JWT-based authentication
- Multi-tenant support with project scoping
- OpenStack resource management (instances, images, volumes, networks)
- Email verification (stubbed)
- PostgreSQL integration for user management

## Prerequisites

- Go 1.22 or higher
- PostgreSQL database
- OpenStack cloud environment with Keystone v3

## Environment Variables

Copy the `env.sample` file to `.env` and update the values:

```bash
cp env.sample .env
```

Required environment variables:

- `API_PORT`: API server port (default: 3075)
- `JWT_SECRET`: Secret key for JWT token generation
- `OS_AUTH_URL`: OpenStack Keystone authentication URL
- `OS_USERNAME`: OpenStack admin username
- `OS_PASSWORD`: OpenStack admin password
- `OS_PROJECT_ID`: OpenStack admin project ID
- `OS_PROJECT_NAME`: OpenStack admin project name
- `OS_DOMAIN_NAME`: OpenStack domain name (default: Default)
- `OS_USER_DOMAIN_NAME`: OpenStack user domain name (default: Default)
- `OS_PROJECT_DOMAIN_NAME`: OpenStack project domain name (default: Default)
- `OS_REGION_NAME`: OpenStack region name
- `OS_IDENTITY_API_VERSION`: OpenStack identity API version (default: 3)
- `OS_INTERFACE`: OpenStack interface type (default: public)
- `OPENSTACK_MEMBER_ROLE_ID`: OpenStack member role ID
- `POSTGRES_HOST`: PostgreSQL host
- `POSTGRES_PORT`: PostgreSQL port
- `POSTGRES_USER`: PostgreSQL user
- `POSTGRES_PASSWORD`: PostgreSQL password
- `POSTGRES_DB`: PostgreSQL database name
- `POSTGRES_SSLMODE`: PostgreSQL SSL mode

## PostgreSQL Setup

Create the following tables in your PostgreSQL database:

```sql
-- Create users table
CREATE TABLE lineserve_cloud_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    openstack_user_id TEXT,
    openstack_project_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified BOOLEAN DEFAULT FALSE
);

-- Create email verifications table
CREATE TABLE lineserve_cloud_email_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES lineserve_cloud_users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create user projects table
CREATE TABLE lineserve_cloud_user_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES lineserve_cloud_users(id) ON DELETE CASCADE,
    project_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Installation

```bash
# Clone the repository
git clone https://github.com/lineserve/lineserve-api.git
cd lineserve-api

# Install dependencies
go mod tidy

# Build the application
go build -o lineserve-api

# Run the application
./lineserve-api
```

## API Endpoints

### Authentication Endpoints

- `POST /v1/login`: Authenticate a user with unscoped token
  - Returns a JWT token and list of available projects

- `POST /v1/projects/token`: Get a project-scoped token
  - Returns a JWT token scoped to the specified project

- `POST /v1/register`: Register a new user
  - Creates a user in the database
  - Creates an OpenStack user
  - Creates an OpenStack project
  - Assigns the member role to the user for the project

### Protected Endpoints (require project-scoped JWT token)

- `GET /v1/instances`: List instances
- `POST /v1/instances`: Create an instance
- `GET /v1/instances/:id`: Get instance details
- `DELETE /v1/instances/:id`: Delete an instance
- `GET /v1/images`: List images
- `GET /v1/images/:id`: Get image details
- `GET /v1/flavors`: List flavors
- `GET /v1/networks`: List networks
- `GET /v1/networks/:id`: Get network details
- `GET /v1/volumes`: List volumes
- `POST /v1/volumes`: Create a volume
- `GET /v1/volumes/:id`: Get volume details

## OpenStack Integration

This API uses Gophercloud v2.7.0 for OpenStack integration with the following features:

- Unscoped authentication for initial login
- Project-scoped authentication for resource operations
- Context-aware service client creation
- Proper token extraction and validation
- Multi-tenant isolation through project scoping

## License

MIT 