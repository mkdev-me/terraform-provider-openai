---
page_title: "OpenAI: openai_speech_to_text"
subcategory: ""
description: |-
  Data source for an OpenAI speech-to-text transcription.
---

# Data Source: openai_speech_to_text

This data source provides a placeholder for documentation and import purposes for OpenAI speech-to-text transcriptions. Since the OpenAI API does not provide a way to retrieve transcriptions after they've been created, this data source primarily serves as a reference.

## Example Usage

```terraform
# Reference an existing speech-to-text transcription by ID
data "openai_speech_to_text" "example" {
  transcription_id = "transcription-1234567890"
  
  # These fields can be set for documentation purposes
  model = "whisper-1"
  text = "This is the transcribed text from the audio file."
  created_at = 1680000000
}
```

## Argument Reference

* `transcription_id` - (Required) The ID of the speech-to-text transcription.
* `model` - (Optional) The model used for transcription.
* `text` - (Optional) The transcribed text.
* `created_at` - (Optional) The timestamp when the transcription was generated.

## Attribute Reference

* `id` - The ID of the transcription (same as transcription_id). 