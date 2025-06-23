#!/bin/bash
###############################################################################
# OpenAI Project Import Script
###############################################################################
# This script demonstrates how to import an existing OpenAI project into
# Terraform state.
#
# Usage:
#   ./import.sh <project_id>
#
# Arguments:
#   project_id - ID of the project to import (format: proj_XXXX)
#
# Example:
#   ./import.sh proj_abc123def456
#
# Requirements:
#   - Terraform must be initialized in this directory
#   - OPENAI_ADMIN_KEY environment variable must be set
#
###############################################################################

# Check if project ID is provided
if [ $# -ne 1 ]; then
  echo "Error: Missing project ID"
  echo "Usage: $0 <project_id>"
  echo "Example: $0 proj_abc123def456"
  exit 1
fi

PROJECT_ID="$1"

# Validate project ID format
if [[ ! $PROJECT_ID =~ ^proj_ ]]; then
  echo "Error: Invalid project ID format. Expected format: proj_XXXX"
  exit 1
fi

# Check if OPENAI_ADMIN_KEY is set
if [ -z "$OPENAI_ADMIN_KEY" ]; then
  echo "Error: OPENAI_ADMIN_KEY environment variable not set"
  echo "Please set the OPENAI_ADMIN_KEY environment variable with your OpenAI admin API key."
  exit 1
fi

echo "Importing OpenAI Project: $PROJECT_ID"
echo ""

# Create a temporary import configuration
cat > import_project.tf << EOF
# Temporary configuration for importing project
resource "openai_project" "imported" {
  name = "Imported Project"
}
EOF

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Import the project
echo "Importing project..."
terraform import openai_project.imported "$PROJECT_ID"

if [ $? -eq 0 ]; then
  echo ""
  echo "Project imported successfully!"
  echo ""
  echo "Next steps:"
  echo "1. Run 'terraform show' to see the imported project details"
  echo "2. Update resource.tf with the actual project configuration"
  echo "3. Run 'terraform plan' to verify no changes are needed"
  echo "4. Remove import_project.tf when done"
else
  echo ""
  echo "Import failed. Please check the project ID and your permissions."
fi

# Clean up temporary file
rm -f import_project.tf