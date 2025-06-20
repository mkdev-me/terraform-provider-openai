#!/bin/bash
###############################################################################
# OpenAI Admin API Key Test Script
###############################################################################
# This script demonstrates a complete workflow for managing OpenAI Admin API 
# keys through the following actions:
#   1. List existing admin keys
#   2. Create a new admin key with a unique name
#   3. List all keys again to verify the new key appears
#
# Usage:
#   ./test_admin_key.sh <admin_api_key>
#
# Arguments:
#   admin_api_key          - A valid OpenAI Admin API key with api.management
#                            read and write permissions
#
# Examples:
#   ./test_admin_key.sh sk-admin-xxxxxxxxxxxx
#
# Requirements:
#   - ./create_admin_key.sh and ./import_admin_key.sh scripts in same directory
#   - curl and jq installed
#
# Security Notes:
#   - Remember to delete test keys after running this script
#   - The script will create a file containing the key value
#
###############################################################################

# Set the API key from the parameter
if [ ! -z "$1" ]; then
  export OPENAI_ADMIN_KEY="$1"
  echo "Using provided API key"
else
  echo "Usage: $0 <admin_api_key>"
  echo "Example: $0 sk-admin-xxxxxxxxxxxx"
  exit 1
fi

# Function to display section headers
section() {
  echo ""
  echo "===== TEST $1: $2 ====="
  echo ""
}

# Validate prerequisites
if [ ! -f "./import_admin_key.sh" ] || [ ! -f "./create_admin_key.sh" ]; then
  echo "Error: Required scripts not found in current directory"
  echo "Make sure import_admin_key.sh and create_admin_key.sh are in the same directory"
  exit 1
fi

# Ensure scripts are executable
chmod +x ./import_admin_key.sh ./create_admin_key.sh ./delete_admin_key.sh 2>/dev/null

# Test 1: List existing keys
section "1" "Listing existing API keys"
./import_admin_key.sh list

# Test 2: Create a new key with timestamp for uniqueness
section "2" "Creating a new API key"
TIMESTAMP=$(date +%s)
KEY_NAME="test-admin-key-$TIMESTAMP"
./create_admin_key.sh "$KEY_NAME"

# Test 3: List keys again to see the newly created key
section "3" "Verifying new key appears in list"
./import_admin_key.sh list

# Clean-up instructions
section "CLEANUP" "Instructions"
echo "Tests completed. Check the output above to verify results."
echo ""
echo "IMPORTANT: Remember to delete test keys when you're done!"
echo "Run: ./delete_admin_key.sh <key_id>"
echo ""
echo "The test key was named: $KEY_NAME"
echo "Search for this name in the list above to find its key_id" 
