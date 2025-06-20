# Fine-Tuning Checkpoint Permissions Data Source

This data source provides information about permissions associated with a fine-tuning checkpoint. It allows you to retrieve the list of projects that have access to a specific fine-tuned model checkpoint.

## Example Usage

```hcl
data "openai_fine_tuning_checkpoint_permissions" "example" {
  # IMPORTANT: Use the fine-tuned model ID format, not the checkpoint ID format
  checkpoint_id = "ft:gpt-4o-mini-2024-07-18:fodoj-gmbh::BGvDTdTK"
  
  # Optional parameters
  limit = 20
  # after = "permission-xyz789"  # For pagination
}

output "checkpoint_permissions" {
  value = data.openai_fine_tuning_checkpoint_permissions.example.permissions
}
```

## Admin API Key Support

This data source supports using an admin API key for operations. The admin API key is automatically read from the `OPENAI_ADMIN_KEY` environment variable. You don't need to explicitly provide it in the data source configuration.

```bash
# Set the admin API key as an environment variable
export OPENAI_ADMIN_KEY="your-admin-api-key"

# Run Terraform commands
terraform apply
```

## Argument Reference

* `checkpoint_id` - (Required) The ID of the checkpoint to fetch permissions for. Note that this should be in the fine-tuned model format (e.g., `ft:gpt-4o-mini-2024-07-18:org:model:id`) rather than the checkpoint ID format (e.g., `ftckpt_abc123`).
* `limit` - (Optional) The maximum number of permissions to retrieve. Defaults to 20.
* `after` - (Optional) Identifier for the last permission from the previous pagination request, used for retrieving the next page of results.

## Attributes Reference

* `permissions` - A list of permissions. Each permission contains:
  * `id` - The unique identifier for this permission.
  * `object` - The type of object, always "checkpoint.permission".
  * `created_at` - The Unix timestamp when the permission was created.
  * `checkpoint_id` - The ID of the checkpoint.
  * `project_ids` - A list of project IDs that have access to this checkpoint.
  * `allow_view` - Whether viewing the checkpoint is allowed.
  * `allow_create` - Whether creating from the checkpoint is allowed.
* `has_more` - Whether there are more permissions available that weren't included in this response due to the `limit` parameter.

## Required Permissions

This operation requires admin privileges. The API key must have:
- Owner role in the organization
- The `api.fine_tuning.checkpoints.read` scope

## Notes

If you receive a 401 Unauthorized error with a message about missing scopes, ensure that:
1. You are using an admin API key with the appropriate scope
2. The key is properly configured in the `OPENAI_ADMIN_KEY` environment variable
3. The user associated with the API key has sufficient permissions in the OpenAI organization 