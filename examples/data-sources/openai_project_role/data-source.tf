# Look up a project role by name
data "openai_project_role" "deployer" {
  project_id = var.project_id
  name       = "Model Deployer"
}

output "role_id" {
  value = data.openai_project_role.deployer.role_id
}
