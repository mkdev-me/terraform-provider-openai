#!/bin/bash
###############################################################################
# OpenAI Project API Key Import Script
###############################################################################
# This script demonstrates how to properly reference an existing OpenAI Project 
# API key in Terraform.
#
# Usage:
#   ./import_project_key.sh <project_id> <api_key_name_or_id>
#
# Arguments:
#   project_id         - ID of the project (format: proj_XXXX)
#   api_key_name_or_id - Name or ID of the API key
#
# Example:
#   ./import_project_key.sh proj_abc123 terraform
#
# Requirements:
#   - Terraform must be initialized in this directory
#   - The Project API Key must already exist in the OpenAI dashboard
#   - OPENAI_ADMIN_KEY environment variable must be set
#
###############################################################################

# Define colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if arguments are provided
if [ $# -ne 2 ]; then
  echo -e "${RED}ERROR: Missing required arguments${NC}"
  echo "Usage: $0 <project_id> <api_key_name_or_id>"
  echo "Example: $0 proj_abc123 terraform"
  exit 1
fi

# Check if OPENAI_ADMIN_KEY is set
if [ -z "$OPENAI_ADMIN_KEY" ]; then
  echo -e "${RED}ERROR: OPENAI_ADMIN_KEY environment variable not set${NC}"
  echo "Please set the OPENAI_ADMIN_KEY environment variable with your OpenAI admin API key."
  exit 1
fi

PROJECT_ID="$1"
API_KEY_NAME_OR_ID="$2"

# Validate the format of the project ID
if [[ ! $PROJECT_ID =~ ^proj_ ]]; then
  echo -e "${RED}ERROR: Invalid project ID format. Expected format: proj_XXXX${NC}"
  exit 1
fi

# Check if API_KEY_NAME_OR_ID is an ID (starts with key_) or a name
if [[ $API_KEY_NAME_OR_ID =~ ^key_ ]]; then
  # It's already an ID
  API_KEY_ID="$API_KEY_NAME_OR_ID"
  echo -e "${YELLOW}Using API Key ID: $API_KEY_ID${NC}"

  # Try to fetch the key details to get the name
  echo -e "${YELLOW}Fetching key details...${NC}"
  KEY_DETAILS=$(curl -s "https://api.openai.com/v1/organization/projects/$PROJECT_ID/api_keys/$API_KEY_ID" \
    -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
    -H "Content-Type: application/json")
  
  if echo "$KEY_DETAILS" | grep -q "error"; then
    echo -e "${RED}Error fetching key details:${NC}"
    echo "$KEY_DETAILS" | jq .
    exit 1
  fi
  
  API_KEY_NAME=$(echo "$KEY_DETAILS" | jq -r '.name')
  echo -e "${GREEN}Found key name: $API_KEY_NAME${NC}"
else
  # It's a name, need to find the ID
  API_KEY_NAME="$API_KEY_NAME_OR_ID"
  echo -e "${YELLOW}Looking up key ID for name: $API_KEY_NAME${NC}"
  
  # Fetch all keys and find the one with matching name
  ALL_KEYS=$(curl -s "https://api.openai.com/v1/organization/projects/$PROJECT_ID/api_keys" \
    -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
    -H "Content-Type: application/json")
  
  if echo "$ALL_KEYS" | grep -q "error"; then
    echo -e "${RED}Error fetching keys:${NC}"
    echo "$ALL_KEYS" | jq .
    exit 1
  fi
  
  API_KEY_ID=$(echo "$ALL_KEYS" | jq -r --arg NAME "$API_KEY_NAME" '.data[] | select(.name == $NAME) | .id')
  
  if [ -z "$API_KEY_ID" ]; then
    echo -e "${RED}ERROR: No API key found with name '$API_KEY_NAME'${NC}"
    echo "Available keys:"
    echo "$ALL_KEYS" | jq -r '.data[] | "- \(.name) (ID: \(.id))"'
    exit 1
  fi
  
  echo -e "${GREEN}Found key ID: $API_KEY_ID${NC}"
fi

echo -e "${GREEN}===== OpenAI Project API Key Setup =====${NC}"
echo ""
echo -e "${YELLOW}IMPORTANT: The OpenAI API doesn't allow creating project API keys programmatically.${NC}"
echo -e "${YELLOW}This script will create a representation of your existing API key in Terraform.${NC}"
echo ""
echo "Project ID: $PROJECT_ID"
echo "API Key Name: $API_KEY_NAME"
echo "API Key ID: $API_KEY_ID"
echo ""

# Initialize Terraform with upgrade to handle new dependencies
echo "Initializing Terraform with upgrade..."
terraform init -upgrade
echo ""

# Update main.tf with correct values before applying
echo "Updating main.tf with key information..."
# Create temporary file with updated variables
cat > main.tf.tmp <<EOF
# Example: OpenAI Project API Keys
# This example demonstrates how to use and track project API keys in Terraform

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
  }
}

# Configure the OpenAI Provider
provider "openai" {
  # Admin API key will be sourced from environment variable:
  # OPENAI_ADMIN_KEY
  admin_key = var.openai_admin_key
}

# Admin API key for authentication
variable "openai_admin_key" {
  type        = string
  description = "OpenAI Admin API key (requires administrative permissions)"
  sensitive   = true
}

# Define your project ID here
variable "project_id" {
  description = "The ID of the OpenAI project"
  type        = string
  default     = "${PROJECT_ID}"
}

variable "name" {
  description = "The name of the API key in the OpenAI dashboard"
  type        = string
  default     = "${API_KEY_NAME}"
}

variable "api_key_id" {
  description = "The ID of the API key (format: key_XXXX)"
  type        = string
  default     = "${API_KEY_ID}"
}

# IMPORTANT: This module is for TRACKING project API keys in Terraform.
# Project API keys MUST be manually created in the OpenAI dashboard at:
# https://platform.openai.com/api-keys
#
# The workflow is:
# 1. Create an API key manually in the OpenAI dashboard
# 2. Either:
#    a. Use this module directly with the correct ID
#    b. Use the import_project_key.sh script to import existing keys
#
# This creates a resource representation of your API key in Terraform.
module "test_key" {
  source = "../../modules/project_api"

  project_id = var.project_id
  name       = var.name
  api_key_id = var.api_key_id
}

# Outputs
output "project_key" {
  description = "Project API key details"
  value = {
    id         = module.test_key.api_key_id
    name       = module.test_key.name
    project_id = module.test_key.project_id
    created_at = module.test_key.created_at
  }
}
EOF

# Replace main.tf with the updated version
mv main.tf.tmp main.tf
echo -e "${GREEN}Done!${NC}"
echo ""

# Run terraform apply to create the resource
echo "Running terraform apply to create the resource representation..."
echo -e "${YELLOW}This will create the key representation in your state file.${NC}"
terraform apply -auto-approve
echo ""

echo -e "${GREEN}===== Setup Complete =====${NC}"
echo ""
echo "Your Terraform state now contains a reference to your OpenAI project API key."
echo ""
echo -e "You can verify the key details with:"
echo -e "${YELLOW}terraform state show module.test_key.null_resource.project_api_key${NC}"
echo ""
echo -e "${RED}IMPORTANT: The API key value itself is not accessible via the API and"
echo -e "will not be available in your Terraform state. Only the metadata"
echo -e "(ID, name, etc.) about the key is tracked in Terraform.${NC}" 