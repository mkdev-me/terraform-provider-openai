---
page_title: "OpenAI: openai_audio_translation Data Source"
subcategory: ""
description: |-
  Data source for an OpenAI audio translation.
---

# Data Source: openai_audio_translation

This data source provides a placeholder for documentation and import purposes for OpenAI audio translations. Since the OpenAI API does not provide a way to retrieve translations after they've been created, this data source primarily serves as a reference.

## Example Usage

```terraform
# Reference an existing audio translation by ID
data "openai_audio_translation" "example" {
  translation_id = "translation-1234567890"
  
  # These fields can be set for documentation purposes
  model = "whisper-1"
  text = "This is the translated text from the audio file."
  duration = 120.5
  
  # Optional segments information
  segments = [
    {
      id = 0
      start = 0.0
      end = 10.5
      text = "This is the first segment of translated text."
    },
    {
      id = 1
      start = 10.5
      end = 20.2
      text = "This is the second segment of translated text."
    }
  ]
}
```

## Argument Reference

* `translation_id` - (Required) The ID of the audio translation.
* `model` - (Optional) The model used for audio translation.
* `text` - (Optional) The translated text from the audio.
* `duration` - (Optional) The duration of the audio file in seconds.
* `segments` - (Optional) The segments of the audio translation, with timing information. Each segment contains:
  * `id` - The ID of the segment.
  * `start` - The start time of the segment in seconds.
  * `end` - The end time of the segment in seconds.
  * `text` - The translated text for this segment.

## Attribute Reference

* `id` - The ID of the translation (same as translation_id). 