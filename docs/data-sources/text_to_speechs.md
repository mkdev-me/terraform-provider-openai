---
page_title: "OpenAI: openai_text_to_speechs Data Source"
subcategory: ""
description: |-
  Data source for OpenAI text-to-speech conversions (Not supported by OpenAI API).
---

# Data Source: openai_text_to_speechs

> **Important Note:** This data source is included for documentation purposes, but it will result in an error when used. The OpenAI API does not currently support listing all text-to-speech conversions. You can only retrieve individual conversions using the `openai_text_to_speech` (singular) data source with a specific `file_path`.

This data source would theoretically allow you to list all text-to-speech conversions, but the OpenAI API does not provide an endpoint for this operation.

## Example Usage (Will Result in Error)

```terraform
data "openai_text_to_speechs" "all" {
  # Optional filters (not supported)
  model = "tts-1"
  voice = "alloy"
}

# This will fail with an error from the OpenAI API
output "all_tts" {
  value = data.openai_text_to_speechs.all.text_to_speechs
}
```

## Supported Alternative

Instead of trying to list all text-to-speech conversions, you should use the `openai_text_to_speech` (singular) data source to retrieve information about a specific TTS file:

```terraform
data "openai_text_to_speech" "example" {
  file_path = "./output/speech.mp3"
}

output "tts_file_exists" {
  value = data.openai_text_to_speech.example.exists
}

output "tts_file_size" {
  value = data.openai_text_to_speech.example.file_size_bytes
}
```

## API Limitation Details

The OpenAI API does not provide endpoints for listing operations for audio resources. This is a limitation of the API itself, not the Terraform provider. If you need to track multiple text-to-speech conversions, consider implementing your own tracking system outside of this provider.

For more information about working with individual text-to-speech files, see the [openai_text_to_speech](./text_to_speech.md) data source documentation. 