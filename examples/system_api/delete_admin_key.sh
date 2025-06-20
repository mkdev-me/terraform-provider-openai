#!/bin/bash
###############################################################################
# OpenAI Admin API Key Deletion Script
###############################################################################
# This script allows listing and deleting OpenAI Admin API keys by making
# direct API calls to the OpenAI API.
#
# Usage:
#   ./delete_admin_key.sh list                  - List all available Admin keys
#   ./delete_admin_key.sh <key_id>              - Delete a specific key by ID
#
# Arguments:
#   list                   - Command to list all available keys
#   key_id                 - ID of the key to delete (format: key_XXXX)
#
# Examples:
#   ./delete_admin_key.sh list
#   ./delete_admin_key.sh key_abc123def456
#
# Requirements:
#   - curl: For making HTTP requests
#   - jq: For parsing JSON responses
#   - OPENAI_ADMIN_KEY environment variable set with valid admin key
#
# Security Notes:
#   - Deleting an API key is irreversible
#   - This script requires confirmation by typing "YES" in uppercase
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
  echo "To delete a key, run this script with the key ID:"
  echo "./delete_admin_key.sh key_XXXX"
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

# Verify confirmation with a safety prompt
echo "Are you sure you want to delete Admin Key with ID: $KEY_ID?"
echo "This action cannot be undone."
echo "Type 'YES' (all uppercase) to confirm:"
read -p "> " CONFIRM

if [ "$CONFIRM" != "YES" ]; then
  echo "Operation cancelled."
  exit 0
fi

echo "Deleting Admin Key with ID: $KEY_ID..."

# Make API request to delete the key
RESPONSE=$(curl -s -X DELETE "https://api.openai.com/v1/organization/admin_api_keys/$KEY_ID" \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
  -H "OpenAI-Organization: $OPENAI_ORGANIZATION_ID" \
  -H "Content-Type: application/json")

# Check for errors
if echo "$RESPONSE" | grep -q "error"; then
  echo "Error deleting Admin Key:"
  echo "$RESPONSE" | jq .
  
  ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message')
  if [[ "$ERROR_MSG" == *"insufficient permissions"* ]]; then
    echo ""
    echo "===== RECOMMENDED SOLUTION ====="
    echo "The API key used does not have permissions to delete Admin Keys."
    echo "You need an API key with the 'api.management.write' scope"
  fi
  
  exit 1
fi

# Check if response is empty (success)
if [ -z "$RESPONSE" ] || [ "$RESPONSE" == "{}" ]; then
  echo "Admin Key deleted successfully!"
  echo ""
  echo "To verify deletion, run: $0 list"
else
  echo "API Response:"
  echo "$RESPONSE" | jq .
fi 
