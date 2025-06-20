# OpenAI Terraform Provider Resources

This document provides a comprehensive overview of all resources implemented in the OpenAI Terraform Provider.

## Resource Overview

The provider implements a total of **40 resources** in its Go code, covering various OpenAI API functionalities:

| # | Resource Name | Description | Documentation |
|---|--------------|-------------|---------------|
| 1 | `openai_file` | Manage file uploads to OpenAI | [Documentation](docs/resources/file.md) |
| 2 | `openai_chat_completion` | Create chat completions with OpenAI models | [Documentation](docs/resources/chat_completion.md) |
| 3 | `openai_moderation` | Check content against OpenAI's moderation policies | [Documentation](docs/resources/moderation.md) |
| 4 | `openai_embedding` | Generate embeddings for text | [Documentation](docs/resources/embedding.md) |
| 5 | `openai_model` | Manage custom models | [Documentation](docs/resources/model.md) |
| 6 | `openai_project` | Manage OpenAI projects | [Documentation](docs/resources/project.md) |
| 7 | `openai_batch` | Process batch operations | [Documentation](docs/resources/batch.md) |
| 8 | `openai_text_to_speech` | Convert text to speech | [Documentation](docs/resources/text_to_speech.md) |
| 9 | `openai_audio_transcription` | Transcribe audio to text | [Documentation](docs/resources/audio_transcription.md) |
| 10 | `openai_audio_translation` | Translate audio to text | [Documentation](docs/resources/audio_translation.md) |
| 11 | `openai_speech_to_text` | Convert speech to text | [Documentation](docs/resources/speech_to_text.md) |
| 12 | `openai_image_generation` | Generate images from prompts | [Documentation](docs/resources/image_generation.md) |
| 13 | `openai_image_edit` | Edit existing images | [Documentation](docs/resources/image_edit.md) |
| 14 | `openai_image_variation` | Create variations of images | [Documentation](docs/resources/image_variation.md) |
| 15 | `openai_completion` | Generate text completions | [Documentation](docs/resources/completion.md) |
| 16 | `openai_edit` | Edit text with instructions | [Documentation](docs/resources/edit.md) |
| 17 | `openai_fine_tuned_model` | Create fine-tuned models | [Documentation](docs/resources/fine_tuned_model.md) |
| 18 | `openai_assistant` | Manage OpenAI assistants | [Documentation](docs/resources/assistant.md) |
| 19 | `openai_thread` | Manage conversation threads | [Documentation](docs/resources/thread.md) |
| 20 | `openai_message` | Manage messages in threads | [Documentation](docs/resources/message.md) |
| 21 | `openai_run` | Execute assistants on threads | [Documentation](docs/resources/run.md) |
| 22 | `openai_model_response` | Generate responses with detailed token usage | [Documentation](docs/resources/model_response.md) |
| 23 | `openai_vector_store` | Create and manage a vector store | [Documentation](docs/resources/vector_store.md) |
| 24 | `openai_vector_store_file` | Add a file to a vector store | [Documentation](docs/resources/vector_store_file.md) |
| 25 | `openai_vector_store_file_batch` | Add multiple files to a vector store at once | [Documentation](docs/resources/vector_store_file_batch.md) |
| 26 | `openai_thread_run` | Run assistants on threads (alternative to openai_run) | [Documentation](docs/resources/thread_run.md) |
| 27 | `openai_rate_limit` | Manage rate limits for API requests | [Documentation](docs/resources/rate_limit.md) |
| 28 | `openai_project_user` | Manage users in OpenAI projects | [Documentation](docs/resources/project_user.md) |
| 29 | `openai_invite` | Manage invitations to OpenAI projects | [Documentation](docs/resources/invite.md) |
| 30 | `openai_fine_tuning_job` | Create and manage fine-tuning jobs | [Documentation](docs/resources/fine_tuning_job.md) |
| 31 | `openai_fine_tuning_checkpoint_permission` | Manage permissions for fine-tuning checkpoints | [Documentation](docs/resources/fine_tuning_checkpoint_permission.md) |
| 32 | `openai_upload` | Upload files to OpenAI | [Documentation](docs/resources/upload.md) |
| 33 | `openai_project_api_key` | Manage API keys for projects | [Documentation](docs/resources/system_api.md) |
| 34 | `openai_project_service_account` | Manage service accounts for projects | [Documentation](docs/resources/project_service_account.md) |
| 35 | `openai_admin_api_key` | Manage admin API keys | [Documentation](docs/resources/system_api.md) |
| 36 | `openai_assistant_file` | Manage files attached to assistants | [Documentation](docs/resources/assistant.md) |
| 37 | `openai_run_step` | Manage individual steps within a run | [Documentation](docs/resources/run.md) |
| 38 | `openai_evaluation` | Evaluate model outputs | [Documentation](docs/resources/README.md) |
| 39 | `openai_playground_config` | Configure OpenAI playground settings | [Documentation](docs/resources/README.md) |

The provider supports a wide range of OpenAI resources, covering all core functionalities as well as advanced features.


## Example Directories

The provider includes numerous example directories to demonstrate resource usage:

- **audio/**: Audio processing examples including text-to-speech, transcription, and translation
- **assistants/**: Assistants examples showing how to create, configure, and list assistants
- **batch/**: Batch processing examples
- **chat_completion/**: Examples of chat completion functionality
- **embeddings/**: Text embedding examples
- **files/**: File upload and management examples
- **fine_tuning/**: Fine-tuned model examples
- **image/**: Image generation, editing, and variation examples  
- **model_response/**: Model response generation with comprehensive usage statistics
- **moderation/**: Content moderation examples
- **project/**: Project management examples
- **project_api/**: Project API key management examples
- **service_account/**: Service account management examples
- **simple/**: Simple examples demonstrating basic provider usage
- **system_api/**: System-level API management examples
- **upload/**: File upload examples
- **vector_store/**: Vector database integration examples

