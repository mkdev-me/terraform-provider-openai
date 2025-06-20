---
page_title: "openai_fine_tuning_events Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  Provides a list of events for a specific fine-tuning job.
---

# Data Source: openai_fine_tuning_events

This data source provides a list of events for a specific fine-tuning job from your OpenAI account. Events provide information about the fine-tuning job's progress and any issues that may arise during training.

## Example Usage

```terraform
data "openai_fine_tuning_events" "example" {
  fine_tuning_job_id = "ftjob-abc123"
}

output "job_events" {
  value = data.openai_fine_tuning_events.example.events
}
```

## Argument Reference

* `fine_tuning_job_id` - (Required) The ID of the fine-tuning job to retrieve events for.
* `after` - (Optional) Identifier for the last event from a previous request. Use this to get the next batch of events.
* `limit` - (Optional) Number of events to retrieve (1-100).

## Attributes Reference

The following attributes are exported:

* `id` - An identifier for this data source.
* `events` - A list of events for the fine-tuning job. Each event has the following attributes:
  * `object` - The object type, which is always "fine_tuning.job.event".
  * `id` - The ID of the event.
  * `created_at` - The Unix timestamp (in seconds) for when the event was created.
  * `level` - The level of the event, which can be "info", "warning", or "error".
  * `message` - The message associated with the event, providing details about what happened.
  * `data` - Additional data for the event, which varies depending on the event type.
    * `step` - Current step during training.
    * `train_loss` - Training loss at this step.
    * `valid_loss` - Validation loss at this step.
* `has_more` - Whether there are more events available than returned in this request.

## Import

This is a read-only data source and cannot be imported. 