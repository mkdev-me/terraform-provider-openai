resource "openai_text_to_speech" "example" {
  model       = "tts-1"
  input       = "Hello world! This is a test of the OpenAI text-to-speech API."
  voice       = "alloy"
  output_path = "output/speech.mp3"

  # Optional parameters
  response_format = "mp3"
  speed           = 1.0
}

output "audio_file_path" {
  value = openai_text_to_speech.example.output_path
}

