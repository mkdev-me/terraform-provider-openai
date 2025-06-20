# Single user mode outputs
output "user" {
  description = "Complete user object"
  value       = var.list_mode ? null : data.openai_organization_user.single[0]
}

output "user_id" {
  description = "The ID of the user"
  value       = var.list_mode ? null : data.openai_organization_user.single[0].id
}

output "email" {
  description = "The email address of the user"
  value       = var.list_mode ? null : data.openai_organization_user.single[0].email
}

output "name" {
  description = "The name of the user"
  value       = var.list_mode ? null : data.openai_organization_user.single[0].name
}

output "role" {
  description = "The role of the user in the organization"
  value       = var.list_mode ? null : data.openai_organization_user.single[0].role
}

output "added_at" {
  description = "The Unix timestamp when the user was added"
  value       = var.list_mode ? null : data.openai_organization_user.single[0].added_at
}

# List mode outputs
output "all_users" {
  description = "List of all users with complete details"
  value       = var.list_mode ? local.list_users : []
}

output "user_count" {
  description = "Total number of users in the result"
  value       = var.list_mode ? length(local.list_users) : 0
}

output "owners" {
  description = "List of users with the 'owner' role"
  value       = var.list_mode ? local.owners : []
}

output "members" {
  description = "List of users with the 'member' role"
  value       = var.list_mode ? local.members : []
}

output "readers" {
  description = "List of users with the 'reader' role"
  value       = var.list_mode ? local.readers : []
}

output "first_id" {
  description = "ID of the first user in the result"
  value       = var.list_mode ? data.openai_organization_users.list[0].first_id : null
}

output "last_id" {
  description = "ID of the last user in the result"
  value       = var.list_mode ? data.openai_organization_users.list[0].last_id : null
}

output "has_more" {
  description = "Whether there are more users available"
  value       = var.list_mode ? data.openai_organization_users.list[0].has_more : false
} 