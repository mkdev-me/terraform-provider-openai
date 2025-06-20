# Resource: openai_audio_translation

The `openai_audio_translation` resource allows you to translate audio from any language into English text using OpenAI's Whisper model. This resource supports various audio formats and provides options for translation quality and output format.

> **Important**: This resource is immutable. Any configuration change will create a new resource rather than updating an existing one.

## Example Usage

```hcl
resource "openai_audio_translation" "example" {
  model           = "whisper-1"
  file            = "/path/to/audio/file.mp3"
  prompt          = "Translation of a business meeting"
  response_format = "verbose_json"
  temperature     = 0
  
  # Prevent Terraform from attempting to update computed attributes
  lifecycle {
    ignore_changes = [text, duration, segments]
  }
}

# Use a data source to access the resource attributes
data "openai_audio_translation" "example_data" {
  translation_id = openai_audio_translation.example.id
}

output "translation" {
  value = openai_audio_translation.example.text
}

output "segments" {
  value = openai_audio_translation.example.segments
}
```

## Argument Reference

* `model` - (Required) The ID of the model to use. Currently only `whisper-1` is supported.

* `file` - (Required) The path to the audio file to translate. Must be a supported format (mp3, mp4, mpeg, mpga, m4a, wav, or webm).

* `prompt` - (Optional) An optional text to guide the model's style or continue a previous audio segment.

* `response_format` - (Optional) The format of the translation output. Options are `json`, `text`, `srt`, `verbose_json`, or `vtt`. Default is `json`. Use `verbose_json` to get detailed segmentation information.

* `temperature` - (Optional) The sampling temperature, between 0 and 1. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic. Default is 0.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A unique identifier for this translation.

* `text` - The translated text (in English).

* `duration` - The duration of the audio file in seconds.

* `segments` - A list of segments, each containing:
  * `id` - The ID of the segment.
  * `start` - The start time of the segment in seconds.
  * `end` - The end time of the segment in seconds.
  * `text` - The translated text of the segment.

## Import

This resource can be imported using the translation ID:

```shell
terraform import openai_audio_translation.example translation-123456789
```

When importing, the resource will be populated with default values for required fields. You may need to adjust these values in your configuration to match your actual requirements.

## Notes

* The maximum file size is 25 MB.
* The maximum audio duration is 6 hours for paid API users.
* The `whisper-1` model can translate from many languages to English.
* Unlike the transcription API, the translation API will not provide the transcript in the original language and will only return English.
* For best results, provide the original audio with minimal background noise and clear speech.
* To get detailed segmentation information, use `response_format = "verbose_json"`.
* The temperature parameter can be adjusted if you need more creative or varied outputs, but lower values generally produce more accurate translations.
* **Immutability**: This resource is immutable. Any changes to its configuration will result in a new resource being created rather than updating the existing one. Use the `lifecycle` block with `ignore_changes` for computed attributes to prevent unnecessary recreation. 