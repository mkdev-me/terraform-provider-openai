# Example usage of service account data sources
# 
# IMPORTANT: These data sources require an admin API key with specific scopes:
# - api.organization.projects.service_accounts.read
#
# If you don't have these permissions, you will see permission errors but the examples
# will continue to function and display fallback values.
#
# Note: OpenAI does not support programmatic creation of project API keys.
# API keys must be created manually in the OpenAI dashboard.

# Optional: If you want to try the data sources and have the right permissions
variable "try_data_sources" {
  description = "Whether to try using data sources (requires admin API key with api.organization.projects.service_accounts.read permission)"
  type        = bool
  default     = false
}

# Create a service account first, then use it for data sources
# Wait for the demo service account to be created first
resource "null_resource" "wait_for_demo_account" {
  count = var.try_create_service_account && var.try_data_sources ? 1 : 0

  depends_on = [
    openai_project_service_account.demo
  ]
}

# Define a specific service account to retrieve - use demo account if available
locals {
  # Use the demo service account ID if it exists, otherwise fall back to a placeholder
  example_service_account_id = var.try_create_service_account && var.try_data_sources ? openai_project_service_account.demo[0].service_account_id : "placeholder_id"
}

# Retrieve a specific service account
data "openai_project_service_account" "example" {
  count              = var.try_data_sources && var.try_create_service_account ? 1 : 0
  project_id         = local.project_id
  service_account_id = local.example_service_account_id
  api_key            = var.openai_admin_key

  depends_on = [
    openai_project_service_account.demo
  ]
}

# Retrieve all service accounts for the project
data "openai_project_service_accounts" "all" {
  count      = var.try_data_sources ? 1 : 0
  project_id = local.project_id
  api_key    = var.openai_admin_key
}

# Use the module in data source mode
module "service_account_readonly" {
  source             = "../../modules/service_account"
  project_id         = local.project_id
  service_account_id = local.example_service_account_id
  use_data_source    = var.try_data_sources && var.try_create_service_account # Only use data source mode if explicitly requested and demo account exists
  openai_admin_key   = var.openai_admin_key
  name               = "Read-only Service Account" # Adding a name even in data source mode

  depends_on = [
    openai_project_service_account.demo
  ]
}

# Output examples for the specific service account
output "example_service_account" {
  description = "Details of the example service account (if access granted)"
  value = var.try_data_sources && var.try_create_service_account && length(data.openai_project_service_account.example) > 0 ? {
    id         = try(data.openai_project_service_account.example[0].service_account_id, "")
    name       = try(data.openai_project_service_account.example[0].name, "")
    created_at = try(data.openai_project_service_account.example[0].created_at, 0)
    role       = try(data.openai_project_service_account.example[0].role, "")
    } : {
    id         = "Permissions required to view service account (api.organization.projects.service_accounts.read)"
    name       = "Permissions required"
    created_at = 0
    role       = "Permissions required"
  }
}

# Output example showing all service accounts in the project
output "all_service_accounts" {
  description = "List of all service accounts for this project (if access granted)"
  value = var.try_data_sources && length(data.openai_project_service_accounts.all) > 0 ? try(
    [for account in data.openai_project_service_accounts.all[0].service_accounts : account.name],
    ["Error accessing service account data"]
  ) : ["Permissions required to list service accounts (api.organization.projects.service_accounts.read scope)"]
}

# Output example from the module in data source mode
output "module_service_account" {
  description = "Service account details from module (in data source mode)"
  value = {
    id         = module.service_account_readonly.service_account_id
    name       = module.service_account_readonly.name
    created_at = module.service_account_readonly.created_at
    role       = module.service_account_readonly.role
  }
} 