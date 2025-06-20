---
page_title: "openai_fine_tuning_jobs Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  Provides a list of fine-tuning jobs.
---

# Data Source: openai_fine_tuning_jobs

This data source provides a list of fine-tuning jobs from your OpenAI account.

## Example Usage

```terraform
data "openai_fine_tuning_jobs" "all" {}

output "all_jobs" {
  value = data.openai_fine_tuning_jobs.all.jobs
}
```

## Attributes Reference

The following attributes are exported:

* `id` - An identifier for this data source.
* `jobs` - A list of fine-tuning jobs. Each job has the following attributes:
  * `id` - The ID of the fine-tuning job.
  * `object` - The object type, which is always "fine_tuning.job".
  * `model` - The base model that is being fine-tuned.
  * `created_at` - The Unix timestamp (in seconds) for when the fine-tuning job was created.
  * `finished_at` - The Unix timestamp (in seconds) for when the fine-tuning job was finished. Will be null if the job is still running.
  * `fine_tuned_model` - The name of the fine-tuned model that is being created. Will be null if the job is still running.
  * `organization_id` - The organization that owns the fine-tuning job.
  * `status` - The status of the fine-tuning job, which can be either created, pending, running, succeeded, failed, or cancelled.
  * `training_file` - The file ID used for training.
  * `validation_file` - The file ID used for validation.
  * `hyperparameters` - The hyperparameters used for the fine-tuning job.
    * `n_epochs` - The number of epochs to train the model for. An epoch refers to one full cycle through the training dataset.
  * `trained_tokens` - The number of tokens in the training dataset.
  * `error` - In the case of a failed fine-tuning job, this will contain the error message.

## Import

This is a read-only data source and cannot be imported. 