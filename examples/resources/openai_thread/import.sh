#!/bin/bash
###############################################################################
# OpenAI Thread Import Script
###############################################################################
# This script demonstrates how to import an existing OpenAI thread into
# Terraform state.
#
# Usage:
#   ./import.sh <thread_id>
#
# Arguments:
#   thread_id - ID of the thread to import (format: thread_XXXX)
#
# Example:
#   ./import.sh thread_abc123def456
#
# Requirements:
#   - Terraform must be initialized in this directory
#   - OPENAI_API_KEY environment variable must be set
#
###############################################################################

# Check if thread ID is provided
if [ $# -ne 1 ]; then
  echo "Error: Missing thread ID"
  echo "Usage: $0 <thread_id>"
  echo "Example: $0 thread_abc123def456"
  exit 1
fi

THREAD_ID="$1"

# Validate thread ID format
if [[ ! $THREAD_ID =~ ^thread_ ]]; then
  echo "Error: Invalid thread ID format. Expected format: thread_XXXX"
  exit 1
fi

# Check if OPENAI_API_KEY is set
if [ -z "$OPENAI_API_KEY" ]; then
  echo "Error: OPENAI_API_KEY environment variable not set"
  echo "Please set the OPENAI_API_KEY environment variable with your OpenAI API key."
  exit 1
fi

echo "Importing OpenAI Thread: $THREAD_ID"
echo ""

# Create a temporary import configuration
cat > import_thread.tf << EOF
# Temporary configuration for importing thread
resource "openai_thread" "imported" {
  # Minimal configuration for import
  metadata = {}
}
EOF

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Import the thread
echo "Importing thread..."
terraform import openai_thread.imported "$THREAD_ID"

if [ $? -eq 0 ]; then
  echo ""
  echo "Thread imported successfully!"
  echo ""
  echo "Next steps:"
  echo "1. Run 'terraform show' to see the imported thread details"
  echo "2. Update resource.tf with the actual thread configuration"
  echo "3. Run 'terraform plan' to verify no changes are needed"
  echo "4. Remove import_thread.tf when done"
else
  echo ""
  echo "Import failed. Please check the thread ID and your permissions."
fi

# Clean up temporary file
rm -f import_thread.tf