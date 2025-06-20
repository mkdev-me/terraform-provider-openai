# Resource: openai_project

The `openai_project` resource allows you to manage projects in OpenAI. Projects in OpenAI are containers that help organize models, fine-tuning, and other functionalities.

## Example Usage

```hcl
resource "openai_project" "example" {
  name        = "my-production-project"
}
```

## Import

This resource can be imported using the project ID, for example:

```
terraform import openai_project.example proj_abc123
```


## Arguments

* `name` - (Required) The name of the project.


-> **Note:** For more granular rate limiting by model, use the [`openai_rate_limit`](rate_limit.md) resource instead, which supports additional parameters such as `max_requests_per_minute`, `max_tokens_per_minute`, `max_images_per_minute`, `max_audio_megabytes_per_1_minute`, `max_requests_per_1_day`, and `batch_1_day_max_input_tokens`.

## Attributes

* `id` - The unique ID of the project assigned by OpenAI.
* `created_at` - The timestamp when the project was created.
* `archived_at` - The timestamp when the project was archived (if applicable).
* `status` - The current status of the project (e.g., "active", "archived").
* `billing_mode` - The billing mode configured for the project.
* `api_keys` - A list of API keys associated with this project.
  * `id` - The ID of the API key.
  * `name` - The name of the API key.
  * `created_at` - The timestamp when the API key was created.
  * `last_used_at` - The timestamp when the API key was last used.

## Archiving Operations

When using `terraform destroy` on a project, OpenAI does not permanently delete the project but archives it instead. This means the project will no longer appear in the list of active projects but will continue to exist in an archived state.

To list archived projects, you can use the OpenAI API with the `include_archived=true` parameter:

```bash
curl "https://api.openai.com/v1/organization/projects?include_archived=true" \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
  -H "Content-Type: application/json"
```

## Important Notes

* To manage projects, you need an API key with administrator permissions (`api.management.read` and `api.management.write`).
* Archived projects cannot be "unarchived" through the API. You would need to use the OpenAI web interface for that operation.
