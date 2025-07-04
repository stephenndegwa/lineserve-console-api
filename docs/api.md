# Lineserve API Documentation

This document provides detailed information about the Lineserve API endpoints.

## Authentication

### Login

Authenticates a user against OpenStack and returns a JWT token.

**URL**: `/login`

**Method**: `POST`

**Auth required**: No

**Request Body**:

```json
{
  "username": "string",
  "password": "string"
}
```

**Success Response**:

- **Code**: 200 OK
- **Content**:

```json
{
  "token": "string"
}
```

**Error Response**:

- **Code**: 401 Unauthorized
- **Content**:

```json
{
  "error": "Invalid credentials"
}
```

## Instances

### List Instances

Returns a list of all instances.

**URL**: `/api/instances`

**Method**: `GET`

**Auth required**: Yes (JWT)

**Success Response**:

- **Code**: 200 OK
- **Content**:

```json
[
  {
    "id": "string",
    "name": "string",
    "status": "string",
    "flavor": "string",
    "image": "string",
    "addresses": {
      "network_name": [
        {
          "type": "string",
          "address": "string"
        }
      ]
    },
    "created": "timestamp"
  }
]
```

### Create Instance

Creates a new instance.

**URL**: `/api/instances`

**Method**: `POST`

**Auth required**: Yes (JWT)

**Request Body**:

```json
{
  "name": "string",
  "flavor_id": "string",
  "image_id": "string",
  "network_id": "string",
  "key_name": "string" // optional
}
```

**Success Response**:

- **Code**: 201 Created
- **Content**:

```json
{
  "id": "string",
  "name": "string",
  "status": "string",
  "created": "timestamp"
}
```

### Get Instance

Returns details for a specific instance.

**URL**: `/api/instances/:id`

**Method**: `GET`

**Auth required**: Yes (JWT)

**URL Parameters**:

- `id`: Instance ID

**Success Response**:

- **Code**: 200 OK
- **Content**:

```json
{
  "id": "string",
  "name": "string",
  "status": "string",
  "flavor": "string",
  "image": "string",
  "addresses": {
    "network_name": [
      {
        "type": "string",
        "address": "string"
      }
    ]
  },
  "created": "timestamp"
}
```

## Images

### List Images

Returns a list of all images.

**URL**: `/api/images`

**Method**: `GET`

**Auth required**: Yes (JWT)

**Success Response**:

- **Code**: 200 OK
- **Content**:

```json
[
  {
    "id": "string",
    "name": "string",
    "status": "string",
    "size": "number",
    "visibility": "string",
    "tags": ["string"],
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "properties": {
      "key": "value"
    }
  }
]
```

### Get Image

Returns details for a specific image.

**URL**: `/api/images/:id`

**Method**: `GET`

**Auth required**: Yes (JWT)

**URL Parameters**:

- `id`: Image ID

**Success Response**:

- **Code**: 200 OK
- **Content**:

```json
{
  "id": "string",
  "name": "string",
  "status": "string",
  "size": "number",
  "visibility": "string",
  "tags": ["string"],
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "properties": {
    "key": "value"
  }
}
```

## Flavors

### List Flavors

Returns a list of all flavors.

**URL**: `/api/flavors`

**Method**: `GET`

**Auth required**: Yes (JWT)

**Success Response**:

- **Code**: 200 OK
- **Content**:

```json
[
  {
    "id": "string",
    "name": "string",
    "ram": "number",
    "disk": "number",
    "vcpus": "number",
    "is_public": "boolean"
  }
]
```

## Networks

- `GET /api/v1/networks` - List all networks
- `GET /api/v1/networks/:id` - Get network details
- `POST /api/v1/networks` - Create a new network
- `DELETE /api/v1/networks/:id` - Delete a network

### List Networks

**Curl Example:**

```bash
curl -X GET http://localhost:8080/api/v1/networks \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Get Network

**Curl Example:**

```bash
curl -X GET http://localhost:8080/api/v1/networks/NETWORK_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Create Network

**Request:**

```json
{
  "name": "my-network",
  "shared": false,
  "external": false,
  "admin_state_up": true
}
```

**Curl Example:**

```bash
curl -X POST http://localhost:8080/api/v1/networks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "my-network",
    "shared": false,
    "external": false,
    "admin_state_up": true
  }'
```

### Delete Network

**Curl Example:**

```bash
curl -X DELETE http://localhost:8080/api/v1/networks/NETWORK_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Volumes

### List Volumes

Returns a list of all volumes.

**URL**: `/api/volumes`

**Method**: `GET`

**Auth required**: Yes (JWT)

**Success Response**:

- **Code**: 200 OK
- **Content**:

```json
[
  {
    "id": "string",
    "name": "string",
    "status": "string",
    "size": "number",
    "volume_type": "string",
    "availability_zone": "string",
    "created_at": "timestamp",
    "attachments": [
      {
        "server_id": "string",
        "attachment_id": "string",
        "device_name": "string"
      }
    ]
  }
]
```

### Create Volume

Creates a new volume.

**URL**: `/api/volumes`

**Method**: `POST`

**Auth required**: Yes (JWT)

**Request Body**:

```json
{
  "name": "string",
  "size": "number",
  "volume_type": "string", // optional
  "availability_zone": "string" // optional
}
```

**Success Response**:

- **Code**: 201 Created
- **Content**:

```json
{
  "id": "string",
  "name": "string",
  "status": "string",
  "size": "number",
  "volume_type": "string",
  "availability_zone": "string",
  "created_at": "timestamp"
}
```

### Get Volume

Returns details for a specific volume.

**URL**: `/api/volumes/:id`

**Method**: `GET`

**Auth required**: Yes (JWT)

**URL Parameters**:

- `id`: Volume ID

**Success Response**:

- **Code**: 200 OK
- **Content**:

```json
{
  "id": "string",
  "name": "string",
  "status": "string",
  "size": "number",
  "volume_type": "string",
  "availability_zone": "string",
  "created_at": "timestamp",
  "attachments": [
    {
      "server_id": "string",
      "attachment_id": "string",
      "device_name": "string"
    }
  ]
}
```

## Projects

### List Projects

Returns a list of all projects.

**URL**: `/api/projects`

**Method**: `GET`

**Auth required**: Yes (JWT)

**Success Response**:

- **Code**: 200 OK
- **Content**:

```json
[
  {
    "id": "string",
    "name": "string",
    "description": "string",
    "enabled": "boolean",
    "domain_id": "string"
  }
]
```

### Get Project

Returns details for a specific project.

**URL**: `/api/projects/:id`

**Method**: `GET`

**Auth required**: Yes (JWT)

**URL Parameters**:

- `id`: Project ID

**Success Response**:

- **Code**: 200 OK
- **Content**:

```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "enabled": "boolean",
  "domain_id": "string"
}
```

## Key Pairs

- `GET /api/v1/keypairs` - List all key pairs
- `GET /api/v1/keypairs/:name` - Get key pair details
- `POST /api/v1/keypairs` - Create a new key pair
- `DELETE /api/v1/keypairs/:name` - Delete a key pair

### Create Key Pair

**Request:**

```json
{
  "name": "my-keypair",
  "public_key": "ssh-rsa AAAAB3NzaC1yc2EAAA..." // Optional, if not provided a new key pair will be generated
}
```

**Response:**

```json
{
  "name": "my-keypair",
  "fingerprint": "7e:eb:ab:24:ba:d1:e1:88:ae:9a:fb:66:53:df:d3:bd",
  "public_key": "ssh-rsa AAAAB3NzaC1yc2EAAA...",
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAw2..." // Only returned when creating a new key pair
}
``` 