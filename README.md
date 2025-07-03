# Lineserve API

A Go-based REST API for Lineserve, a public cloud company powered by OpenStack.

## Features

- Authentication with JWT
- Instance (VM) management
- Image management
- Flavor management
- Network management
- Volume management
- Project management

## Prerequisites

- Go 1.20 or higher
- OpenStack credentials
- Access to an OpenStack environment

## Installation

1. Clone the repository:

```bash
git clone https://github.com/lineserve/lineserve-api.git
cd lineserve-api
```

2. Copy the sample environment file and update it with your OpenStack credentials:

```bash
cp env.sample .env
```

3. Update the `.env` file with your OpenStack credentials and JWT secret.

4. Build the application:

```bash
go build -o lineserve-api
```

## Usage

### Running the API

```bash
./lineserve-api
```

The API will start on port 8080 by default, or the port specified in the `.env` file.

### API Endpoints

#### Authentication

- `POST /api/login` - Authenticate and get a JWT token

#### Instances

- `GET /api/instances` - List all instances
- `POST /api/instances` - Create a new instance
- `GET /api/instances/:id` - Get instance details

#### Images

- `GET /api/images` - List all images
- `GET /api/images/:id` - Get image details

#### Flavors

- `GET /api/flavors` - List all flavors

#### Networks

- `GET /api/networks` - List all networks
- `GET /api/networks/:id` - Get network details

#### Volumes

- `GET /api/volumes` - List all volumes
- `POST /api/volumes` - Create a new volume
- `GET /api/volumes/:id` - Get volume details

#### Projects

- `GET /api/projects` - List all projects
- `GET /api/projects/:id` - Get project details

## Authentication

To access protected endpoints, include the JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

You can obtain a token by making a POST request to `/api/login` with your username and password.

## Environment Variables

The following environment variables are required:

```
# API Configuration
PORT=8080
JWT_SECRET=your-jwt-secret-key

# OpenStack Configuration
OS_AUTH_URL=http://your-openstack-auth-url
OS_USERNAME=your-username
OS_PASSWORD=your-password
OS_PROJECT_ID=your-project-id
OS_PROJECT_NAME=your-project-name
OS_USER_DOMAIN_NAME=Default
OS_REGION_NAME=your-region-name
```

## Development

### Project Structure

- `cmd/` - Command-line tools
- `docs/` - Documentation
- `internal/` - Internal packages
- `pkg/` - Public packages
  - `client/` - OpenStack client
  - `config/` - Configuration
  - `handlers/` - API handlers
  - `middleware/` - Middleware
  - `models/` - Data models

### Running Tests

```bash
go test ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 