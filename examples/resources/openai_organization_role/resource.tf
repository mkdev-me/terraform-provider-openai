# Create a custom organization role
resource "openai_organization_role" "billing_viewer" {
  name        = "Billing Viewer"
  permissions = ["api.organization.billing.read"]
  description = "Can view billing information"
}

output "role_id" {
  value = openai_organization_role.billing_viewer.id
}
