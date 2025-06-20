# OpenAI Invites Data Source

The `openai_invites` data source allows you to retrieve a list of all pending invitations in your OpenAI organization.

> **Warning:** Organizations with many pending invitations may experience timeouts when using this data source. The OpenAI API can take a long time to return a large number of invitations, potentially exceeding the default timeout. See the "Performance Considerations" section below for recommendations.

## Example Usage

```hcl
# Retrieve all pending invitations
data "openai_invites" "all" {}

output "invitation_count" {
  value = length(data.openai_invites.all.invites)
}

output "invitation_emails" {
  value = [for invite in data.openai_invites.all.invites : invite.email]
}
```

### With Custom Authentication

```hcl
data "openai_invites" "all" {
  api_key = var.openai_admin_key  # Use a specific API key for authentication
}
```

### Conditional Usage (Recommended for Large Organizations)

```hcl
# Variable to control whether to list all invitations
variable "list_invites" {
  description = "Whether to list all invitations (can cause timeouts in organizations with many invites)"
  type        = bool
  default     = false
}

# Only attempt to list invitations if explicitly enabled
data "openai_invites" "all" {
  count = var.list_invites ? 1 : 0
}

# Safely reference the invites with fallback for when data source isn't used
locals {
  invites = var.list_invites ? data.openai_invites.all[0].invites : []
}

output "invitation_count" {
  value = length(local.invites)
}
```

## Mapping and Filtering Example

```hcl
data "openai_invites" "all" {}

locals {
  # Create a map of invite IDs to their details
  invites_by_id = {
    for invite in data.openai_invites.all.invites :
    invite.id => invite
  }
  
  # Filter to only organization owners
  owner_invites = [
    for invite in data.openai_invites.all.invites :
    invite if invite.role == "owner"
  ]
}

output "owner_invite_count" {
  value = length(local.owner_invites)
}
```

## Combining With Invitation Creation

```hcl
# Create a new invitation
resource "openai_invite" "example" {
  email = "someone@example.com"
  role  = "reader"
}

# List all invitations including the one we just created
data "openai_invites" "all" {
  depends_on = [openai_invite.example]
}

output "total_pending_invites" {
  value = length(data.openai_invites.all.invites)
}
```

## Argument Reference

* `api_key` - (Optional) API key for authentication. If not provided, the provider's default API key will be used.

## Attribute Reference

* `invites` - A list of pending invitations with the following attributes:
  * `id` - The ID of the invitation.
  * `email` - The email address of the invited user.
  * `role` - The role assigned to the invited user (owner or reader).
  * `status` - The status of the invitation.
  * `created_at` - A timestamp of when the invitation was created, formatted as an RFC3339 string.
  * `expires_at` - A timestamp of when the invitation expires, formatted as an RFC3339 string.

## Related Resources

* [`openai_invite` Resource](../resources/invite.md)
* [`openai_invite` Data Source](../data-sources/invite.md)

## Notes

* Administrative permissions are required to list invitations.
* This data source only returns pending invitations, not accepted or expired ones.
* When combined with invitation creation, use `depends_on` to ensure the data source reads the latest state.
* If a timeout occurs, the provider will return an empty list and a warning diagnostic message.

## Performance Considerations

Organizations with a large number of pending invitations may experience timeout issues when using this data source, as the OpenAI API can be slow to respond when retrieving many invitations. To address this:

1. **Use Conditional Data Sources**: As shown in the "Conditional Usage" example, consider making the data source optional using a variable.
2. **Filter Invitations Locally**: Instead of relying on API-side filtering, retrieve all invitations and filter them in Terraform using local expressions.
3. **Handle Timeout Gracefully**: The provider will return an empty list and a warning if a timeout occurs, allowing your configuration to continue running without failing.
4. **Consider Alternative Approaches**: For organizations with many invitations, you might need to use the OpenAI API directly with higher timeouts or pagination. 