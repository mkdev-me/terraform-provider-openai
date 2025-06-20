# OpenAI Audio Examples

This directory contains examples demonstrating the usage of OpenAI's audio processing capabilities through the Terraform provider.

## Prerequisites

- Terraform installed
- OpenAI API key set in environment variable: `OPENAI_API_KEY`
- Sample audio file in the `samples` directory (speech.mp3)

## Important: API Limitations for Audio Resources

### No List Operations Support

**IMPORTANT**: The OpenAI API does **not** support operations that list all audio resources. While you can create audio resources and retrieve individual resources by ID, you cannot list all resources of a particular type. The following data sources will result in errors:

- `openai_audio_transcriptions` (plural)
- `openai_audio_translations` (plural)
- `openai_speech_to_texts` (plural)
- `openai_text_to_speechs` (plural)

Only individual resource retrieval is supported via:
- `openai_audio_transcription` (singular)
- `openai_audio_translation` (singular)
- `openai_speech_to_text` (singular)
- `openai_text_to_speech` (singular)

### Immutable Resources

The audio resources (`openai_speech_to_text`, `openai_audio_transcription`, `openai_audio_translation`) are **immutable** - they do not support updates. Any configuration change will create a new resource.

### Handling Computed Attributes with Lifecycle Blocks

To prevent Terraform from attempting to update computed attributes (which would fail), you should use lifecycle blocks:

```hcl
resource "openai_audio_transcription" "example" {
  # ... configuration ...
  
  lifecycle {
    ignore_changes = [text, duration, segments]
  }
}

resource "openai_audio_translation" "example" {
  # ... configuration ...
  
  lifecycle {
    ignore_changes = [text, duration, segments]
  }
}

resource "openai_speech_to_text" "example" {
  # ... configuration ...
  
  lifecycle {
    ignore_changes = [text, created_at]
  }
}
```

### Using Data Sources with Immutable Resources

Data sources provide a way to read information from existing resources without attempting to modify them:

```hcl
data "openai_audio_transcription" "example_data" {
  transcription_id = openai_audio_transcription.example.id
}

output "transcription_text" {
  value = data.openai_audio_transcription.example_data.text
}
```

## Resources Demonstrated

### 1. Speech-to-Text (`openai_speech_to_text`)

Converts speech to text using OpenAI's models:

```hcl
resource "openai_speech_to_text" "example" {
  model           = "gpt-4o-transcribe"     # Options: gpt-4o-transcribe, gpt-4o-mini-transcribe, whisper-1
  file            = "./samples/speech.mp3"
  language        = "en"                    # Optional, ISO-639-1 format
  prompt          = "This is a sample speech file for transcription"
  response_format = "json"                  # json, text, srt, verbose_json, vtt (only json for gpt-4o models)
  temperature     = 0.2                     # Between 0 and 1
  
  # Advanced options
  # include      = ["logprobs"]            # Optional, only for gpt-4o models with json response
  # stream       = false                   # Optional, streaming not supported for whisper-1
  # timestamp_granularities = ["segment"]  # Optional, requires verbose_json response format
  
  lifecycle {
    ignore_changes = [text, created_at]
  }
}
```

### 2. Audio Transcription (`openai_audio_transcription`)

Transcribes audio files to text:

```hcl
resource "openai_audio_transcription" "example" {
  model           = "whisper-1"             # Options: gpt-4o-transcribe, gpt-4o-mini-transcribe, whisper-1
  file            = "./samples/speech.mp3"
  language        = "en"                    # Optional, ISO-639-1 format
  prompt          = "This is a sample audio file for transcription"
  response_format = "json"                  # Options: json, text, srt, verbose_json, vtt
  temperature     = 0.2                     # Between 0 and 1
  
  # Advanced options
  # include      = ["logprobs"]            # Optional, only for gpt-4o models with json response
  # stream       = false                   # Optional, streaming not supported for whisper-1
  # timestamp_granularities = ["segment"]  # Optional, requires verbose_json response format
  
  lifecycle {
    ignore_changes = [text, duration, segments]
  }
}
```

### 3. Audio Translation (`openai_audio_translation`)

Translates audio from any language to English:

```hcl
resource "openai_audio_translation" "example" {
  model           = "whisper-1"                     # Only whisper-1 is currently available
  file            = "./samples/speech.mp3"
  prompt          = "This is a sample audio file for translation"  # Should be in English
  response_format = "json"                          # Options: json, text, srt, verbose_json, vtt
  temperature     = 0.2                             # Between 0 and 1
  
  lifecycle {
    ignore_changes = [text, duration, segments]
  }
}
```

### 4. Text-to-Speech (`openai_text_to_speech`)

Converts text to spoken audio:

```hcl
resource "openai_text_to_speech" "example" {
  model           = "tts-1"                        # Options: tts-1, tts-1-hd, gpt-4o-mini-tts
  input           = "Hello, this is a sample text for speech synthesis."  # Max 4096 characters
  voice           = "alloy"                        # Options: alloy, ash, echo, fable, onyx, nova, shimmer, etc.
  response_format = "mp3"                          # Options: mp3, opus, aac, flac, wav, pcm
  speed           = 1.0                            # Speed control, between 0.25 and 4.0
  
  # Advanced options
  # instructions  = "Speak in a calm and clear voice" # Does not work with tts-1 or tts-1-hd
}
```

## Usage

1. Ensure you have your OpenAI API key set:
```bash
export OPENAI_API_KEY="your-api-key"
```

2. Initialize Terraform:
```bash
terraform init
```

3. Apply the configuration:
```bash
terraform apply
```

## Outputs

