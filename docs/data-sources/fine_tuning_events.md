---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "openai_fine_tuning_events Data Source - terraform-provider-openai"
subcategory: ""
description: |-
  
---

# openai_fine_tuning_events (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `fine_tuning_job_id` (String) The ID of the fine-tuning job to get events for

### Optional

- `after` (String) Identifier for the last event from the previous pagination request
- `limit` (Number) Number of events to retrieve (default: 20)

### Read-Only

- `events` (List of Object) (see [below for nested schema](#nestedatt--events))
- `has_more` (Boolean) Whether there are more events to retrieve
- `id` (String) The ID of this resource.

<a id="nestedatt--events"></a>
### Nested Schema for `events`

Read-Only:

- `created_at` (Number)
- `data_json` (String)
- `id` (String)
- `level` (String)
- `message` (String)
- `object` (String)
- `type` (String)
