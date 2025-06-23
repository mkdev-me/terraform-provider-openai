# Example: Translating audio from various languages to English
# The Audio Translation API takes audio in any supported language and translates it to English

# Translate audio directly from file path
resource "openai_audio_translation" "translate_spanish" {
  # The audio file to translate (local file path)
  file = "./speech.mp3"

  # Model to use for translation - currently only "whisper-1" is available
  model = "whisper-1"

  # Optional: Provide a prompt to guide the translation style
  prompt = "Translate this Spanish podcast about technology into English."

  # Optional: Response format (json, text, srt, verbose_json, or vtt)
  # Default is "json" which includes the translated text
  response_format = "text"

  # Optional: Temperature for sampling (0 to 1)
  # Lower values make output more focused and deterministic
  temperature = 0.2
}

# Example with subtitle generation
resource "openai_audio_translation" "translate_with_subtitles" {
  file  = "./speech.mp3"
  model = "whisper-1"

  # Generate subtitles in SRT format
  response_format = "srt"

  # Guide the translation with context
  prompt = "This is a business presentation about quarterly results."
}

# Example with verbose JSON output
resource "openai_audio_translation" "translate_verbose" {
  file  = "./speech.mp3"
  model = "whisper-1"

  # Get detailed output with segments
  response_format = "verbose_json"
  temperature     = 0
}

# Output the translation result
output "translated_text" {
  value = openai_audio_translation.translate_spanish.text
}
