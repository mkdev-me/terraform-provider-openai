#!/bin/bash
###############################################################################
# OpenAI Admin API Key Import Script
###############################################################################
# This script demonstrates how to import an existing OpenAI admin API key into
# Terraform state.
#
# Usage:
#   ./import.sh <admin_api_key_id>
#
# Arguments:
#   admin_api_key_id - ID of the admin API key to import (format: key_XXXX)
#
# Example:
#   ./import.sh key_abc123def456
#
# Requirements:
#   - Terraform must be initialized in this directory
#   - OPENAI_ADMIN_KEY environment variable must be set
#
###############################################################################

# Check if admin API key ID is provided
if [ $# -ne 1 ]; then
  echo "Error: Missing admin API key ID"
  echo "Usage: $0 <admin_api_key_id>"
  echo "Example: $0 key_abc123def456"
  exit 1
fi

ADMIN_API_KEY_ID="$1"

# Validate admin API key ID format
if [[ ! $ADMIN_API_KEY_ID =~ ^key_ ]]; then
  echo "Error: Invalid admin API key ID format. Expected format: key_XXXX"
  exit 1
fi

# Check if OPENAI_ADMIN_KEY is set
if [ -z "$OPENAI_ADMIN_KEY" ]; then
  echo "Error: OPENAI_ADMIN_KEY environment variable not set"
  echo "Please set the OPENAI_ADMIN_KEY environment variable with your OpenAI admin API key."
  exit 1
fi

echo "Importing OpenAI Admin API Key: $ADMIN_API_KEY_ID"
echo ""

# Create a temporary import configuration
cat > import_admin_api_key.tf << EOF
# Temporary configuration for importing admin API key
resource "openai_admin_api_key" "imported" {
  name = "Imported Admin Key"
}
EOF

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Import the admin API key
echo "Importing admin API key..."
terraform import openai_admin_api_key.imported "$ADMIN_API_KEY_ID"

if [ $? -eq 0 ]; then
  echo ""
  echo "Admin API key imported successfully!"
  echo ""
  echo "Next steps:"
  echo "1. Run 'terraform show' to see the imported admin API key details"
  echo "2. Update resource.tf with the actual admin API key configuration"
  echo "3. Run 'terraform plan' to verify no changes are needed"
  echo "4. Remove import_admin_api_key.tf when done"
  echo ""
  echo "Note: The API key value itself cannot be retrieved after creation."
else
  echo ""
  echo "Import failed. Please check the admin API key ID and your permissions."
fi

# Clean up temporary file
rm -f import_admin_api_key.tf