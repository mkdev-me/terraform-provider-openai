#!/bin/bash
###############################################################################
# OpenAI Service Account Import Script
###############################################################################
# This script demonstrates how to import an existing OpenAI service account into
# Terraform state.
#
# Usage:
#   ./import.sh <service_account_id>
#
# Arguments:
#   service_account_id - ID of the service account to import (format: svc_XXXX)
#
# Example:
#   ./import.sh svc_abc123def456
#
# Requirements:
#   - Terraform must be initialized in this directory
#   - OPENAI_ADMIN_KEY environment variable must be set
#
###############################################################################

# Check if service account ID is provided
if [ $# -ne 1 ]; then
  echo "Error: Missing service account ID"
  echo "Usage: $0 <service_account_id>"
  echo "Example: $0 svc_abc123def456"
  exit 1
fi

SERVICE_ACCOUNT_ID="$1"

# Validate service account ID format
if [[ ! $SERVICE_ACCOUNT_ID =~ ^svc_ ]]; then
  echo "Error: Invalid service account ID format. Expected format: svc_XXXX"
  exit 1
fi

# Check if OPENAI_ADMIN_KEY is set
if [ -z "$OPENAI_ADMIN_KEY" ]; then
  echo "Error: OPENAI_ADMIN_KEY environment variable not set"
  echo "Please set the OPENAI_ADMIN_KEY environment variable with your OpenAI admin API key."
  exit 1
fi

echo "Importing OpenAI Service Account: $SERVICE_ACCOUNT_ID"
echo ""

# Create a temporary import configuration
cat > import_service_account.tf << EOF
# Temporary configuration for importing service account
resource "openai_service_account" "imported" {
  name = "Imported Service Account"
}
EOF

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Import the service account
echo "Importing service account..."
terraform import openai_service_account.imported "$SERVICE_ACCOUNT_ID"

if [ $? -eq 0 ]; then
  echo ""
  echo "Service account imported successfully!"
  echo ""
  echo "Next steps:"
  echo "1. Run 'terraform show' to see the imported service account details"
  echo "2. Update resource.tf with the actual service account configuration"
  echo "3. Run 'terraform plan' to verify no changes are needed"
  echo "4. Remove import_service_account.tf when done"
else
  echo ""
  echo "Import failed. Please check the service account ID and your permissions."
fi

# Clean up temporary file
rm -f import_service_account.tf