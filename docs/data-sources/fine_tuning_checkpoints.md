---
page_title: "openai_fine_tuning_checkpoints Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  Provides a list of checkpoints for a specific fine-tuning job.
---

# Data Source: openai_fine_tuning_checkpoints

This data source provides a list of checkpoints for a specific fine-tuning job from your OpenAI account.

## Example Usage

```terraform
data "openai_fine_tuning_checkpoints" "example" {
  fine_tuning_job_id = "ftjob-abc123"
}

output "job_checkpoints" {
  value = data.openai_fine_tuning_checkpoints.example.checkpoints
}
```

## Argument Reference

* `fine_tuning_job_id` - (Required) The ID of the fine-tuning job to retrieve checkpoints for.
* `after` - (Optional) Identifier for the last checkpoint from a previous request. Use this to get the next batch of checkpoints.
* `limit` - (Optional) Number of checkpoints to retrieve (1-100).

## Attributes Reference

The following attributes are exported:

* `id` - An identifier for this data source.
* `checkpoints` - A list of checkpoints for the fine-tuning job. Each checkpoint has the following attributes:
  * `checkpoint_id` - The ID of the checkpoint.
  * `step` - The step at which the checkpoint was created.
  * `train_loss` - The training loss at this checkpoint.
  * `valid_loss` - The validation loss at this checkpoint.
  * `created_at` - The Unix timestamp (in seconds) for when the checkpoint was created.
* `has_more` - Whether there are more checkpoints available than returned in this request.

## Import

This is a read-only data source and cannot be imported. 