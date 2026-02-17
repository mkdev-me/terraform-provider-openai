# First, create a project
resource "openai_project" "production" {
  name = "Production API"
}

# Look up project-level roles
data "openai_project_role" "member" {
  project_id = openai_project.production.id
  name       = "member"
}

# Add a group to the project with one or more roles
resource "openai_project_group" "engineering" {
  project_id = openai_project.production.id
  group_id   = var.group_id
  role_ids   = [data.openai_project_role.member.id]
}

output "group_name" {
  value = openai_project_group.engineering.group_name
}
