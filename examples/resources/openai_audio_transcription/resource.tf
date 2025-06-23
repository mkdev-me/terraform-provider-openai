resource "openai_audio_transcription" "example" {
  file  = "./speech.mp3"
  model = "whisper-1"

  # Optional parameters
  language        = "en"
  prompt          = "This is a transcription of a meeting"
  response_format = "json"
  temperature     = 0
}

# Example with verbose output
resource "openai_audio_transcription" "verbose_example" {
  file  = "./speech.mp3"
  model = "whisper-1"

  response_format = "verbose_json"
  temperature     = 0.2

  # Get word-level timestamps (only works with verbose_json)
  timestamp_granularities = ["word", "segment"]
}

output "transcription_text" {
  value = openai_audio_transcription.example.text
}

