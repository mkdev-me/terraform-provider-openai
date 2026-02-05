# Create a group in the organization
resource "openai_group" "engineering" {
  name = "Engineering Team"
}

# Create another group
resource "openai_group" "support" {
  name = "Support Team"
}

# Output the group IDs
output "engineering_group_id" {
  value = openai_group.engineering.id
}

output "support_group_id" {
  value = openai_group.support.id
}

# Check if groups are SCIM managed
output "engineering_is_scim_managed" {
  value = openai_group.engineering.is_scim_managed
}
