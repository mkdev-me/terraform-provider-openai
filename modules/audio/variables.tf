# Speech-to-Text Variables
variable "enable_speech_to_text" {
  description = "Whether to enable speech-to-text conversion"
  type        = bool
  default     = false
}

variable "speech_to_text_model" {
  description = "The model to use for speech-to-text conversion (e.g., 'whisper-1')"
  type        = string
  default     = "whisper-1"
}

variable "speech_to_text_file" {
  description = "Path to the audio file for speech-to-text conversion"
  type        = string
  default     = ""
}

variable "speech_to_text_language" {
  description = "The language of the input audio for speech-to-text (ISO-639-1 format)"
  type        = string
  default     = null
}

variable "speech_to_text_prompt" {
  description = "Optional text to guide the model's style or continue a previous audio segment"
  type        = string
  default     = null
}

variable "speech_to_text_response_format" {
  description = "The format of the transcript output (json, text, srt, verbose_json, vtt)"
  type        = string
  default     = "json"
}

variable "speech_to_text_temperature" {
  description = "The sampling temperature, between 0 and 1"
  type        = number
  default     = 0
}

# Audio Transcription Variables
variable "enable_audio_transcription" {
  description = "Whether to enable audio transcription"
  type        = bool
  default     = false
}

variable "audio_transcription_model" {
  description = "The model to use for audio transcription (e.g., 'whisper-1')"
  type        = string
  default     = "whisper-1"
}

variable "audio_transcription_file" {
  description = "Path to the audio file for transcription"
  type        = string
  default     = ""
}

variable "audio_transcription_language" {
  description = "The language of the input audio for transcription (ISO-639-1 format)"
  type        = string
  default     = null
}

variable "audio_transcription_prompt" {
  description = "Optional text to guide the model's style or continue a previous audio segment"
  type        = string
  default     = null
}

variable "audio_transcription_response_format" {
  description = "The format of the transcript output (json, text, srt, verbose_json, vtt)"
  type        = string
  default     = "json"
}

variable "audio_transcription_temperature" {
  description = "The sampling temperature, between 0 and 1"
  type        = number
  default     = 0
}

# Audio Translation Variables
variable "enable_audio_translation" {
  description = "Whether to enable audio translation"
  type        = bool
  default     = false
}

variable "audio_translation_model" {
  description = "The model to use for audio translation (e.g., 'whisper-1')"
  type        = string
  default     = "whisper-1"
}

variable "audio_translation_file" {
  description = "Path to the audio file for translation"
  type        = string
  default     = ""
}

variable "audio_translation_prompt" {
  description = "Optional text to guide the model's style or continue a previous audio segment"
  type        = string
  default     = null
}

variable "audio_translation_response_format" {
  description = "The format of the translation output (json, text, srt, verbose_json, vtt)"
  type        = string
  default     = "json"
}

variable "audio_translation_temperature" {
  description = "The sampling temperature, between 0 and 1"
  type        = number
  default     = 0
}

# Text-to-Speech Variables
variable "enable_text_to_speech" {
  description = "Whether to enable text-to-speech conversion"
  type        = bool
  default     = false
}

variable "text_to_speech_model" {
  description = "The model to use for text-to-speech (e.g., 'tts-1', 'tts-1-hd')"
  type        = string
  default     = "tts-1"
}

variable "text_to_speech_input" {
  description = "The text to convert to speech"
  type        = string
  default     = ""
}

variable "text_to_speech_voice" {
  description = "The voice to use for speech (e.g., 'alloy', 'echo', 'fable', 'onyx', 'nova', 'shimmer')"
  type        = string
  default     = "alloy"
}

variable "text_to_speech_response_format" {
  description = "The format of the audio output (mp3, opus, aac, flac)"
  type        = string
  default     = "mp3"
}

variable "text_to_speech_speed" {
  description = "The speed of the audio output, between 0.25 and 4.0"
  type        = number
  default     = 1.0
}

variable "text_to_speech_output_file" {
  description = "The path where the generated audio file will be saved"
  type        = string
  default     = ""
}

# Data source variables

# Text-to-Speech data source variables
variable "verify_tts_file" {
  description = "Whether to verify the text-to-speech output file using the data source"
  type        = bool
  default     = false
}

# Speech-to-Text data source variables
variable "use_stt_data_source" {
  description = "Whether to use the speech-to-text data source"
  type        = bool
  default     = false
}

variable "existing_stt_id" {
  description = "ID of an existing speech-to-text transcription to use with the data source"
  type        = string
  default     = ""
}

# Audio Transcription data source variables
variable "use_transcription_data_source" {
  description = "Whether to use the audio transcription data source"
  type        = bool
  default     = false
}

variable "existing_transcription_id" {
  description = "ID of an existing audio transcription to use with the data source"
  type        = string
  default     = ""
}

# Audio Translation data source variables
variable "use_translation_data_source" {
  description = "Whether to use the audio translation data source"
  type        = bool
  default     = false
}

variable "existing_translation_id" {
  description = "ID of an existing audio translation to use with the data source"
  type        = string
  default     = ""
}

# Multi-item data source variables

# Text-to-Speech list data source variables
variable "retrieve_all_text_to_speech" {
  description = "Whether to retrieve all text-to-speech conversions"
  type        = bool
  default     = false
}

variable "filter_tts_by_model" {
  description = "Filter text-to-speech conversions by model (e.g., 'tts-1', 'tts-1-hd')"
  type        = string
  default     = null
}

variable "filter_tts_by_voice" {
  description = "Filter text-to-speech conversions by voice (e.g., 'alloy', 'echo', 'fable', 'onyx', 'nova', 'shimmer')"
  type        = string
  default     = null
}

# Speech-to-Text list data source variables
variable "retrieve_all_speech_to_text" {
  description = "Whether to retrieve all speech-to-text conversions"
  type        = bool
  default     = false
}

variable "filter_stt_by_model" {
  description = "Filter speech-to-text conversions by model (e.g., 'whisper-1')"
  type        = string
  default     = null
}

# Audio Transcription list data source variables
variable "retrieve_all_transcriptions" {
  description = "Whether to retrieve all audio transcriptions"
  type        = bool
  default     = false
}

variable "filter_transcriptions_by_model" {
  description = "Filter audio transcriptions by model (e.g., 'whisper-1')"
  type        = string
  default     = null
}

# Audio Translation list data source variables
variable "retrieve_all_translations" {
  description = "Whether to retrieve all audio translations"
  type        = bool
  default     = false
}

variable "filter_translations_by_model" {
  description = "Filter audio translations by model (e.g., 'whisper-1')"
  type        = string
  default     = null
} 