# OpenAI Terraform Provider Modules

This directory contains reusable Terraform modules for managing OpenAI resources through the Terraform provider.

## Available Modules

| Module | Description |
|--------|-------------|
| [audio](./audio) | Convert text to speech with OpenAI TTS models |
| [batch](./batch) | Process large-scale data with OpenAI batch processing endpoints |
| [chat_completion](./chat_completion) | Generate conversational responses with OpenAI Chat models |
| [embeddings](./embeddings) | Create vector embeddings of text for semantic search |
| [files](./files) | Upload and manage files for use with various OpenAI endpoints |
| [fine_tuning](./fine_tuning) | Create and manage fine-tuned models |
| [image](./image) | Generate, edit, and create variations of images |
| [invite](./invite) | Manage user invitations to OpenAI organizations |
| [model_response](./model_response) | Process text with various OpenAI models |
| [moderation](./moderation) | Check content for policy compliance |
| [organization_users](./organization_users) | Retrieve and manage organization users |
| [project_api](./project_api) | Retrieve and manage project API keys |
| [project_user](./project_user) | Manage users in OpenAI projects |
| [projects](./projects) | Create and manage OpenAI projects |
| [rate_limit](./rate_limit) | Configure rate limits for OpenAI projects |
| [service_account](./service_account) | Manage service accounts for OpenAI projects |
| [system_api](./system_api) | Create and manage OpenAI Admin API Keys through the Terraform provider |
| [upload](./upload) | Upload files for fine-tuning and other operations |
| [vector_store](./vector_store) | Create and manage vector stores for RAG applications |
| [model](model) | Retrieve information about a specific OpenAI model |

## System API Key Module

The `system_api` module creates OpenAI Admin API keys through the native Terraform provider resources. It provides a standardized interface for creating, retrieving and managing OpenAI system-level API keys with specific permissions and expiration settings.

### Features

- Creates OpenAI Admin API keys with customizable settings
- Supports permanent keys or temporary keys with expiration dates
- Restricts key permissions with specific scopes for granular access control
- Returns key values as Terraform outputs (with sensitive flag for security)

### Requirements

- OpenAI account with administrator privileges
- Admin API key with `api.management.write` permissions
- Terraform 0.13+

### Usage

```hcl
module "admin_key" {
  source = "../../modules/system_api"
  
  name       = "terraform-admin-key"
  expires_at = 1772382952  # Optional: Unix timestamp for expiration
  scopes     = ["api.management.read"]  # Optional: Permission scopes
}

output "admin_key_id" {
  value = module.admin_key.key_id
}

output "admin_key_value" {
  value     = module.admin_key.key_value
  sensitive = true
}
```

### Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| name | The name of the system API key | string | n/a | yes |
| expires_at | Unix timestamp for when the API key expires | number | null | no |
| scopes | The scopes this key is restricted to | list(string) | [] | no |

### Outputs

| Name | Description | Type |
|------|-------------|------|
| key_id | The ID of the API key (format: key_XXXX) | string |
| key_value | The value of the API key (sensitive) | string |
| created_at | Creation timestamp | string |
| name | The name of the API key | string |
| expires_at | Expiration date of the API key | number |

### Examples

See the [examples/system_api](../examples/system_api) directory for complete examples including:

- Creating admin keys with and without expiration
- Importing existing admin keys into Terraform state
- Deleting admin keys

## Module: project_api

The `project_api` module allows importing and managing existing OpenAI Project API Keys. 

**IMPORTANT:** Unlike admin keys, project keys **cannot be created via the API** and must be imported after manual creation in the OpenAI dashboard.

### Features

- Import existing project API keys
- Delete project API keys
- Associate keys with Terraform resources for tracking and management

### Usage

```hcl
module "project_key" {
  source = "../../modules/project_api"
  
  project_id = "proj_abc123"   # Your actual project ID
  name       = "development-key"  # Should match your key name in the dashboard
}
```

### Import Process

