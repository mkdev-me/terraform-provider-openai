# Data sources for OpenAI audio processing

# Text-to-Speech data source for file verification
data "openai_text_to_speech" "tts_file" {
  count     = var.enable_text_to_speech && var.verify_tts_file ? 1 : 0
  file_path = var.text_to_speech_output_file

  # Only check for the file if text-to-speech was enabled
  depends_on = [
    openai_text_to_speech.this
  ]
}

# Speech-to-Text data source
data "openai_speech_to_text" "stt" {
  count            = var.enable_speech_to_text && var.use_stt_data_source ? 1 : 0
  transcription_id = length(openai_speech_to_text.this) > 0 ? openai_speech_to_text.this[0].id : var.existing_stt_id

  depends_on = [
    openai_speech_to_text.this
  ]
}

# Audio Transcription data source
data "openai_audio_transcription" "transcription" {
  count            = var.enable_audio_transcription && var.use_transcription_data_source ? 1 : 0
  transcription_id = length(openai_audio_transcription.this) > 0 ? openai_audio_transcription.this[0].id : var.existing_transcription_id

  depends_on = [
    openai_audio_transcription.this
  ]
}

# Audio Translation data source
data "openai_audio_translation" "translation" {
  count          = var.enable_audio_translation && var.use_translation_data_source ? 1 : 0
  translation_id = length(openai_audio_translation.this) > 0 ? openai_audio_translation.this[0].id : var.existing_translation_id

  depends_on = [
    openai_audio_translation.this
  ]
}

# NOTE: The following data sources are not currently supported by the OpenAI API
# and will result in error messages when used. They are commented out to prevent
# Terraform from attempting to use them.

# The OpenAI API does not currently provide endpoints to list audio resources.
# You can only retrieve individual resources using their specific data sources.

/*
# Get all Text-to-Speech data
data "openai_text_to_speechs" "all_tts" {
  count = var.retrieve_all_text_to_speech ? 1 : 0
  
  # Optional filtering
  model = var.filter_tts_by_model
  voice = var.filter_tts_by_voice
}

# Get all Speech-to-Text data
data "openai_speech_to_texts" "all_stt" {
  count = var.retrieve_all_speech_to_text ? 1 : 0
  
  # Optional filtering
  model = var.filter_stt_by_model
}

# Get all Audio Transcription data
data "openai_audio_transcriptions" "all_transcriptions" {
  count = var.retrieve_all_transcriptions ? 1 : 0
  
  # Optional filtering
  model = var.filter_transcriptions_by_model
}

# Get all Audio Translation data
data "openai_audio_translations" "all_translations" {
  count = var.retrieve_all_translations ? 1 : 0
  
  # Optional filtering
  model = var.filter_translations_by_model
}
*/ 