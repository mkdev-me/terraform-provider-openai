#!/bin/bash
###############################################################################
# OpenAI Fine-Tuning Job Import Script
###############################################################################
# This script demonstrates how to import an existing OpenAI fine-tuning job into
# Terraform state.
#
# Usage:
#   ./import.sh <fine_tuning_job_id>
#
# Arguments:
#   fine_tuning_job_id - ID of the fine-tuning job to import (format: ftjob-XXXX)
#
# Example:
#   ./import.sh ftjob-abc123def456
#
# Requirements:
#   - Terraform must be initialized in this directory
#   - OPENAI_API_KEY environment variable must be set
#
###############################################################################

# Check if fine-tuning job ID is provided
if [ $# -ne 1 ]; then
  echo "Error: Missing fine-tuning job ID"
  echo "Usage: $0 <fine_tuning_job_id>"
  echo "Example: $0 ftjob-abc123def456"
  exit 1
fi

FINE_TUNING_JOB_ID="$1"

# Validate fine-tuning job ID format
if [[ ! $FINE_TUNING_JOB_ID =~ ^ftjob- ]]; then
  echo "Error: Invalid fine-tuning job ID format. Expected format: ftjob-XXXX"
  exit 1
fi

# Check if OPENAI_API_KEY is set
if [ -z "$OPENAI_API_KEY" ]; then
  echo "Error: OPENAI_API_KEY environment variable not set"
  echo "Please set the OPENAI_API_KEY environment variable with your OpenAI API key."
  exit 1
fi

echo "Importing OpenAI Fine-Tuning Job: $FINE_TUNING_JOB_ID"
echo ""

# Create a temporary import configuration
cat > import_fine_tuning_job.tf << EOF
# Temporary configuration for importing fine-tuning job
resource "openai_fine_tuning_job" "imported" {
  training_file = "file-placeholder"
  model         = "gpt-3.5-turbo"
}
EOF

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Import the fine-tuning job
echo "Importing fine-tuning job..."
terraform import openai_fine_tuning_job.imported "$FINE_TUNING_JOB_ID"

if [ $? -eq 0 ]; then
  echo ""
  echo "Fine-tuning job imported successfully!"
  echo ""
  echo "Next steps:"
  echo "1. Run 'terraform show' to see the imported fine-tuning job details"
  echo "2. Update resource.tf with the actual fine-tuning job configuration"
  echo "3. Run 'terraform plan' to verify no changes are needed"
  echo "4. Remove import_fine_tuning_job.tf when done"
  echo ""
  echo "Note: Fine-tuning jobs are immutable once created, so the imported"
  echo "resource will be read-only in Terraform."
else
  echo ""
  echo "Import failed. Please check the fine-tuning job ID and your permissions."
fi

# Clean up temporary file
rm -f import_fine_tuning_job.tf