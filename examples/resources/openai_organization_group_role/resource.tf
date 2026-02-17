# Assign an organization role to a group
resource "openai_organization_group_role" "eng_billing" {
  group_id = var.group_id
  role_id  = var.role_id
}
