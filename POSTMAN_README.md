# LineServe API Postman Collection

This repository contains a comprehensive Postman collection for the LineServe API, which provides cloud infrastructure management capabilities through OpenStack.

## Files

- `LineServe-API.postman_collection.json`: The Postman collection containing all API endpoints
- `LineServe-API-Environment.postman_environment.json`: The Postman environment variables

## Getting Started

### Prerequisites

- [Postman](https://www.postman.com/downloads/) installed on your machine
- LineServe API server running (default: http://localhost:8080)

### Importing the Collection and Environment

1. Open Postman
2. Click on "Import" button in the top-left corner
3. Select the `LineServe-API.postman_collection.json` and `LineServe-API-Environment.postman_environment.json` files
4. Both the collection and environment will be imported into Postman

### Setting Up the Environment

1. In Postman, click on the "Environment" dropdown in the top-right corner
2. Select "LineServe API Environment"
3. Update the `baseUrl` variable if your API is running on a different URL

## Using the Collection

The collection is organized into several folders, each containing related API endpoints:

### Authentication

1. **Register**: Create a new user account
   - Update the request body with your desired user information
   - Send the request to register a new user

2. **Login**: Authenticate with your credentials
   - Update the request body with your username (email), password, and domain name
   - Send the request to get an authentication token
   - The token will be automatically saved to the `authToken` environment variable

3. **Get Project Token**: Get a project-scoped token
   - After logging in, use this request to get a project-scoped token
   - The token will be automatically saved to the `projectToken` environment variable

### Projects

- **List Projects**: View all projects accessible to your user
- **Get Project**: Get details of a specific project

### Instances

- **List Instances**: View all compute instances in your project
- **Create Instance**: Create a new compute instance
- **Get Instance**: Get details of a specific instance

### Images

- **List Images**: View all available images
- **Get Image**: Get details of a specific image

### Flavors

- **List Flavors**: View all available instance flavors

### Networks

- **List Networks**: View all available networks
- **Get Network**: Get details of a specific network

### Volumes

- **List Volumes**: View all volumes in your project
- **Create Volume**: Create a new volume
- **Get Volume**: Get details of a specific volume

## Authentication Flow

1. Register a new user using the **Register** endpoint
2. Login using the **Login** endpoint (saves `authToken`)
3. Note the `projects` array in the response and choose a project ID
4. Get a project-scoped token using the **Get Project Token** endpoint (saves `projectToken`)
5. Use the project-scoped token for all project-specific operations

## Environment Variables

The collection uses the following environment variables:

- `baseUrl`: The base URL of the LineServe API
- `authToken`: User authentication token (set automatically after login)
- `projectToken`: Project-scoped token (set automatically after getting a project token)
- `userId`: ID of the authenticated user
- `projectId`: ID of the selected project
- `instanceId`: ID of a compute instance (set manually)
- `imageId`: ID of an image (set manually)
- `flavorId`: ID of a flavor (set manually)
- `networkId`: ID of a network (set manually)
- `volumeId`: ID of a volume (set manually)

## Tips

- After listing resources (instances, images, etc.), manually set the corresponding ID variables in the environment for use in subsequent requests
- The collection includes test scripts that automatically set authentication tokens and IDs when possible
- For project-scoped operations, make sure to use the `projectToken` rather than the `authToken` 