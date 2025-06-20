# Module outputs for OpenAI audio processing resources and data sources

# Speech-to-Text resource outputs
output "speech_to_text_text" {
  description = "The transcribed text from speech-to-text conversion"
  value       = var.enable_speech_to_text ? openai_speech_to_text.this[0].text : null
}

output "speech_to_text_created_at" {
  description = "The timestamp when the speech-to-text transcription was generated"
  value       = var.enable_speech_to_text ? openai_speech_to_text.this[0].created_at : null
}

output "speech_to_text_id" {
  description = "The ID of the speech-to-text resource"
  value       = var.enable_speech_to_text ? openai_speech_to_text.this[0].id : null
}

# Audio Transcription resource outputs
output "audio_transcription_text" {
  description = "The transcribed text from audio transcription"
  value       = var.enable_audio_transcription ? openai_audio_transcription.this[0].text : null
}

output "audio_transcription_duration" {
  description = "The duration of the audio file in seconds for audio transcription"
  value       = var.enable_audio_transcription ? openai_audio_transcription.this[0].duration : null
}

output "audio_transcription_id" {
  description = "The ID of the audio transcription resource"
  value       = var.enable_audio_transcription ? openai_audio_transcription.this[0].id : null
}

# Audio Translation resource outputs
output "audio_translation_text" {
  description = "The translated text from audio translation"
  value       = var.enable_audio_translation ? openai_audio_translation.this[0].text : null
}

output "audio_translation_duration" {
  description = "The duration of the audio file in seconds for audio translation"
  value       = var.enable_audio_translation ? openai_audio_translation.this[0].duration : null
}

output "audio_translation_id" {
  description = "The ID of the audio translation resource"
  value       = var.enable_audio_translation ? openai_audio_translation.this[0].id : null
}

# Text-to-Speech resource outputs
output "text_to_speech_created_at" {
  description = "The timestamp when the text-to-speech conversion was generated"
  value       = var.enable_text_to_speech ? openai_text_to_speech.this[0].created_at : null
}

output "text_to_speech_output_file" {
  description = "The path to the generated audio file"
  value       = var.enable_text_to_speech ? var.text_to_speech_output_file : null
}

# Data source outputs

# Text-to-Speech data source outputs
output "tts_file_exists" {
  description = "Whether the text-to-speech file exists"
  value       = var.enable_text_to_speech && var.verify_tts_file ? data.openai_text_to_speech.tts_file[0].exists : null
}

output "tts_file_size" {
  description = "Size of the text-to-speech file in bytes"
  value       = var.enable_text_to_speech && var.verify_tts_file ? data.openai_text_to_speech.tts_file[0].file_size_bytes : null
}

output "tts_file_last_modified" {
  description = "Last modified timestamp of the text-to-speech file"
  value       = var.enable_text_to_speech && var.verify_tts_file ? data.openai_text_to_speech.tts_file[0].last_modified : null
}

# Speech-to-Text data source outputs
output "stt_data_source_id" {
  description = "ID of the speech-to-text data source"
  value       = var.enable_speech_to_text && var.use_stt_data_source ? data.openai_speech_to_text.stt[0].id : null
}

# Audio Transcription data source outputs
output "transcription_data_source_id" {
  description = "ID of the audio transcription data source"
  value       = var.enable_audio_transcription && var.use_transcription_data_source ? data.openai_audio_transcription.transcription[0].id : null
}

# Audio Translation data source outputs
output "translation_data_source_id" {
  description = "ID of the audio translation data source"
  value       = var.enable_audio_translation && var.use_translation_data_source ? data.openai_audio_translation.translation[0].id : null
}

# NOTE: The following outputs are not currently functional because the OpenAI API 
# does not support listing audio resources. These are commented out to prevent errors.

/*
# Multi-item data source outputs

# Text-to-Speech list data source outputs
output "all_text_to_speechs" {
  description = "List of all text-to-speech conversions"
  value       = var.retrieve_all_text_to_speech ? data.openai_text_to_speechs.all_tts[0].text_to_speechs : null
  sensitive   = true
}

output "text_to_speechs_count" {
  description = "Count of retrieved text-to-speech conversions"
  value       = var.retrieve_all_text_to_speech ? length(data.openai_text_to_speechs.all_tts[0].text_to_speechs) : 0
}

# Speech-to-Text list data source outputs
output "all_speech_to_texts" {
  description = "List of all speech-to-text conversions"
  value       = var.retrieve_all_speech_to_text ? data.openai_speech_to_texts.all_stt[0].speech_to_texts : null
  sensitive   = true
}

output "speech_to_texts_count" {
  description = "Count of retrieved speech-to-text conversions"
  value       = var.retrieve_all_speech_to_text ? length(data.openai_speech_to_texts.all_stt[0].speech_to_texts) : 0
}

# Audio Transcription list data source outputs
output "all_audio_transcriptions" {
  description = "List of all audio transcriptions"
  value       = var.retrieve_all_transcriptions ? data.openai_audio_transcriptions.all_transcriptions[0].transcriptions : null
  sensitive   = true
}

output "audio_transcriptions_count" {
  description = "Count of retrieved audio transcriptions"
  value       = var.retrieve_all_transcriptions ? length(data.openai_audio_transcriptions.all_transcriptions[0].transcriptions) : 0
}

# Audio Translation list data source outputs
output "all_audio_translations" {
  description = "List of all audio translations"
  value       = var.retrieve_all_translations ? data.openai_audio_translations.all_translations[0].translations : null
  sensitive   = true
}

output "audio_translations_count" {
  description = "Count of retrieved audio translations"
  value       = var.retrieve_all_translations ? length(data.openai_audio_translations.all_translations[0].translations) : 0
}
*/ 