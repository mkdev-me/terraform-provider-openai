# OpenAI Organization Users Module
# This module provides a standardized way to retrieve information about users in an OpenAI organization.

# Configuration for single user mode
data "openai_organization_user" "single" {
  count   = var.list_mode ? 0 : 1
  user_id = var.user_id
  api_key = var.api_key
}

# Configuration for list mode
data "openai_organization_users" "list" {
  count   = var.list_mode ? 1 : 0
  after   = var.after
  limit   = var.limit
  emails  = var.emails
  api_key = var.api_key
}

# In list mode, filter users by role
locals {
  list_users = var.list_mode ? data.openai_organization_users.list[0].users : []
  owners     = [for user in local.list_users : user if user.role == "owner"]
  members    = [for user in local.list_users : user if user.role == "member"]
  readers    = [for user in local.list_users : user if user.role == "reader"]
} 