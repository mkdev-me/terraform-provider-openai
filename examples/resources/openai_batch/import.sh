#!/bin/bash
###############################################################################
# OpenAI Batch Import Script
###############################################################################
# This script demonstrates how to import an existing OpenAI batch into
# Terraform state.
#
# Usage:
#   ./import.sh <batch_id>
#
# Arguments:
#   batch_id - ID of the batch to import (format: batch_XXXX)
#
# Example:
#   ./import.sh batch_abc123def456
#
# Requirements:
#   - Terraform must be initialized in this directory
#   - OPENAI_API_KEY environment variable must be set
#
###############################################################################

# Check if batch ID is provided
if [ $# -ne 1 ]; then
  echo "Error: Missing batch ID"
  echo "Usage: $0 <batch_id>"
  echo "Example: $0 batch_abc123def456"
  exit 1
fi

BATCH_ID="$1"

# Validate batch ID format
if [[ ! $BATCH_ID =~ ^batch_ ]]; then
  echo "Error: Invalid batch ID format. Expected format: batch_XXXX"
  exit 1
fi

# Check if OPENAI_API_KEY is set
if [ -z "$OPENAI_API_KEY" ]; then
  echo "Error: OPENAI_API_KEY environment variable not set"
  echo "Please set the OPENAI_API_KEY environment variable with your OpenAI API key."
  exit 1
fi

echo "Importing OpenAI Batch: $BATCH_ID"
echo ""

# Create a temporary import configuration
cat > import_batch.tf << EOF
# Temporary configuration for importing batch
resource "openai_batch" "imported" {
  input_file_id    = "file-placeholder"
  endpoint         = "/v1/chat/completions"
  completion_window = "24h"
}
EOF

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Import the batch
echo "Importing batch..."
terraform import openai_batch.imported "$BATCH_ID"

if [ $? -eq 0 ]; then
  echo ""
  echo "Batch imported successfully!"
  echo ""
  echo "Next steps:"
  echo "1. Run 'terraform show' to see the imported batch details"
  echo "2. Update resource.tf with the actual batch configuration"
  echo "3. Run 'terraform plan' to verify no changes are needed"
  echo "4. Remove import_batch.tf when done"
  echo ""
  echo "Note: Batches are immutable once created, so the imported"
  echo "resource will be read-only in Terraform."
else
  echo ""
  echo "Import failed. Please check the batch ID and your permissions."
fi

# Clean up temporary file
rm -f import_batch.tf