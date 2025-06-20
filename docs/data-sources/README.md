# OpenAI Provider Data Sources

This directory contains documentation for all data sources in the OpenAI Terraform Provider. Data sources allow you to fetch information about existing resources from the OpenAI API.

## General Usage

Data sources follow this general pattern:

```hcl
data "openai_resource_name" "example" {
  # Required parameters
  id = "resource-id"
  
  # Optional parameters
  project_id = "project-id"
}

output "resource_info" {
  value = data.openai_resource_name.example.attribute
}
```

### Alternative Approaches

Instead of listing operations, use the individual resource data sources to retrieve specific resources:

```hcl
# Retrieve a specific audio transcription by ID
data "openai_audio_transcription" "example" {
  transcription_id = "transcription-1234567890"
}

# Retrieve a specific audio translation by ID
data "openai_audio_translation" "example" {
  translation_id = "translation-1234567890"
}

# Retrieve a specific speech-to-text conversion by ID
data "openai_speech_to_text" "example" {
  transcription_id = "transcription-1234567890"
}

# Retrieve a specific text-to-speech file
data "openai_text_to_speech" "example" {
  file_path = "path/to/audio/file.mp3"
}
```

If you need to keep track of multiple audio resources, consider implementing your own tracking system or database. 

