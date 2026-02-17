# Create a custom project-level role
resource "openai_project_role" "deployer" {
  project_id  = var.project_id
  name        = "Model Deployer"
  permissions = ["api.project.models.write", "api.project.models.read"]
  description = "Can deploy and read models"
}

output "role_id" {
  value = openai_project_role.deployer.id
}