1. Create a project API key manually in the OpenAI dashboard
   (https://platform.openai.com/api-keys)
2. Run `terraform apply` to create a placeholder in Terraform state
3. Import your existing key with this command:

```bash
terraform import module.project_key.openai_project_api_key.this "PROJECT_ID:API_KEY_ID"
```

For example:
```bash
terraform import module.project_key.openai_project_api_key.this "proj_abc123:key_def456"
```

### Known Limitations

The OpenAI API does not allow programmatic creation of project API keys. Any attempt to create a new key through Terraform will result in mock data being generated. Always use the import workflow described above.

## Module: projects

The `projects` module creates and manages OpenAI Projects, providing an organizational structure for models, files, and API keys.

### Features

- Create projects with custom names and descriptions
- Set default projects
- Configure project attributes
- Create multiple projects from a single configuration

### Usage

```hcl
module "development_project" {
  source = "../../modules/projects"
  
  name        = "Development Environment"
  description = "Project for development and testing workloads"
  is_default  = false
}
```

## Module: project_user

The `project_user` module manages users within OpenAI projects, allowing you to add users and assign roles.

### Features

- Add existing users to projects with specific roles (owner or member)
- Update user roles within projects
- Manage user permissions across multiple projects
- Retrieve user information such as email and add date

### Important Technical Details

- Users must be invited via the OpenAI dashboard first
- The API uses `POST` method for updating user roles, not `PATCH`
- User IDs must be discovered using the API (see examples)
- Proper dependency management is required

### Usage

```hcl
module "project_user" {
  source     = "../../modules/project_user"
  project_id = openai_project.example.id
  user_id    = "user-abc123"  # Must be a real user ID from your organization
  role       = "member"       # Can be "owner" or "member"
  
  depends_on = [openai_project.example]
}
```

### Finding User IDs

To use this module, you need to find the user's ID first. This can be done with:

```bash
# List all users in the organization
curl https://api.openai.com/v1/organization/users \
  -H "Authorization: Bearer $OPENAI_ADMIN_KEY" \
  -H "Content-Type: application/json"
```

See the [examples/project_user](../examples/project_user) directory for complete examples and workflow.

## Module: invite

The `invite` module allows you to send invitations to users to join your OpenAI organization.

### Features

- Send invitations to users to join your OpenAI organization
- Assign organization-level roles (owner or reader) to invited users
- Track invitation status and expiration

### Important Note

The OpenAI API does not support adding users to projects through invitations. To add users to projects, you should follow this 3-step process:

1. Invite the user to your organization
2. Wait for the user to accept the invitation
3. Add the user to projects using the `project_user` module

### Usage

```hcl
# Step 1: Invite user to the organization
module "invite_user" {
  source = "../../modules/invite"
  
  email = "user@example.com"
  role  = "reader"  # Can be "owner" or "reader"
}

# Step 2: Get user ID after acceptance (using data source)
data "openai_organization_user" "new_user" {
  email = "user@example.com"
}

# Step 3: Add user to project
module "project_user" {
  source     = "../../modules/project_user"
  project_id = "proj_abc123"
  user_id    = data.openai_organization_user.new_user.id
  role       = "member"  # Can be "owner" or "member"
}
```

### Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| email | The email address of the user to invite | string | n/a | yes |
| role | The role to assign to the invited user (owner or reader) | string | n/a | yes |
| api_key | Custom API key to use for this module | string | null | no |

### Outputs

| Name | Description |
|------|-------------|
| id | The unique identifier for the invitation |
| invite_id | The ID of the invitation |
| email | The email address of the invited user |
| role | The role assigned to the invited user |

## Invitation Workflow Module

The `invite` module provides a reusable implementation of the proper OpenAI invitation workflow.

### Known Issues with OpenAI Invitations

When working with OpenAI invitations through the API, several challenges exist:

1. **Project Assignments in Invitations Don't Work**: When a user accepts an invitation, any project assignments included in the invitation aren't actually applied by the OpenAI API, even though they're accepted in the API request.

2. **Accepted Invitations Can't Be Deleted**: The OpenAI API returns an error when attempting to delete an invitation that has already been accepted.

3. **User ID Required for Project Assignment**: After invitation acceptance, you need to perform a lookup to get the user's ID before you can add them to projects.

### How the Module Solves These Issues

The `invite` module implements a proper workflow that:

1. Sends invitations to the organization only (not relying on project assignments)
2. Uses the `organization_users` data source to check for acceptance and retrieve user IDs
3. Explicitly adds users to projects after invitation acceptance
4. Handles the "already accepted" error during invitation deletion

See the [invite module documentation](invite/README.md) for detailed usage instructions.

## Module: audio

The `audio` module provides a comprehensive interface for working with OpenAI's audio processing capabilities, including speech-to-text, audio transcription, audio translation, and text-to-speech.

### Features

- Speech-to-Text: Convert speech to text using OpenAI's models (gpt-4o-transcribe, gpt-4o-mini-transcribe, whisper-1)
- Audio Transcription: Detailed transcription of audio files with timestamps and metadata
- Audio Translation: Translate audio from any language to English
- Text-to-Speech: Convert text to spoken audio with multiple voice options
- Configurable parameters for each resource
- Conditional resource creation using enable flags
- Comprehensive outputs for all operations
- Support for advanced features like logprobs, streaming, and timestamp granularities

### Requirements

- OpenAI API key with appropriate permissions
- Audio files in supported formats for input (FLAC, MP3, MP4, MPEG, MPGA, M4A, OGG, WAV, WEBM)
- Terraform 0.13+

### Usage

```hcl
module "audio_processing" {
  source = "../../modules/audio"

  # Speech-to-Text configuration
  enable_speech_to_text = true
  speech_to_text_model = "gpt-4o-transcribe"
  speech_to_text_file = "/path/to/audio/file.mp3"
  speech_to_text_language = "en"
  speech_to_text_prompt = "Transcript of a meeting"
  speech_to_text_response_format = "json"
  speech_to_text_temperature = 0.2
  # speech_to_text_include = ["logprobs"]
  # speech_to_text_timestamp_granularities = ["segment"]

  # Audio Transcription configuration
  enable_audio_transcription = true
  audio_transcription_model = "whisper-1"
  audio_transcription_file = "/path/to/audio/file.mp3"
  audio_transcription_language = "en"
  audio_transcription_prompt = "Transcript of a meeting"
  audio_transcription_response_format = "json"
  audio_transcription_temperature = 0.2

  # Audio Translation configuration
  enable_audio_translation = true
  audio_translation_model = "whisper-1"
  audio_translation_file = "/path/to/audio/file.mp3"
  audio_translation_prompt = "Translation of a conversation"
  audio_translation_response_format = "json"
  audio_translation_temperature = 0.2
  
  # Text-to-Speech configuration
  enable_text_to_speech = true
  text_to_speech_model = "tts-1"
  text_to_speech_input = "Text to convert to speech"
  text_to_speech_voice = "alloy"
  text_to_speech_response_format = "mp3"
  text_to_speech_speed = 1.0
  text_to_speech_output_file = "/path/to/output/speech.mp3"
  # text_to_speech_instructions = "Speak in a calm and clear voice" # Only with gpt-4o-mini-tts
}
```

### Inputs

#### Speech-to-Text Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| enable_speech_to_text | Whether to create the speech-to-text resource | bool | false | no |
| speech_to_text_model | The model to use for speech-to-text | string | "whisper-1" | no |
| speech_to_text_file | Path to the audio file to transcribe | string | n/a | yes |
| speech_to_text_language | Language of the audio (ISO-639-1 format) | string | null | no |
| speech_to_text_prompt | Optional prompt to guide transcription | string | null | no |
| speech_to_text_response_format | Format of the response | string | "json" | no |
| speech_to_text_temperature | Temperature for the model (0.0 to 1.0) | number | 0.2 | no |
| speech_to_text_include | Additional information to include (e.g., logprobs) | list(string) | null | no |
| speech_to_text_stream | Whether to stream the response | bool | false | no |
| speech_to_text_timestamp_granularities | Timestamp granularities to include | list(string) | null | no |

#### Audio Transcription Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| enable_audio_transcription | Whether to create the audio transcription resource | bool | false | no |
| audio_transcription_model | The model to use for transcription | string | "whisper-1" | no |
| audio_transcription_file | Path to the audio file to transcribe | string | n/a | yes |
| audio_transcription_language | Language of the audio (ISO-639-1 format) | string | null | no |
| audio_transcription_prompt | Optional prompt to guide transcription | string | null | no |
| audio_transcription_response_format | Format of the response | string | "json" | no |
| audio_transcription_temperature | Temperature for the model (0.0 to 1.0) | number | 0.2 | no |
| audio_transcription_include | Additional information to include (e.g., logprobs) | list(string) | null | no |
| audio_transcription_stream | Whether to stream the response | bool | false | no |
| audio_transcription_timestamp_granularities | Timestamp granularities to include | list(string) | null | no |

#### Audio Translation Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| enable_audio_translation | Whether to create the audio translation resource | bool | false | no |
| audio_translation_model | The model to use for translation | string | "whisper-1" | no |
| audio_translation_file | Path to the audio file to translate | string | n/a | yes |
| audio_translation_prompt | Optional prompt to guide translation (in English) | string | null | no |
| audio_translation_response_format | Format of the response | string | "json" | no |
| audio_translation_temperature | Temperature for the model (0.0 to 1.0) | number | 0.2 | no |

#### Text-to-Speech Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| enable_text_to_speech | Whether to create the text-to-speech resource | bool | false | no |
| text_to_speech_model | The model to use for text-to-speech | string | "tts-1" | no |
| text_to_speech_input | Text to convert to speech (max 4096 characters) | string | n/a | yes |
| text_to_speech_voice | Voice to use | string | "alloy" | no |
| text_to_speech_response_format | Audio format of the response | string | "mp3" | no |
| text_to_speech_speed | Speed of the speech (0.25 to 4.0) | number | 1.0 | no |
| text_to_speech_output_file | Path to save the output audio | string | n/a | yes |
| text_to_speech_instructions | Instructions to guide the speech (only with gpt-4o-mini-tts) | string | null | no |

### Outputs

#### Speech-to-Text Outputs

| Name | Description | Type |
|------|-------------|------|
| speech_to_text_text | The transcribed text | string |
| speech_to_text_created_at | Timestamp of transcription | string |

#### Audio Transcription Outputs

| Name | Description | Type |
|------|-------------|------|
| audio_transcription_text | The transcribed text | string |
| audio_transcription_duration | Duration of the audio in seconds | number |

#### Audio Translation Outputs

| Name | Description | Type |
|------|-------------|------|
| audio_translation_text | The translated text | string |
| audio_translation_duration | Duration of the audio in seconds | number |

#### Text-to-Speech Outputs

| Name | Description | Type |
|------|-------------|------|
| text_to_speech_created_at | Timestamp of generation | string |
| text_to_speech_output_file | Path to the generated audio file | string |

### Supported Response Formats

Speech-to-Text and Audio Transcription/Translation support these formats:
- `json`: JSON object with transcription/translation data (only format for gpt-4o models)
- `text`: Plain text output
- `srt`: SubRip subtitle format
- `verbose_json`: Detailed JSON with additional metadata
- `vtt`: Web Video Text Tracks format

Text-to-Speech supports these formats:
- `mp3`: MP3 audio format
- `opus`: Opus audio format
- `aac`: AAC audio format
- `flac`: FLAC audio format
- `wav`: WAV audio format
- `pcm`: PCM audio format

### Examples

See the [examples/audio](../examples/audio) directory for complete examples including:
- Basic speech-to-text conversion with different models
- Detailed audio transcription
- Audio translation from various languages
- Text-to-speech with different voices and formats
- Using different response formats
- Working with multiple audio files

## Module: batch

The `batch` module allows creating and managing batch processing jobs in OpenAI for efficiently processing large volumes of data.

### Features

- Creation of batch jobs for different endpoints (embeddings, chat completions, etc.)
- Management of input and output files
- Monitoring of batch job status

### Usage

```hcl
module "embedding_batch" {
  source           = "../../modules/batch"
  input_file_id    = openai_file.my_file.id
  endpoint         = "/v1/embeddings"
  model            = "text-embedding-ada-002"
  completion_window = "24h"
}
```

### Inputs

| Name | Description | Type | Required | Default Value |
|--------|-------------|------|-----------|-------------------|
| input_file_id | ID of the input file uploaded to OpenAI | `string` | Yes | - |
| endpoint | API endpoint to use (e.g., "/v1/embeddings") | `string` | Yes | - |
| model | ID of the model to use | `string` | Yes | - |
| completion_window | Time window to complete the job | `string` | No | "24h" |
| project_id | OpenAI project ID (optional) | `string` | No | "" |
| metadata | Metadata to attach to the job | `map(string)` | No | {} |

### Outputs

| Name | Description |
|--------|-------------|
| batch_id | ID of the created batch job |
| batch_status | Current status of the batch job |
| output_file_id | ID of the output file with the results |
| created_at | Timestamp when the job was created |
| expires_at | Timestamp when the job expires |
| error | Error details if the job failed |

### Supported Endpoints

The batch module supports various endpoints including:
- `/v1/embeddings`: For generating embeddings for multiple texts
- `/v1/chat/completions`: For generating chat completions for multiple requests
- `/v1/completions`: For generating completions for multiple prompts
- `/v1/responses`: For accessing already generated responses

See the [module documentation](./batch/README.md) for complete information.

## Module: embeddings

The `embeddings` module provides a simple way to generate embeddings using the OpenAI API for tasks such as search, clustering, recommendations, and other natural language processing applications.

### Features

- Generate vector representations of text that capture semantic meaning
- Support for various embedding models
- Control over embedding dimensions
- Multiple text inputs in a single request

### Usage

```hcl
module "text_embedding" {
  source = "../../modules/embeddings"
  
  input = "This is an example text to generate an embedding"
  model = "text-embedding-ada-002"
}

output "embedding_result" {
  value = module.text_embedding.embeddings
  sensitive = true
}
```

### Inputs

| Name | Description | Type | Default | Required |
|--------|-------------|------|------------------|-----------|
| `model` | ID of the model to use (e.g., text-embedding-ada-002) | `string` | `"text-embedding-ada-002"` | No |
| `input` | Text to generate the embedding for. Can be a string or a JSON array of strings | `string` | - | Yes |
| `user` | Optional unique identifier representing the end user | `string` | `null` | No |
| `encoding_format` | Format to return the embeddings: 'float' or 'base64' | `string` | `"float"` | No |
| `dimensions` | Number of dimensions for the embeddings (only for text-embedding-3 and later) | `number` | `null` | No |
| `project_id` | OpenAI project ID to use for this request | `string` | `null` | No |

### Outputs

| Name | Description |
|--------|-------------|
| `embeddings` | The generated embeddings |
| `model_used` | The model used to generate the embeddings |
| `usage` | Token usage statistics for the request |
| `embedding_id` | The ID of the generated embedding |

See the [module documentation](./embeddings/README.md) for complete information.

## Module: fine_tuning

The `fine_tuning` module provides a streamlined interface for creating and managing OpenAI fine-tuning jobs, allowing you to easily create customized models tailored to your specific use cases.

### Features

- Create fine-tuning jobs for OpenAI models
- Support for all supported OpenAI fine-tuning parameters
- Customizable hyperparameters for training
- Optional completion waiting during creation
- Detailed job status and model outputs

### Usage

```hcl
module "fine_tuned_model" {
  source        = "../../modules/fine_tuning"
  
  model         = "gpt-3.5-turbo"  # Currently supported model for fine-tuning
  training_file = "file-abc123"  # ID of your training data file
}

output "model_id" {
  value = module.fine_tuned_model.fine_tuned_model_id
}
```

For advanced usage with hyperparameters and further options, see the [module documentation](./fine_tuning/README.md).

## Module: chat_completion

The `chat_completion` module provides a wrapper around the OpenAI Chat Completions API, allowing you to generate conversational text using OpenAI's state-of-the-art language models.

### Features

- Generate conversational responses using OpenAI's powerful models (GPT-4, GPT-3.5-turbo, etc.)
- Easily configure message sequences with different roles (system, user, assistant)
- Fine-tune generation parameters (temperature, max tokens, etc.)
- Support for function calling
- Proper error handling and state management

### Usage

```hcl
module "chat" {
  source = "path/to/modules/chat_completion"
  
  model = "gpt-3.5-turbo"
  
  messages = [
    {
      role    = "system"
      content = "You are a helpful assistant."
    },
    {
      role    = "user"
      content = "Tell me about the solar system."
    }
  ]
  
  temperature = 0.7
  max_tokens  = 500
}

output "assistant_response" {
  value = module.chat.content
}
```

For function calling examples and detailed parameter information, see the [module documentation](./chat_completion/README.md).

## Module: image

The `image` module provides a unified interface for working with OpenAI's DALL-E image generation capabilities, including image generation, editing, and creating variations.

### Features

- Supports all three OpenAI image operations:
  - **Image Generation**: Create images from textual descriptions
  - **Image Editing**: Edit existing images with a text prompt and optional mask
  - **Image Variation**: Create variations of existing images
- Configurable outputs in URL or base64 format
- Support for DALL-E 2 and DALL-E 3 models
- Advanced controls for image quality, style, and size

### Usage

```hcl
module "cat_image" {
  source    = "../../modules/image"
  
  operation = "generation"
  prompt    = "A photorealistic cat wearing a space suit on the moon"
  model     = "dall-e-3"  # Optional, defaults to "dall-e-3" for generation
  size      = "1024x1024"
  quality   = "hd"        # Optional, only for dall-e-3
  style     = "vivid"     # Optional, only for dall-e-3
}

output "image_url" {
  value = module.cat_image.image_urls[0]
}
```

For image editing, variations, and detailed parameter information, see the [module documentation](./image/README.md).

## Module: files

The `files` module provides a simplified interface for uploading and managing files with the OpenAI API. Files are a fundamental component for many OpenAI features including fine-tuning, assistants, and batch processing.

### Features

- Upload files to OpenAI with proper purpose designation
- Track file metadata including status, size, and creation time
- Associate files with specific OpenAI projects (for organizational purposes)
- Handle various file types for different OpenAI services

### Usage

```hcl
module "openai_fine_tune_file" {
  source = "../../modules/files"

  file_path  = "path/to/training_data.jsonl"
  purpose    = "fine-tune"
  project_id = "optional-project-id"  # Optional
}

# The file ID can be used in other resources
resource "openai_fine_tuning" "custom_model" {
  training_file = module.openai_fine_tune_file.file_id
  model         = "gpt-3.5-turbo"
}
```

For more examples and detailed information, see the [module documentation](./files/README.md).

## Security Notes

- Admin API keys have powerful permissions - handle them with care
- Store key values securely in encrypted storage
- Follow the principle of least privilege when assigning scopes
- Ensure Terraform state files are properly secured since they will contain sensitive values
- Consider using a secret management solution for production environments 

## Rate Limits

Rate limits help control API usage and manage costs by setting caps on:

1. **Request Rate Limits (rpm)**: Control how many API requests can be made per minute
2. **Token Rate Limits (tpm)**: Control how many tokens can be processed per minute

Best practices for rate limits:

- Start with conservative limits and adjust as needed
- Set different limits for development and production environments
- Monitor usage patterns to optimize limit settings
- Use higher limits for batch processing jobs and lower limits for real-time applications 

## API Limitations

### Audio Resource Listing Operations

The OpenAI API has limitations regarding audio resources. While you can create audio resources (transcriptions, translations, text-to-speech, speech-to-text) and retrieve individual resources using their specific IDs, the API does not currently support listing operations for these resources.

When using the audio module, be aware that the following data sources are not supported by the OpenAI API:

- `openai_audio_transcriptions`
- `openai_audio_translations` 
- `openai_speech_to_texts`
- `openai_text_to_speechs`

The audio module includes these data sources but they are commented out to prevent errors. If you need to retrieve lists of audio resources, you'll need to implement your own tracking system outside of the provider. 

## Featured Modules

### File Upload with Import Support

The [upload](./upload/) module provides a convenient way to:

- Upload files to OpenAI for various purposes
- Import existing files into Terraform state
- Manage file lifecycle with proper change detection

Example usage:

```hcl
# Upload a new file
module "new_file" {
  source = "../modules/upload"

  purpose   = "fine-tune"
  file_path = "./training_data.jsonl"
}

# After importing an existing file:
# terraform import module.existing_file.openai_file.file file-abc123
module "existing_file" {
  source = "../modules/upload"

  purpose   = "assistants"
  file_path = "./placeholder.jsonl"  # Path is ignored for imported files
}
```

The module automatically detects import mode and adjusts behavior accordingly, using lifecycle configuration to avoid unnecessary recreation of resources.

## Module Design Principles

All modules in this collection follow these principles:

1. **Consistency**: Uniform interface and output structure
2. **Flexibility**: Support for all relevant API features
3. **Isolation**: Modules can be used independently
4. **Documentation**: Comprehensive README with examples
5. **Error Handling**: Robust handling of API errors
6. **Security**: Sensitive values handling with Terraform best practices

## Using the Modules

To use a module in your Terraform configuration:

```hcl
module "example_module" {
  source = "github.com/fjcorp/terraform-provider-openai//modules/module_name"
  
  # Module-specific variables
  variable_1 = "value_1"
  variable_2 = "value_2"
}
```

## Authentication

All modules inherit authentication from the provider configuration:

```hcl
provider "openai" {
  # Authentication via environment variables:
  # OPENAI_API_KEY, OPENAI_ORGANIZATION, etc.
}
```

## Examples

For complete working examples, see the [examples](../examples/) directory.

## Contributing

Contributions to improve these modules are welcome! Please ensure:

1. Modules follow the design principles
2. Documentation is comprehensive
3. Tests are included where applicable 

### Model

The [model](model) module retrieves information about a specific OpenAI model.

Example:

```hcl
module "model" {
  source = "../../modules/model"
  model_id = "gpt-4o"
}

output "model_info" {
  value = "Model ${module.model.model.id} is owned by ${module.model.model.owned_by}"
} 