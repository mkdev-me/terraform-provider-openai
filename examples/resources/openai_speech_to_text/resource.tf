# Example: Converting speech to text using OpenAI's Whisper model
# The Speech-to-Text API transcribes audio into text in the original language

# Basic transcription
resource "openai_speech_to_text" "podcast_transcript" {
  # The audio file to transcribe
  file = "speech.mp3"

  # Model to use - currently only "whisper-1" is available
  model = "whisper-1"

  # Optional: Language of the audio (ISO-639-1 format)
  # If not specified, Whisper will auto-detect
  language = "en"

  # Optional: Response format (json, text, srt, verbose_json, or vtt)
  # Default is "json"
  response_format = "text"

  # Optional: Temperature for sampling (0 to 1)
  # Lower values are more deterministic
  temperature = 0.0
}

# Transcription with timestamps (SRT format for subtitles)
resource "openai_speech_to_text" "podcast_subtitles" {
  file  = "speech.mp3"
  model = "whisper-1"

  # SRT format includes timestamps
  response_format = "srt"

  # Provide context to improve accuracy
  prompt = "This is a tech podcast discussing AI and machine learning trends."
}

# Verbose transcription with word-level timestamps
resource "openai_speech_to_text" "detailed_transcript" {
  file  = "speech.mp3"
  model = "whisper-1"

  # Verbose JSON includes word-level timestamps and confidence scores
  response_format = "verbose_json"

  # Help the model with proper nouns and technical terms
  prompt = "Meeting participants: John Smith, Sarah Johnson. Topics: Q4 revenue, Project Alpha, TensorFlow implementation."

  # Auto-detect language
  # language = null
}

# Transcription for multi-language content
resource "openai_speech_to_text" "spanish_transcript" {
  file     = "speech.mp3"
  model    = "whisper-1"
  language = "es" # Spanish

  # Get clean text output
  response_format = "text"

  # Context in the target language
  prompt = "Entrevista sobre tecnología e innovación en América Latina."
}

# WebVTT format for web video players
resource "openai_speech_to_text" "video_captions" {
  file  = "speech.mp3"
  model = "whisper-1"

  # VTT format for HTML5 video
  response_format = "vtt"

  temperature = 0.2 # Slightly higher for natural flow
}

# Output the transcription text
output "podcast_text" {
  value = openai_speech_to_text.podcast_transcript.text
}
