# Resource: openai_text_to_speech

The `openai_text_to_speech` resource allows you to convert text into natural-sounding spoken audio using OpenAI's text-to-speech models. This resource provides options for voice selection, speech quality, and output format.

## Example Usage

```hcl
resource "openai_text_to_speech" "example" {
  model           = "tts-1"
  input           = "Hello world! This is text being converted to natural-sounding speech using OpenAI's text-to-speech API."
  voice           = "alloy"
  response_format = "mp3"
  speed           = 1.0
  output_file     = "/path/to/output/speech.mp3"
}

output "created_at" {
  value = openai_text_to_speech.example.created_at
}
```

## Argument Reference

* `model` - (Required) The ID of the model to use. Options are `tts-1` (standard quality) and `tts-1-hd` (higher quality).

* `input` - (Required) The text to convert to speech. Maximum length is 4096 characters.

* `voice` - (Required) The voice to use for the spoken audio. Options are:
  * `alloy` - A neutral voice with mild emotional range
  * `echo` - A deeper, serious voice
  * `fable` - A soft, expressive voice
  * `onyx` - A deep, authoritative voice
  * `nova` - A professional, encouraging voice
  * `shimmer` - A bright, warm voice

* `response_format` - (Optional) The format of the audio output. Options are `mp3`, `opus`, `aac`, and `flac`. Default is `mp3`.

* `speed` - (Optional) The speed of the generated audio. Must be between 0.25 and 4.0, with 1.0 being the default.

* `output_file` - (Required) The path where the generated audio file will be saved.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A unique identifier for this text-to-speech generation.

* `created_at` - The timestamp when the speech was generated.

## Import

This resource can be imported using the speech ID:

```shell
terraform import openai_text_to_speech.example speech-123456789
```

When importing, the resource will be populated with default values for required fields. You may need to adjust these values in your configuration to match your actual requirements.

## Notes

* The maximum input length is 4096 characters.
* The `tts-1` model provides standard quality audio sufficient for most use cases.
* The `tts-1-hd` model provides higher quality audio with more natural prosody and pronunciation.
* Each voice has its own characteristics and tone. Test different voices to find the best fit for your application.
* The speed parameter can be used to adjust the pace of speech to better fit the content's context.
* For longer text, consider splitting into smaller chunks to stay within the character limit.
* The resource automatically creates the directory structure if it doesn't exist for the output file.
* On resource deletion, the generated audio file is also removed. 