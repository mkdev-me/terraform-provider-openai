# Fine-Tuning Checkpoint Permissions

This document explains how to use the `openai_fine_tuning_checkpoint_permission` resource to manage fine-tuning checkpoint permissions in your OpenAI organization.

## Prerequisites

Before using this resource, ensure you have:

1. **Admin API Key**: You need an API key with admin privileges, which includes:
   - The `api.fine_tuning.checkpoints.write` scope
   - Owner role in your organization

2. **Valid Checkpoint ID**: A fine-tuning checkpoint ID to share
   - Format example: `ft:gpt-4o-mini-2024-07-18:org-xyz::ABcDeFgH`
   - You can get this from a completed fine-tuning job

3. **Valid Project ID**: The project(s) you want to grant access to
   - Format example: `proj_AbCdEfGhIj123456`
   - This MUST be a different project than the one that owns the checkpoint

## Important Limitations

1. **Admin API Key Required**: Both creation and deletion of checkpoint permissions require an admin API key
2. **Target Project Must Be Different**: You cannot grant permission to the project that already owns the checkpoint
3. **Must Use Target Flag for Deletion**: When destroying a specific permission, you must use the `-target` flag

## Creating Checkpoint Permissions

### 1. Set the Required Variables

```bash
export OPENAI_API_KEY="your-regular-api-key"
export OPENAI_ADMIN_KEY="your-admin-api-key-with-proper-scope"
```

### 2. Run Terraform with Admin Key

```bash
terraform apply \
  -var="admin_api_key=$OPENAI_ADMIN_KEY" \
  -var="checkpoint_id=ft:gpt-4o-mini-2024-07-18:org-xyz::ABcDeFgH" \
  -var="project_id=proj_AbCdEfGhIj123456" \
  -var="use_admin_key=true"
```

## Deleting Checkpoint Permissions

Checkpoint permissions must be deleted using the admin API key and with the `-target` flag:

```bash
terraform destroy \
  -target='openai_fine_tuning_checkpoint_permission.checkpoint_permission[0]' \
  -var="admin_api_key=$OPENAI_ADMIN_KEY" \
  -var="checkpoint_id=ft:gpt-4o-mini-2024-07-18:org-xyz::ABcDeFgH" \
  -var="project_id=proj_AbCdEfGhIj123456" \
  -var="use_admin_key=true"
```

**Note**: You MUST specify the `-target` flag with the exact resource name as shown above. Attempting to destroy all resources at once might not properly handle the checkpoint permissions.

## Verifying Permissions

To verify that permissions were created or deleted correctly, you can use the OpenAI API directly:

```bash
curl https://api.openai.com/v1/fine_tuning/checkpoints/ft:gpt-4o-mini-2024-07-18:org-xyz::ABcDeFgH/permissions \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY"
```

## Troubleshooting

### 400 Bad Request - Cannot modify checkpoint permissions for the owning project

If you see this error:

```
Cannot modify checkpoint permissions for the owning project: proj_xyz123
```

This means you're trying to grant permission to the same project that already owns the checkpoint. You need to specify a different project that you want to share the checkpoint with.

### 401 Unauthorized - Missing scopes

If you see an error like:

```
error creating checkpoint permission: 401 Unauthorized - You have insufficient permissions for this operation. 
Missing scopes: api.fine_tuning.checkpoints.write. 
```

This indicates your API key lacks the necessary permissions. Ensure:

1. You're using an admin API key, not a regular API key
2. The key has the `api.fine_tuning.checkpoints.write` scope
3. Your user has the Owner role in the organization
4. You're passing the admin key via the `admin_api_key` variable

### Resource Not Created

If the resource isn't created at all, check:

1. The `use_admin_key` variable is set to `true`
2. Both `checkpoint_id` and `project_id` variables are properly set

## Example Resource Configuration

```hcl
resource "openai_fine_tuning_checkpoint_permission" "example" {
  count = var.use_admin_key && var.checkpoint_id != "" && var.project_id != "" ? 1 : 0
  
  checkpoint_id = var.checkpoint_id
  project_ids   = [var.project_id]
  admin_api_key = var.admin_api_key
}
```

### Multiple Projects Example

To share a checkpoint with multiple projects:

```hcl
resource "openai_fine_tuning_checkpoint_permission" "multi_project" {
  count = var.use_admin_key && var.checkpoint_id != "" ? 1 : 0
  
  checkpoint_id = var.checkpoint_id
  project_ids   = [
    var.project_id_1,
    var.project_id_2,
    var.project_id_3
  ]
  admin_api_key = var.admin_api_key
}
```

