# OpenAI Audio Module

terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
    }
  }
}

# Text-to-Speech Resource (must be created first)
resource "openai_text_to_speech" "this" {
  count           = var.enable_text_to_speech ? 1 : 0
  model           = var.text_to_speech_model
  input           = var.text_to_speech_input
  voice           = var.text_to_speech_voice
  response_format = var.text_to_speech_response_format
  speed           = var.text_to_speech_speed
  output_file     = var.text_to_speech_output_file
}

# Speech-to-Text Resource
resource "openai_speech_to_text" "this" {
  count           = var.enable_speech_to_text ? 1 : 0
  model           = var.speech_to_text_model
  file            = var.speech_to_text_file
  language        = var.speech_to_text_language
  prompt          = var.speech_to_text_prompt
  response_format = var.speech_to_text_response_format
  temperature     = var.speech_to_text_temperature

  depends_on = [openai_text_to_speech.this]
}

# Audio Transcription Resource
resource "openai_audio_transcription" "this" {
  count           = var.enable_audio_transcription ? 1 : 0
  model           = var.audio_transcription_model
  file            = var.audio_transcription_file
  language        = var.audio_transcription_language
  prompt          = var.audio_transcription_prompt
  response_format = var.audio_transcription_response_format
  temperature     = var.audio_transcription_temperature

  depends_on = [openai_text_to_speech.this]
}

# Audio Translation Resource
resource "openai_audio_translation" "this" {
  count           = var.enable_audio_translation ? 1 : 0
  model           = var.audio_translation_model
  file            = var.audio_translation_file
  prompt          = var.audio_translation_prompt
  response_format = var.audio_translation_response_format
  temperature     = var.audio_translation_temperature

  depends_on = [openai_text_to_speech.this]
}
