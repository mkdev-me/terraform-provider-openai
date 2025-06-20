---
page_title: "OpenAI: openai_audio_transcription Data Source"
subcategory: ""
description: |-
  Data source for an OpenAI audio transcription.
---

# Data Source: openai_audio_transcription

This data source provides a placeholder for documentation and import purposes for OpenAI audio transcriptions. Since the OpenAI API does not provide a way to retrieve transcriptions after they've been created, this data source primarily serves as a reference.

## Example Usage

```terraform
# Reference an existing audio transcription by ID
data "openai_audio_transcription" "example" {
  transcription_id = "transcription-1234567890"
  
  # These fields can be set for documentation purposes
  model = "whisper-1"
  text = "This is the transcribed text from the audio file."
  duration = 120.5
  
  # Optional segments information
  segments = [
    {
      id = 0
      start = 0.0
      end = 10.5
      text = "This is the first segment of transcribed text."
    },
    {
      id = 1
      start = 10.5
      end = 20.2
      text = "This is the second segment of transcribed text."
    }
  ]
}
```

## Argument Reference

* `transcription_id` - (Required) The ID of the audio transcription.
* `model` - (Optional) The model used for audio transcription.
* `text` - (Optional) The transcribed text from the audio.
* `duration` - (Optional) The duration of the audio file in seconds.
* `segments` - (Optional) The segments of the audio transcription, with timing information. Each segment contains:
  * `id` - The ID of the segment.
  * `start` - The start time of the segment in seconds.
  * `end` - The end time of the segment in seconds.
  * `text` - The transcribed text for this segment.

## Attribute Reference

* `id` - The ID of the transcription (same as transcription_id). 