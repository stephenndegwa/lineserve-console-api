#!/bin/bash

# Get a token first (replace with your credentials)
echo "Getting token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_username",
    "password": "your_password",
    "domain_name": "your_domain"
  }')

# Extract token
TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -z "$TOKEN" ]; then
  echo "Failed to get token"
  echo "Response: $TOKEN_RESPONSE"
  exit 1
fi

echo "Token obtained"

# Check OpenStack client status
echo "Checking OpenStack client status..."
curl -s -X GET http://localhost:8080/api/v1/openstack-status \
  -H "Authorization: Bearer $TOKEN" | jq .

# Get networks
echo "Getting networks..."
curl -s -X GET http://localhost:8080/api/v1/networks \
  -H "Authorization: Bearer $TOKEN" | jq .

echo "Test complete" 