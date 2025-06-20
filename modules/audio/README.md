# OpenAI Audio Module

This Terraform module provides easy-to-use resources for OpenAI's audio capabilities, including:

- Speech-to-Text: Convert spoken language in audio files to written text
- Audio Transcription: Transcribe audio files with detailed information
- Audio Translation: Translate audio files from any language to English
- Text-to-Speech: Convert text to natural-sounding speech

## Usage

```hcl
module "audio" {
  source = "path/to/modules/audio"

  # Speech-to-Text Configuration
  enable_speech_to_text      = true
  speech_to_text_model       = "whisper-1"
  speech_to_text_file        = "path/to/audio/file.mp3"
  speech_to_text_language    = "en"  # Optional: ISO-639-1 format
  speech_to_text_prompt      = "Transcript of a meeting"  # Optional
  speech_to_text_temperature = 0     # Optional: 0-1 range

  # Audio Transcription Configuration 
  enable_audio_transcription      = true
  audio_transcription_model       = "whisper-1"
  audio_transcription_file        = "path/to/audio/file.mp3"
  audio_transcription_language    = "en"  # Optional: ISO-639-1 format
  audio_transcription_prompt      = "Transcript of a lecture"  # Optional
  audio_transcription_temperature = 0     # Optional: 0-1 range

  # Audio Translation Configuration
  enable_audio_translation      = true
  audio_translation_model       = "whisper-1"
  audio_translation_file        = "path/to/audio/file.mp3"
  audio_translation_prompt      = "Translation of a speech"  # Optional
  audio_translation_temperature = 0     # Optional: 0-1 range

  # Text-to-Speech Configuration
  enable_text_to_speech      = true
  text_to_speech_model       = "tts-1"  # Options: tts-1, tts-1-hd
  text_to_speech_input       = "Hello world! This is text being converted to speech."
  text_to_speech_voice       = "alloy"  # Options: alloy, echo, fable, onyx, nova, shimmer
  text_to_speech_speed       = 1.0      # Optional: 0.25-4.0 range
  text_to_speech_output_file = "path/to/output/speech.mp3"
}
```

## Requirements

- Terraform >= 0.13.0
- OpenAI Provider

## Input Variables

### Speech-to-Text Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| enable_speech_to_text | Whether to enable speech-to-text conversion | `bool` | `false` | no |
| speech_to_text_model | Model to use for speech-to-text | `string` | `"whisper-1"` | no |
| speech_to_text_file | Path to the audio file to transcribe | `string` | `""` | yes (if enabled) |
| speech_to_text_language | Language of the input audio (ISO-639-1) | `string` | `null` | no |
| speech_to_text_prompt | Text to guide the model's style | `string` | `null` | no |
| speech_to_text_response_format | Format of transcript output | `string` | `"json"` | no |
| speech_to_text_temperature | Sampling temperature (0-1) | `number` | `0` | no |

### Audio Transcription Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| enable_audio_transcription | Whether to enable audio transcription | `bool` | `false` | no |
| audio_transcription_model | Model to use for transcription | `string` | `"whisper-1"` | no |
| audio_transcription_file | Path to the audio file to transcribe | `string` | `""` | yes (if enabled) |
| audio_transcription_language | Language of the input audio (ISO-639-1) | `string` | `null` | no |
| audio_transcription_prompt | Text to guide the model's style | `string` | `null` | no |
| audio_transcription_response_format | Format of transcript output | `string` | `"json"` | no |
| audio_transcription_temperature | Sampling temperature (0-1) | `number` | `0` | no |

### Audio Translation Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| enable_audio_translation | Whether to enable audio translation | `bool` | `false` | no |
| audio_translation_model | Model to use for translation | `string` | `"whisper-1"` | no |
| audio_translation_file | Path to the audio file to translate | `string` | `""` | yes (if enabled) |
| audio_translation_prompt | Text to guide the model's style | `string` | `null` | no |
| audio_translation_response_format | Format of translation output | `string` | `"json"` | no |
| audio_translation_temperature | Sampling temperature (0-1) | `number` | `0` | no |

