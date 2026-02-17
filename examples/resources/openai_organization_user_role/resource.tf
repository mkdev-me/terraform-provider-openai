# Assign an organization role to a user
resource "openai_organization_user_role" "admin" {
  user_id = var.user_id
  role_id = var.role_id
}
