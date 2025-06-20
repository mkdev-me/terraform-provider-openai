terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "1.0.0"
    }
  }
}

provider "openai" {
  # API key sourced from environment variable: OPENAI_API_KEY
}

# Speech-to-Text Example
resource "openai_speech_to_text" "example" {
  model           = "gpt-4o-transcribe" # Options: gpt-4o-transcribe, gpt-4o-mini-transcribe, whisper-1
  file            = "./samples/speech.mp3"
  language        = "en"
  prompt          = "This is a sample speech file for transcription"
  response_format = "json" # Only json supported for gpt-4o models
  temperature     = 0.2
  # New parameters
  # include       = ["logprobs"] # Only works with json response and gpt-4o models
  # stream        = false # Streaming not supported for whisper-1
  # timestamp_granularities = ["segment"] # Requires verbose_json format

  # The warning indicates that text is decided by the provider alone
  # and there can be no configured value to compare with
  lifecycle {
    # Removed text from ignore_changes since it generates a warning
  }
}

# Audio Transcription Example
resource "openai_audio_transcription" "example" {
  model           = "whisper-1"
  file            = "./samples/speech.mp3"
  language        = "en"
  prompt          = "This is a sample audio file for transcription"
  response_format = "json" # Options: json, text, srt, verbose_json, vtt
  temperature     = 0.2
  # New parameters
  # include       = ["logprobs"] # Only works with json response and gpt-4o models
  # stream        = false # Streaming not supported for whisper-1
  # timestamp_granularities = ["segment"] # Requires verbose_json format

  # The warning indicates text is decided by the provider alone
  # Configure the resource as immutable since it doesn't support updates
  lifecycle {
    ignore_changes = [segments] # Only ignore segments to prevent update attempts
    # Text attribute doesn't need to be in ignore_changes (per warning)
    create_before_destroy = true # Treat as immutable - create new before destroying old
  }
}

# Audio Translation Example
resource "openai_audio_translation" "example" {
  model           = "whisper-1" # Only whisper-1 is currently available
  file            = "./samples/speech.mp3"
  prompt          = "This is a sample audio file for translation" # Should be in English
  response_format = "json"                                        # Options: json, text, srt, verbose_json, vtt
  temperature     = 0.2

  # The warning indicates text is decided by the provider alone
  # Configure the resource as immutable since it doesn't support updates
  lifecycle {
    ignore_changes = [segments] # Only ignore segments to prevent update attempts
    # Text attribute doesn't need to be in ignore_changes (per warning)
    create_before_destroy = true # Treat as immutable - create new before destroying old
  }
}

# Text-to-Speech Example
resource "openai_text_to_speech" "example" {
  model           = "tts-1" # Options: tts-1, tts-1-hd, gpt-4o-mini-tts
  input           = "Hello, this is a sample text for speech synthesis."
  voice           = "alloy" # Options: alloy, echo, fable, onyx, nova, shimmer, etc.
  response_format = "mp3"   # Options: mp3, opus, aac, flac, wav, pcm
  speed           = 1.0
  output_file     = "./output/speech.mp3" # Required - path where the audio file will be saved
  # New parameter
  # instructions    = "Speak in a calm and clear voice" # Does not work with tts-1 or tts-1-hd
}

# Data resources to read from the created resources
data "openai_speech_to_text" "example_data" {
  transcription_id = openai_speech_to_text.example.id
}

data "openai_audio_transcription" "example_data" {
  transcription_id = openai_audio_transcription.example.id
}

data "openai_audio_translation" "example_data" {
  translation_id = openai_audio_translation.example.id
}

data "openai_text_to_speech" "example_data" {
  file_path = openai_text_to_speech.example.output_file
}

# Outputs
output "speech_to_text_text" {
  description = "The transcribed text from speech-to-text conversion"
  value       = openai_speech_to_text.example.text
}

output "speech_to_text_created_at" {
  description = "The timestamp when the speech-to-text transcription was generated"
  value       = openai_speech_to_text.example.created_at
}

output "audio_transcription_text" {
  description = "The transcribed text from audio transcription"
  value       = openai_audio_transcription.example.text
}

output "audio_transcription_duration" {
  description = "The duration of the audio file in seconds"
  value       = openai_audio_transcription.example.duration
}

output "audio_translation_text" {
  description = "The translated text from audio translation"
  value       = openai_audio_translation.example.text
}

output "audio_translation_duration" {
  description = "The duration of the audio file in seconds"
  value       = openai_audio_translation.example.duration
}

output "text_to_speech_output" {
  description = "The path to the generated audio file"
  value       = openai_text_to_speech.example.output_file
}

# Data source outputs
output "data_speech_to_text_text" {
  description = "The text retrieved from the speech-to-text data source"
  value       = data.openai_speech_to_text.example_data.text
}

output "data_audio_transcription_text" {
  description = "The text retrieved from the audio transcription data source"
  value       = data.openai_audio_transcription.example_data.text
}

output "data_audio_translation_text" {
  description = "The text retrieved from the audio translation data source"
  value       = data.openai_audio_translation.example_data.text
}

output "data_text_to_speech_file" {
  description = "The file path retrieved from the text-to-speech data source"
  value       = data.openai_text_to_speech.example_data.file_path
} 