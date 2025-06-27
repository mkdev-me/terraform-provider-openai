# Example: Managing an existing organization user's role
# Note: Users cannot be created through this resource - they must already exist in the organization
data "openai_organization_users" "organization_users" {
}

locals {
  owners_by_email = {
    for user in data.openai_organization_users.organization_users.users :
    user.email => user
  }
}

resource "openai_organization_user" "organization_users" {
  user_id = local.owners_by_email["user@mkdev.me"].id
  role    = "owner"
}