The example provides the following outputs:

- `speech_to_text_text`: The transcribed text from speech-to-text conversion
- `speech_to_text_created_at`: Timestamp of speech-to-text generation
- `audio_transcription_text`: The transcribed text from audio transcription
- `audio_transcription_duration`: Duration of the audio file in seconds
- `audio_translation_text`: The translated text from audio translation
- `audio_translation_duration`: Duration of the audio file in seconds
- `text_to_speech_output`: The path to the generated audio file

## Supported Audio Formats

### For Speech-to-Text and Audio Transcription/Translation:
- FLAC
- MP3
- MP4
- MPEG
- MPGA
- M4A
- OGG
- WAV
- WEBM

### For Text-to-Speech Output:
- MP3
- OPUS
- AAC
- FLAC
- WAV
- PCM

## Data Sources

The provider also includes data sources for working with audio files and transcriptions:

### API Limitation: Listing Operations Not Supported

**Important**: The OpenAI API does not currently support listing operations for audio resources. While you can create audio resources and retrieve individual resources by ID, the API doesn't provide endpoints for retrieving lists of resources.

The following data sources will return error messages when used:

- `openai_audio_transcriptions` - Cannot list all transcriptions
- `openai_audio_translations` - Cannot list all translations
- `openai_speech_to_texts` - Cannot list all speech-to-text conversions
- `openai_text_to_speechs` - Cannot list all text-to-speech conversions

These limitations are imposed by the OpenAI API, not the Terraform provider. See `data_sources.tf` for examples that show both the unsupported operations and the supported alternatives.

### Supported Data Sources

The following individual resource data sources are fully supported:

### 1. Text-to-Speech Data Source (`openai_text_to_speech`)

Verifies and provides metadata about a text-to-speech file:

```hcl
data "openai_text_to_speech" "example" {
  file_path = "./output/speech.mp3"
}

output "speech_file_exists" {
  value = data.openai_text_to_speech.example.exists
}

output "speech_file_size" {
  value = data.openai_text_to_speech.example.file_size_bytes
}
```

### 2. Speech-to-Text Data Source (`openai_speech_to_text`)

References an existing speech-to-text transcription:

```hcl
data "openai_speech_to_text" "example" {
  transcription_id = openai_speech_to_text.example.id
}
```

### 3. Audio Transcription Data Source (`openai_audio_transcription`)

References an existing audio transcription:

```hcl
data "openai_audio_transcription" "example" {
  transcription_id = openai_audio_transcription.example.id
}
```

### 4. Audio Translation Data Source (`openai_audio_translation`)

References an existing audio translation:

```hcl
data "openai_audio_translation" "example" {
  translation_id = openai_audio_translation.example.id
}
```

## Example Files

This directory contains several example files demonstrating different aspects of the audio functionality:

1. **main.tf** - Basic usage of the audio resources
2. **data_sources.tf** - Basic examples of using the audio data sources
3. **complete_example.tf** - A complete workflow combining all audio resources and data sources
4. **module_example.tf** - Using the audio module with data source features
5. **import_example.tf** - Importing existing audio resources using data sources

### Complete Workflow Example

The `complete_example.tf` file demonstrates a complete workflow:
1. Generate speech from text
2. Verify the generated file
3. Convert the speech back to text
4. Transcribe and translate the audio
5. Use data sources to access the created resources

To use this example:

```bash
terraform apply -target=openai_text_to_speech.workflow -target=data.openai_text_to_speech.workflow_tts_file
terraform apply
```

### Module Usage Example

The `module_example.tf` file shows how to use the audio module with all capabilities enabled:

```bash
terraform apply -target=module.audio_example
```

### Import Example

The `import_example.tf` file demonstrates how to use data sources to reference existing resources:

```bash
terraform apply -target=data.openai_text_to_speech.imported_tts
```

## Importing Existing Resources

As of the latest version, all audio resources now support importing. This allows you to bring existing audio resources that were created outside of Terraform into Terraform management.

### Import Syntax

```bash
# Import an audio transcription
terraform import openai_audio_transcription.example transcription-1234567890

# Import an audio translation
terraform import openai_audio_translation.example translation-1234567890

# Import a speech-to-text transcription
terraform import openai_speech_to_text.example transcription-1234567890

# Import a text-to-speech resource
terraform import openai_text_to_speech.example speech-1234567890
```

### Import Workflow

1. Define the resource in your Terraform configuration:
```hcl
resource "openai_audio_transcription" "example" {
  model           = "whisper-1"
  file            = "./samples/speech.mp3"
  language        = "en"
  prompt          = "This is a sample audio file for transcription"
  response_format = "json"
  temperature     = 0.2
  
  lifecycle {
    ignore_changes = [text, duration, segments]
  }
}
```

2. Import the existing resource:
```bash
terraform import openai_audio_transcription.example transcription-1234567890
```

3. Verify the import with:
```bash
terraform state show openai_audio_transcription.example
```

4. Run terraform plan to ensure no changes are needed:
```bash
terraform plan
```

### Notes on Importing

- When importing, the provider will set reasonable default values for required fields.
- You may need to adjust these values in your configuration to match your actual requirements.
- Using `lifecycle { ignore_changes = [...] }` prevents Terraform from attempting to update computed attributes.

## Notes

- The examples use appropriate models for each task
- Temperature controls the randomness of the output (0.0 to 1.0)
- Response format options vary by resource and model
- Language parameter is optional and should be in ISO-639-1 format
- Some advanced features like `include`, `stream`, and `timestamp_granularities` are model-specific
- Use `depends_on` when referencing resources with data sources to ensure proper sequencing 