## Best Practices

1. Never hardcode admin API keys in your Terraform configurations
2. Use separate API keys for different operations based on required permissions
3. Consider using Terraform workspaces to separate checkpoint operations from other operations
4. Set appropriate timeouts and retry logic for long-running operations
5. Always use the `-target` flag when destroying checkpoint permissions 

# Terraform State Management

## Managing Existing Resources

Terraform state management is critical when working with OpenAI resources, especially for resources like fine-tuning jobs and checkpoint permissions that are often created outside of Terraform.

### State Commands for Checkpoint Permissions

#### 1. Show Resource Details

To view the current state of a checkpoint permission resource:

```bash
terraform state show 'openai_fine_tuning_checkpoint_permission.checkpoint_permission[0]'
```

Example output:
```
# openai_fine_tuning_checkpoint_permission.checkpoint_permission[0]:
resource "openai_fine_tuning_checkpoint_permission" "checkpoint_permission" {
    admin_api_key = (sensitive value)
    checkpoint_id = "ft:gpt-4o-2024-08-06:org-name::AbCdEfGh"
    created_at    = 1743699340
    id            = "cp_AbCdEfGhIjKlMnOp123456"
    project_ids   = [
        "proj_AbCdEfGhIjKlMnOp",
    ]
}
```

#### 2. Remove Resource from State

To remove a resource from Terraform state WITHOUT destroying the actual resource:

```bash
terraform state rm 'openai_fine_tuning_checkpoint_permission.checkpoint_permission[0]'
```

This is useful when you need to reimport a resource or when you want to stop managing a resource with Terraform.

#### 3. Import Existing Checkpoint Permissions

To import a checkpoint permission that was created outside of Terraform:

```bash
terraform import \
  -var="admin_api_key=$OPENAI_ADMIN_KEY" \
  -var="checkpoint_id=ft:gpt-4o-2024-08-06:org-name::AbCdEfGh" \
  -var="project_id=proj_AbCdEfGhIjKlMnOp" \
  -var="use_admin_key=true" \
  'openai_fine_tuning_checkpoint_permission.checkpoint_permission[0]' cp_AbCdEfGhIjKlMnOp123456
```

Where:
- `cp_AbCdEfGhIjKlMnOp123456` is the actual permission ID you're importing
- The variables passed with `-var` ensure proper API access during import

#### 4. Preventing Resource Recreation

When importing existing resources, it's critical to prevent Terraform from trying to modify them. Always add a `lifecycle` block to the resource definition:

```hcl
resource "openai_fine_tuning_checkpoint_permission" "checkpoint_permission" {
  count = var.use_admin_key && var.checkpoint_id != "" && var.project_id != "" ? 1 : 0
  
  checkpoint_id = var.checkpoint_id
  project_ids   = [var.project_id]
  admin_api_key = var.admin_api_key
  
  # Prevent modifications to imported resources
  lifecycle {
    ignore_changes = all
  }
}
```

### State Commands for Fine-Tuning Jobs

Fine-tuning jobs also support import, but require special handling:

#### 1. Show Job Details

```bash
terraform state show 'openai_fine_tuning_job.basic_example'
```

#### 2. Remove Job from State

```bash
terraform state rm 'openai_fine_tuning_job.basic_example'
```

#### 3. Import Existing Fine-Tuning Job

```bash
terraform import 'openai_fine_tuning_job.basic_example' ftjob-AbCdEfGhIjKlMnOp123456
```

The imported job should have a matching configuration in your .tf files, with a lifecycle block:

```hcl
resource "openai_fine_tuning_job" "basic_example" {
  model         = "gpt-4o-mini-2024-07-18"
  training_file = openai_file.training_file.id
  suffix        = "my-custom-model-v1"
  
  # Prevent modifications to imported resources
  lifecycle {
    ignore_changes = all
  }
}
```

## Complete State Management Workflow

A typical workflow for managing existing OpenAI resources with Terraform:

1. **Identify existing resources** you want to manage with Terraform
2. **Create resource definitions** in your .tf files with the required attributes and lifecycle blocks
3. **Import resources** into Terraform state
4. **Apply the configuration** to confirm that no changes are planned
5. **Continue managing** through Terraform, primarily for visibility and documentation

This approach allows you to gradually bring existing OpenAI resources under Terraform management without risking their recreation or modification. 