---
page_title: "OpenAI: openai_audio_transcriptions Data Source"
subcategory: ""
description: |-
  Data source for OpenAI audio transcriptions (Not supported by OpenAI API).
---

# Data Source: openai_audio_transcriptions

> **Important Note:** This data source is included for documentation purposes, but it will result in an error when used. The OpenAI API does not currently support listing all audio transcriptions. You can only retrieve individual transcriptions using the `openai_audio_transcription` (singular) data source with a specific `transcription_id`.

This data source would theoretically allow you to list all audio transcriptions, but the OpenAI API does not provide an endpoint for this operation.

## Example Usage (Will Result in Error)

```terraform
data "openai_audio_transcriptions" "all" {
  # Optional filter by model (not supported)
  model = "whisper-1"
}

# This will fail with an error from the OpenAI API
output "all_transcriptions" {
  value = data.openai_audio_transcriptions.all.transcriptions
}
```

## Supported Alternative

Instead of trying to list all transcriptions, you should use the `openai_audio_transcription` (singular) data source to retrieve a specific transcription by ID:

```terraform
data "openai_audio_transcription" "example" {
  transcription_id = "transcription-1234567890"
}

output "transcription_text" {
  value = data.openai_audio_transcription.example.text
}
```

## API Limitation Details

The OpenAI API does not provide endpoints for listing operations for audio resources. This is a limitation of the API itself, not the Terraform provider. If you need to track multiple transcriptions, consider implementing your own tracking system outside of this provider.

For more information about working with individual audio transcriptions, see the [openai_audio_transcription](./audio_transcription.md) data source documentation.

## Argument Reference

The following arguments are supported:

* `project_id` - (Optional) The ID of the project to retrieve audio transcriptions from. If not specified, the API key's default project will be used.
* `api_key` - (Optional) Project-specific API key to use for authentication. If not provided, the provider's default API key will be used.
* `model` - (Optional) Filter by model. Options include 'whisper-1'.

## Attributes Reference

The following attributes are exported:

* `id` - A unique identifier for this data source.
* `transcriptions` - A list of audio transcriptions, each containing the following attributes:
  * `id` - The ID of the audio transcription.
  * `created_at` - The timestamp when the transcription was created.
  * `status` - The status of the transcription.
  * `model` - The model used for transcription.
  * `text` - The transcribed text.
  * `duration` - The duration of the audio in seconds.
  * `language` - The language of the transcription (if detected or specified).

## Import

This is a data source and does not support import. 