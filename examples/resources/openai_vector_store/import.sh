#!/bin/bash
###############################################################################
# OpenAI Vector Store Import Script
###############################################################################
# This script demonstrates how to import an existing OpenAI vector store into
# Terraform state.
#
# Usage:
#   ./import.sh <vector_store_id>
#
# Arguments:
#   vector_store_id - ID of the vector store to import (format: vs_XXXX)
#
# Example:
#   ./import.sh vs_abc123def456
#
# Requirements:
#   - Terraform must be initialized in this directory
#   - OPENAI_API_KEY environment variable must be set
#
###############################################################################

# Check if vector store ID is provided
if [ $# -ne 1 ]; then
  echo "Error: Missing vector store ID"
  echo "Usage: $0 <vector_store_id>"
  echo "Example: $0 vs_abc123def456"
  exit 1
fi

VECTOR_STORE_ID="$1"

# Validate vector store ID format
if [[ ! $VECTOR_STORE_ID =~ ^vs_ ]]; then
  echo "Error: Invalid vector store ID format. Expected format: vs_XXXX"
  exit 1
fi

# Check if OPENAI_API_KEY is set
if [ -z "$OPENAI_API_KEY" ]; then
  echo "Error: OPENAI_API_KEY environment variable not set"
  echo "Please set the OPENAI_API_KEY environment variable with your OpenAI API key."
  exit 1
fi

echo "Importing OpenAI Vector Store: $VECTOR_STORE_ID"
echo ""

# Create a temporary import configuration
cat > import_vector_store.tf << EOF
# Temporary configuration for importing vector store
resource "openai_vector_store" "imported" {
  name = "Imported Vector Store"
}
EOF

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Import the vector store
echo "Importing vector store..."
terraform import openai_vector_store.imported "$VECTOR_STORE_ID"

if [ $? -eq 0 ]; then
  echo ""
  echo "Vector store imported successfully!"
  echo ""
  echo "Next steps:"
  echo "1. Run 'terraform show' to see the imported vector store details"
  echo "2. Update resource.tf with the actual vector store configuration"
  echo "3. Run 'terraform plan' to verify no changes are needed"
  echo "4. Remove import_vector_store.tf when done"
else
  echo ""
  echo "Import failed. Please check the vector store ID and your permissions."
fi

# Clean up temporary file
rm -f import_vector_store.tf