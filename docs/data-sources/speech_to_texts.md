---
page_title: "OpenAI: openai_speech_to_texts"
subcategory: ""
description: |-
  Data source for OpenAI speech-to-text conversions (Not supported by OpenAI API).
---

# Data Source: openai_speech_to_texts

> **Important Note:** This data source is included for documentation purposes, but it will result in an error when used. The OpenAI API does not currently support listing all speech-to-text conversions. You can only retrieve individual conversions using the `openai_speech_to_text` (singular) data source with a specific `transcription_id`.

This data source would theoretically allow you to list all speech-to-text conversions, but the OpenAI API does not provide an endpoint for this operation.

## Example Usage (Will Result in Error)

```terraform
data "openai_speech_to_texts" "all" {
  # Optional filter by model (not supported)
  model = "whisper-1"
}

# This will fail with an error from the OpenAI API
output "all_transcriptions" {
  value = data.openai_speech_to_texts.all.speech_to_texts
}
```

## Supported Alternative

Instead of trying to list all speech-to-text conversions, you should use the `openai_speech_to_text` (singular) data source to retrieve a specific conversion by ID:

```terraform
data "openai_speech_to_text" "example" {
  transcription_id = "transcription-1234567890"
}

output "transcription_text" {
  value = data.openai_speech_to_text.example.text
}
```

## API Limitation Details

The OpenAI API does not provide endpoints for listing operations for audio resources. This is a limitation of the API itself, not the Terraform provider. If you need to track multiple speech-to-text conversions, consider implementing your own tracking system outside of this provider.

For more information about working with individual speech-to-text conversions, see the [openai_speech_to_text](./speech_to_text.md) data source documentation. 