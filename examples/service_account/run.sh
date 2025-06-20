#!/bin/bash
# Check if admin key is provided
if [ -z "$1" ]; then
  echo "Usage: ./run.sh YOUR_OPENAI_ADMIN_KEY"
  echo "Example: ./run.sh sk-your-openai-admin-key"
  exit 1
fi

# Store the admin key in a variable
OPENAI_ADMIN_KEY="$1"

# Run terraform apply with the admin key
terraform apply \
  -var="try_create_service_account=true" \
  -var="try_data_sources=false" \
  -var="openai_admin_key=$OPENAI_ADMIN_KEY" \
  -auto-approve 