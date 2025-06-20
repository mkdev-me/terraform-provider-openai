#!/bin/bash
###############################################################################
# OpenAI Admin API Key Creation Script
###############################################################################
# This script creates a new OpenAI Admin API key by making a direct API call.
# 
# Usage:
#   ./create_admin_key.sh <key_name> [expiration_timestamp] [scope1,scope2,...]
#
# Arguments:
#   key_name             - The name for the new API key (required)
#   expiration_timestamp - Unix timestamp when the key should expire (optional)
#   scopes               - Comma-separated list of permission scopes (optional)
#
# Examples:
#   ./create_admin_key.sh "terraform-admin-key"
#   ./create_admin_key.sh "temp-key" 1772382952
#   ./create_admin_key.sh "read-only" null "api.management.read"
#
# Requirements:
#   - curl: For making HTTP requests
#   - jq: For parsing JSON responses
#   - OPENAI_ADMIN_KEY environment variable set with valid admin key
#
###############################################################################

# Check if API key is configured
if [ -z "$OPENAI_ADMIN_KEY" ]; then
  echo "Error: OPENAI_ADMIN_KEY environment variable is not configured"
  echo "Run: export OPENAI_ADMIN_KEY=sk-admin-..."
  exit 1
fi

# Check if Organization ID is configured
if [ -z "$OPENAI_ORGANIZATION_ID" ]; then
  echo "Warning: OPENAI_ORGANIZATION_ID environment variable is not configured"
  echo "Recommended: export OPENAI_ORGANIZATION_ID=org-..."
fi

# Check if at least one parameter for the name is provided
if [ $# -lt 1 ]; then
  echo "Usage: $0 <key_name> [expiration_timestamp] [scope1,scope2,...]"
  echo "Example: $0 'admin-key' 1772382952 'api.management.read,api.management.write'"
  exit 1
fi

# Read parameters
KEY_NAME="$1"
EXPIRES_AT="$2"
SCOPES="$3"

# Build JSON payload
PAYLOAD="{"
PAYLOAD+="\"name\": \"$KEY_NAME\""

# Add expiration date if provided
if [ ! -z "$EXPIRES_AT" ] && [ "$EXPIRES_AT" != "null" ]; then
  PAYLOAD+=", \"expires_at\": $EXPIRES_AT"
fi

# Add scopes if provided
if [ ! -z "$SCOPES" ]; then
  # Convert comma-separated list to JSON array
  IFS=',' read -r -a SCOPE_ARRAY <<< "$SCOPES"
  PAYLOAD+=", \"scopes\": ["
  
  FIRST=true
  for SCOPE in "${SCOPE_ARRAY[@]}"; do
    if [ "$FIRST" = true ]; then
      FIRST=false
    else
      PAYLOAD+=", "
    fi
    PAYLOAD+="\"$SCOPE\""
  done
  
  PAYLOAD+="]"
fi

PAYLOAD+="}"

echo "Creating Admin Key with name '$KEY_NAME'..."
echo "Payload: $PAYLOAD"

# Make API request
RESPONSE=$(curl -s -X POST "https://api.openai.com/v1/organization/admin_api_keys" \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
  -H "OpenAI-Organization: $OPENAI_ORGANIZATION_ID" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD")

# Check for errors
if echo "$RESPONSE" | grep -q "error"; then
  echo "Error creating Admin Key:"
  ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message')
  ERROR_TYPE=$(echo "$RESPONSE" | jq -r '.error.type')
  
  echo "$RESPONSE" | jq .
  
  # Provide advice based on error type
  if [[ "$ERROR_MSG" == *"insufficient permissions"* ]]; then
    echo ""
    echo "===== RECOMMENDED SOLUTION ====="
    echo "The API key used does not have permissions to create Admin Keys."
    echo "You need an API key with the 'api.management.write' scope"
    echo ""
    echo "1. Go to https://platform.openai.com/api-keys"
    echo "2. Create a new Admin key with full permissions"
    echo "3. Set the environment variable: export OPENAI_ADMIN_KEY=sk-admin-..."
    echo "4. Try running this script again"
  fi
  
  exit 1
fi

# Show information about the created API key
echo "Admin Key created successfully!"
echo "ID: $(echo "$RESPONSE" | jq -r .id)"
echo "Name: $(echo "$RESPONSE" | jq -r .name)"
echo "Created: $(echo "$RESPONSE" | jq -r .created_at | xargs -I{} date -r {} "+%Y-%m-%d %H:%M:%S" 2>/dev/null || echo "$(echo "$RESPONSE" | jq -r .created_at)")"
echo "Value: $(echo "$RESPONSE" | jq -r .value)"
echo ""
echo "IMPORTANT: Store this value in a secure location. It will not be shown again."

# Save the response to a file for reference
echo "$RESPONSE" > "adminkey_${KEY_NAME}_response.json"
echo "Complete response saved to adminkey_${KEY_NAME}_response.json"
echo ""
echo "SECURITY NOTE: This file contains sensitive information."
echo "Delete it when no longer needed with: rm adminkey_${KEY_NAME}_response.json" 
