# OpenAI User Role and Project User Data Sources Example
# --------------------------------------------------


# Create a project for data source demonstration
resource "openai_project" "data_source_example" {
  name = "User Data Sources Example"
}

# Create a project user (needs to reference a real user ID)
resource "openai_project_user" "data_source_user" {
  project_id = openai_project.data_source_example.id
  user_id    = "user-yatSd6LuWvgeoqZbd89xzPlJ" # Real user ID from your organization
  role       = "owner"

  depends_on = [openai_project.data_source_example]
}

# 1. Retrieve organization user information using the data source
data "openai_organization_user" "user_info" {
  user_id = "user-yatSd6LuWvgeoqZbd89xzPlJ" # Real user ID from your organization
  api_key = var.openai_admin_key            # Pass the admin key with proper permissions
}

# 2. Retrieve project user information using the data source
data "openai_project_user" "project_user_info" {
  project_id = openai_project.data_source_example.id
  user_id    = openai_project_user.data_source_user.user_id

  # Data source will only be evaluated after the user is added to the project
  depends_on = [openai_project_user.data_source_user]
}

# 3. Retrieve all project users using the data source
data "openai_project_users" "all_project_users" {
  project_id = openai_project.data_source_example.id

  # Data source will only be evaluated after the user is added to the project
  depends_on = [openai_project_user.data_source_user]
}

# 4. Retrieve all organization users using the data source
data "openai_organization_users" "all_org_users" {
  limit   = 20
  api_key = var.openai_admin_key # Pass the admin key with proper permissions
}

# 5. Use the project_user module in list mode
module "list_project_users" {
  source     = "../../modules/project_user"
  project_id = openai_project.data_source_example.id
  list_mode  = true # This enables list mode to fetch all users in the project

  # Data source will only be evaluated after the user is added to the project
  depends_on = [openai_project_user.data_source_user]
}

# Outputs - Organization User Data Source
output "data_source_user_role" {
  value = data.openai_organization_user.user_info.role
}

output "data_source_user_email" {
  value = data.openai_organization_user.user_info.email
}

output "data_source_user_name" {
  value = data.openai_organization_user.user_info.name
}

# Outputs - Project User Data Source
output "data_source_project_user_role" {
  value = data.openai_project_user.project_user_info.role
}

output "data_source_project_user_email" {
  value = data.openai_project_user.project_user_info.email
}

output "data_source_project_user_added_at" {
  value = data.openai_project_user.project_user_info.added_at
}

# Outputs - Project Users Data Source
output "data_source_project_users_count" {
  value = data.openai_project_users.all_project_users.user_count
}

output "data_source_project_users_list" {
  value = data.openai_project_users.all_project_users.user_ids
}

output "data_source_project_owners" {
  value = data.openai_project_users.all_project_users.owner_ids
}

output "data_source_project_members" {
  value = data.openai_project_users.all_project_users.member_ids
}

# Outputs - Organization Users Data Source
output "data_source_org_users_count" {
  value = length(data.openai_organization_users.all_org_users.users)
}

output "data_source_org_owners_count" {
  value = length([
    for user in data.openai_organization_users.all_org_users.users :
    user if user.role == "owner"
  ])
}

# Outputs - Module in List Mode
output "module_list_mode_user_count" {
  value = module.list_project_users.user_count
}

output "module_list_mode_owners" {
  value = module.list_project_users.owner_ids
}

output "module_list_mode_all_users" {
  value = module.list_project_users.all_user_ids
}