### Text-to-Speech Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| enable_text_to_speech | Whether to enable text-to-speech | `bool` | `false` | no |
| text_to_speech_model | Model to use for text-to-speech | `string` | `"tts-1"` | no |
| text_to_speech_input | Text to convert to speech | `string` | `""` | yes (if enabled) |
| text_to_speech_voice | Voice to use for speech | `string` | `"alloy"` | no |
| text_to_speech_response_format | Format of audio output | `string` | `"mp3"` | no |
| text_to_speech_speed | Speed of audio output (0.25-4.0) | `number` | `1.0` | no |
| text_to_speech_output_file | Path to save the audio file | `string` | `""` | yes (if enabled) |

### Data Source Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| verify_tts_file | Whether to verify the TTS output file | `bool` | `false` | no |
| use_stt_data_source | Whether to use STT data source | `bool` | `false` | no |
| existing_stt_id | ID of existing STT resource to import | `string` | `""` | no |
| use_transcription_data_source | Whether to use transcription data source | `bool` | `false` | no |
| existing_transcription_id | ID of existing transcription to import | `string` | `""` | no |
| use_translation_data_source | Whether to use translation data source | `bool` | `false` | no |
| existing_translation_id | ID of existing translation to import | `string` | `""` | no |

## Outputs

| Name | Description |
|------|-------------|
| speech_to_text_text | Transcribed text from speech-to-text |
| speech_to_text_created_at | Timestamp of speech-to-text generation |
| speech_to_text_id | ID of the speech-to-text resource |
| audio_transcription_text | Transcribed text from audio transcription |
| audio_transcription_duration | Duration of audio file in seconds |
| audio_transcription_id | ID of the audio transcription resource |
| audio_translation_text | Translated text from audio translation |
| audio_translation_duration | Duration of audio file in seconds |
| audio_translation_id | ID of the audio translation resource |
| text_to_speech_created_at | Timestamp of text-to-speech generation |
| text_to_speech_output_file | Path to the generated audio file |
| tts_file_exists | Whether the TTS file exists (from data source) |
| tts_file_size | Size of the TTS file in bytes (from data source) |
| tts_file_last_modified | Last modified timestamp of TTS file (from data source) |
| stt_data_source_id | ID from the STT data source |
| transcription_data_source_id | ID from the transcription data source |
| translation_data_source_id | ID from the translation data source |

## Using Data Sources

This module supports both creating new audio resources and importing existing ones:

```hcl
module "audio" {
  source = "path/to/modules/audio"

  # Create and verify a text-to-speech file
  enable_text_to_speech    = true
  text_to_speech_input     = "Hello world!"
  text_to_speech_output_file = "./speech.mp3"
  verify_tts_file          = true  # Enable TTS data source
  
  # Import an existing speech-to-text resource
  use_stt_data_source      = true
  existing_stt_id          = "transcription-1234567890"
  
  # Create a new audio transcription and use data source
  enable_audio_transcription = true
  audio_transcription_file = "./speech.mp3"
  use_transcription_data_source = true
}
```

## API Limitations

### List Operations Not Supported

**Important**: The OpenAI API does not currently support listing operations for audio resources. The following variables and data sources are included in the module but are commented out in the implementation because the OpenAI API does not provide endpoints for these operations:

| Name | Description | Status |
|------|-------------|--------|
| retrieve_all_text_to_speech | Whether to retrieve all text-to-speech conversions | Not supported by API |
| filter_tts_by_model | Filter text-to-speech by model | Not supported by API |
| filter_tts_by_voice | Filter text-to-speech by voice | Not supported by API |
| retrieve_all_speech_to_text | Whether to retrieve all speech-to-text conversions | Not supported by API |
| filter_stt_by_model | Filter speech-to-text by model | Not supported by API |
| retrieve_all_transcriptions | Whether to retrieve all transcriptions | Not supported by API |
| filter_transcriptions_by_model | Filter transcriptions by model | Not supported by API |
| retrieve_all_translations | Whether to retrieve all translations | Not supported by API |
| filter_translations_by_model | Filter translations by model | Not supported by API |

These limitations are imposed by the OpenAI API, not by the Terraform provider. If you need to manage multiple audio resources, consider implementing your own tracking system outside of the provider.

### Supported Data Source Operations

The module fully supports retrieving individual audio resources using their specific IDs:

- `openai_text_to_speech` - Retrieve information about a specific text-to-speech file
- `openai_speech_to_text` - Retrieve a specific speech-to-text conversion by ID
- `openai_audio_transcription` - Retrieve a specific transcription by ID
- `openai_audio_translation` - Retrieve a specific translation by ID

For more examples, see the `examples/audio` directory in the provider repository. 