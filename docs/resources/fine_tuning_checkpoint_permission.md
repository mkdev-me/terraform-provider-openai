---
page_title: "openai_fine_tuning_checkpoint_permission Resource - terraform-provider-openai"
subcategory: ""
description: |-
  Manages permissions for fine-tuning checkpoints, allowing you to share checkpoints with other organizations.
---

# Resource: openai_fine_tuning_checkpoint_permission

This resource allows you to manage permissions for fine-tuning checkpoints in your OpenAI account. You can use this to share checkpoints with other organizations, allowing them to use your checkpoints for their own fine-tuning jobs.

## Example Usage

```terraform
# First, get checkpoints from an existing fine-tuning job
data "openai_fine_tuning_checkpoints" "example" {
  fine_tuning_job_id = "ftjob-abc123"
}

# Then, share a checkpoint with another organization
resource "openai_fine_tuning_checkpoint_permission" "share" {
  checkpoint_id = data.openai_fine_tuning_checkpoints.example.checkpoints[0].checkpoint_id
  organization_id = "org-XYZ789"
}
```

## Argument Reference

* `checkpoint_id` - (Required) The ID of the checkpoint to share.
* `organization_id` - (Required) The ID of the organization to grant access to.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A unique identifier for this permission, in the format of "checkpoint_id:organization_id".

## Import

Checkpoint permissions can be imported using the format `checkpoint_id:organization_id`, e.g.,

```
$ terraform import openai_fine_tuning_checkpoint_permission.example ckpt-abc123:org-xyz789
```

## Timeouts

This resource provides the following timeouts configuration options:

- `create` - Default is 5 minutes.
- `read` - Default is 5 minutes.
- `delete` - Default is 5 minutes.

## Admin API Key Support

This resource requires an admin API key with specific scopes for operations. The provider now automatically reads the admin API key from the `OPENAI_ADMIN_KEY` environment variable, eliminating the need to pass it explicitly in your Terraform configuration.

```bash
# Set the admin API key as an environment variable
export OPENAI_ADMIN_KEY="your-admin-api-key"

# Run Terraform commands
terraform apply
```

This approach enhances security by avoiding placing sensitive keys in your Terraform code or state files.

## Required Permissions

This operation requires admin privileges. The API key must have:
- Owner role in the organization
- The `api.fine_tuning.checkpoints.write` scope

## Notes

If you receive a 401 Unauthorized error with a message about missing scopes, ensure that:
1. You are using an admin API key with the appropriate scope
2. The key is properly configured in the `OPENAI_ADMIN_KEY` environment variable
3. The user associated with the API key has sufficient permissions in the OpenAI organization 