# First, create a project
resource "openai_project" "production" {
  name = "Production API"
}

# Look up project-level roles
data "openai_project_role" "member" {
  project_id = openai_project.production.id
  name       = "member"
}

# Add a user to the project with one or more roles
resource "openai_project_user" "developer" {
  project_id = openai_project.production.id
  user_id    = var.user_id
  role_ids   = [data.openai_project_role.member.id]
}

output "user_email" {
  value = openai_project_user.developer.email
}
