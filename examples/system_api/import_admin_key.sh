#!/bin/bash
###############################################################################
# OpenAI Admin API Key Import Script
###############################################################################
# This script lists existing OpenAI Admin API keys and imports them for use
# with Terraform by generating appropriate configuration files.
#
# Usage:
#   ./import_admin_key.sh list                  - List all available Admin keys
#   ./import_admin_key.sh <key_id>              - Import a specific key by ID
#
# Arguments:
#   list                   - Command to list all available keys
#   key_id                 - ID of the key to import (format: key_XXXX)
#
# Examples:
#   ./import_admin_key.sh list
#   ./import_admin_key.sh key_abc123def456
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

# Handle 'list' command to show all available keys
if [ "$1" == "list" ]; then
  echo "Listing all available Admin Keys..."
  
  # Make API request to list keys
  RESPONSE=$(curl -s -X GET "https://api.openai.com/v1/organization/admin_api_keys" \
    -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
    -H "OpenAI-Organization: $OPENAI_ORGANIZATION_ID" \
    -H "Content-Type: application/json")
  
  # Check for errors
  if echo "$RESPONSE" | grep -q "error"; then
    echo "Error listing Admin Keys:"
    ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message')
    echo "$RESPONSE" | jq .
    
    if [[ "$ERROR_MSG" == *"insufficient permissions"* ]]; then
      echo ""
      echo "===== RECOMMENDED SOLUTION ====="
      echo "The API key used does not have permissions to list Admin Keys."
      echo "You need an API key with the 'api.management.read' scope"
    fi
    
    exit 1
  fi
  
  # Show formatted list of keys
  echo "===== AVAILABLE ADMIN KEYS ====="
  echo "$RESPONSE" | jq -r '.data[] | "ID: \(.id) | Name: \(.name) | Created: \(.created_at)"'
  
  # Provide instructions for using the listed keys
  echo ""
  echo "To import a key, run this script with the key ID:"
  echo "./import_admin_key.sh key_XXXX"
  exit 0
fi

# Check if key ID is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <key_id | list>"
  echo "Example: $0 key_abc123def456"
  echo "Or to list all keys: $0 list"
  exit 1
fi

KEY_ID="$1"

# Validate key ID format
if [[ ! $KEY_ID =~ ^key_ ]]; then
  echo "Error: Invalid key ID format. Key ID should start with 'key_'"
  echo "Run '$0 list' to see available keys with their IDs"
  exit 1
fi

echo "Looking up Admin Key with ID: $KEY_ID"

# Make API request to get key details
RESPONSE=$(curl -s -X GET "https://api.openai.com/v1/organization/admin_api_keys/$KEY_ID" \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
  -H "OpenAI-Organization: $OPENAI_ORGANIZATION_ID" \
  -H "Content-Type: application/json")

# Check for errors
if echo "$RESPONSE" | grep -q "error"; then
  echo "Error retrieving Admin Key:"
  echo "$RESPONSE" | jq .
  exit 1
fi

# Show information about the imported API key
echo "Admin Key found!"
echo "ID: $(echo "$RESPONSE" | jq -r .id)"
echo "Name: $(echo "$RESPONSE" | jq -r .name)"
echo "Created: $(echo "$RESPONSE" | jq -r .created_at | xargs -I{} date -r {} "+%Y-%m-%d %H:%M:%S" 2>/dev/null || echo "$(echo "$RESPONSE" | jq -r .created_at)")"

# Generate a safe resource name from the key name
KEY_NAME=$(echo "$RESPONSE" | jq -r .name | tr ' ' '_' | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9_]//g')
if [ -z "$KEY_NAME" ] || [ "$KEY_NAME" == "null" ]; then
  KEY_NAME="admin_key_${KEY_ID/key_/}"
fi

# Generate Terraform file to import the key
IMPORT_FILE="import_${KEY_NAME}.tf"

cat > "$IMPORT_FILE" << EOL
# Import of Admin Key: $(echo "$RESPONSE" | jq -r .name)
# ID: $(echo "$RESPONSE" | jq -r .id)
# Created: $(echo "$RESPONSE" | jq -r .created_at | xargs -I{} date -r {} "+%Y-%m-%d %H:%M:%S" 2>/dev/null || echo "$(echo "$RESPONSE" | jq -r .created_at)")

terraform {
  required_providers {
    openai = {
      source  = "local/openai"
    }
  }
}

# Resource definition for import
resource "openai_admin_api_key" "${KEY_NAME}" {
  name = "$(echo "$RESPONSE" | jq -r .name)"
  # NOTE: The following fields will be populated after import
  # expires_at = ...
  # scopes = ...
}

# Outputs to display key information
output "${KEY_NAME}_info" {
  value = {
    id         = openai_admin_api_key.${KEY_NAME}.id
    name       = openai_admin_api_key.${KEY_NAME}.name
    created_at = openai_admin_api_key.${KEY_NAME}.created_at
  }
}
EOL

echo ""
echo "Terraform file generated: $IMPORT_FILE"
echo ""
echo "To import the key into Terraform, run:"
echo "terraform init"
echo "terraform import openai_admin_api_key.${KEY_NAME} ${KEY_ID}"
echo ""
echo "NOTE: The API does not allow retrieving the original key value,"
echo "so that value is not included in the imported state."

# Save response in case it's useful
echo "$RESPONSE" > "adminkey_${KEY_NAME}_details.json"
echo "Complete details saved to adminkey_${KEY_NAME}_details.json" 
