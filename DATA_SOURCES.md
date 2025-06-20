# OpenAI Terraform Provider Data Sources

This document provides a comprehensive overview of all data sources implemented in the OpenAI Terraform Provider.

## Data Source Overview

The provider implements a total of **56 data sources** in its Go code, covering various OpenAI API functionalities:

| # | Data Source Name | Description | Documentation |
|---|-----------------|-------------|---------------|
| 1 | `openai_admin_api_key` | Retrieve information about a specific admin API key | [Documentation](docs/data-sources/system_api.md) |
| 2 | `openai_admin_api_keys` | List all admin API keys | [Documentation](docs/data-sources/system_apis.md) |
| 3 | `openai_assistant` | Retrieve information about a specific assistant | [Documentation](docs/data-sources/assistant.md) |
| 4 | `openai_assistants` | List all assistants | [Documentation](docs/data-sources/assistants.md) |
| 5 | `openai_audio_transcription` | Retrieve information about a specific audio transcription | [Documentation](docs/data-sources/audio_transcription.md) |
| 6 | `openai_audio_transcriptions` | List all audio transcriptions | [Documentation](docs/data-sources/audio_transcriptions.md) |
| 7 | `openai_audio_translation` | Retrieve information about a specific audio translation | [Documentation](docs/data-sources/audio_translation.md) |
| 8 | `openai_audio_translations` | List all audio translations | [Documentation](docs/data-sources/audio_translations.md) |
| 9 | `openai_batch` | Retrieve information about a specific batch operation | [Documentation](docs/data-sources/batch.md) |
| 10 | `openai_batches` | List all batch operations | [Documentation](docs/data-sources/batches.md) |
| 11 | `openai_chat_completion` | Retrieve information about a specific chat completion | [Documentation](docs/data-sources/chat_completion.md) |
| 12 | `openai_chat_completion_messages` | List messages for a chat completion | [Documentation](docs/data-sources/chat_completion_messages.md) |
| 13 | `openai_chat_completions` | List all chat completions | [Documentation](docs/data-sources/chat_completions.md) |
| 14 | `openai_file` | Retrieve information about a specific file | [Documentation](docs/data-sources/file.md) |
| 15 | `openai_files` | List all files | [Documentation](docs/data-sources/files.md) |
| 16 | `openai_fine_tuning_checkpoint_permissions` | List permissions for a fine-tuning checkpoint | [Documentation](docs/data-sources/fine_tuning_checkpoint_permissions.md) |
| 17 | `openai_fine_tuning_checkpoints` | List checkpoints for a fine-tuning job | [Documentation](docs/data-sources/fine_tuning_checkpoints.md) |
| 18 | `openai_fine_tuning_events` | List events for a fine-tuning job | [Documentation](docs/data-sources/fine_tuning_events.md) |
| 19 | `openai_fine_tuning_job` | Retrieve information about a specific fine-tuning job | [Documentation](docs/data-sources/fine_tuning_job.md) |
| 20 | `openai_fine_tuning_jobs` | List all fine-tuning jobs | [Documentation](docs/data-sources/fine_tuning_jobs.md) |
| 21 | `openai_invite` | Retrieve information about a specific invite | [Documentation](docs/data-sources/invite.md) |
| 22 | `openai_invites` | List all invites | [Documentation](docs/data-sources/invites.md) |
| 23 | `openai_message` | Retrieve information about a specific thread message | [Documentation](docs/data-sources/message.md) |
| 24 | `openai_messages` | List all messages in a thread | [Documentation](docs/data-sources/messages.md) |
| 25 | `openai_model` | Retrieve information about a specific model | [Documentation](docs/data-sources/model.md) |
| 26 | `openai_model_response` | Retrieve information about a specific model response | [Documentation](docs/data-sources/model_response.md) |
| 27 | `openai_model_response_input_items` | List input items for a model response | [Documentation](docs/data-sources/model_response_input_items.md) |
| 28 | `openai_model_responses` | List all model responses | [Documentation](docs/data-sources/model_responses.md) |
| 29 | `openai_models` | List all available models | [Documentation](docs/data-sources/models.md) |
| 30 | `openai_project` | Retrieve information about a specific project | [Documentation](docs/data-sources/project.md) |
| 31 | `openai_project_api_key` | Retrieve information about a specific project API key | [Documentation](docs/data-sources/project_api_key.md) |
| 32 | `openai_project_api_keys` | List all API keys for a project | [Documentation](docs/data-sources/project_api_keys.md) |
| 33 | `openai_project_resources` | List resources associated with a project | [Documentation](docs/data-sources/README.md) |
| 34 | `openai_project_service_account` | Retrieve information about a specific project service account | [Documentation](docs/data-sources/project_service_account.md) |
| 35 | `openai_project_service_accounts` | List all service accounts for a project | [Documentation](docs/data-sources/project_service_accounts.md) |
| 36 | `openai_project_user` | Retrieve information about a specific project user | [Documentation](docs/data-sources/project_user.md) |
| 37 | `openai_project_users` | List all users in a project | [Documentation](docs/data-sources/project_users.md) |
| 38 | `openai_projects` | List all projects | [Documentation](docs/data-sources/projects.md) |
| 39 | `openai_rate_limit` | Retrieve information about a specific rate limit | [Documentation](docs/data-sources/rate_limit.md) |
| 40 | `openai_rate_limits` | List all rate limits | [Documentation](docs/data-sources/rate_limits.md) |
| 41 | `openai_run` | Retrieve information about a specific run | [Documentation](docs/data-sources/run.md) |
| 42 | `openai_runs` | List all runs | [Documentation](docs/data-sources/runs.md) |
| 43 | `openai_speech_to_text` | Retrieve information about a specific speech-to-text conversion | [Documentation](docs/data-sources/speech_to_text.md) |
| 44 | `openai_speech_to_texts` | List all speech-to-text conversions | [Documentation](docs/data-sources/speech_to_texts.md) |
| 45 | `openai_text_to_speech` | Retrieve information about a specific text-to-speech conversion | [Documentation](docs/data-sources/text_to_speech.md) |
| 46 | `openai_text_to_speechs` | List all text-to-speech conversions | [Documentation](docs/data-sources/text_to_speechs.md) |
| 47 | `openai_thread` | Retrieve information about a specific thread | [Documentation](docs/data-sources/thread.md) |
| 48 | `openai_thread_run` | Retrieve information about a specific thread run | [Documentation](docs/data-sources/thread_run.md) |
| 49 | `openai_organization_user` | Retrieve information about a specific user in your organization | [Documentation](docs/data-sources/organization_user.md) |
| 50 | `openai_organization_users` | List all users in your organization | [Documentation](docs/data-sources/organization_users.md) |
| 51 | `openai_vector_store` | Retrieve information about a specific vector store | [Documentation](docs/data-sources/vector_store.md) |
| 52 | `openai_vector_store_file` | Retrieve information about a specific file in a vector store | [Documentation](docs/data-sources/vector_store_file.md) |
| 53 | `openai_vector_store_file_batch` | Retrieve information about a specific file batch in a vector store | [Documentation](docs/data-sources/vector_store_file_batch.md) |
| 54 | `openai_vector_store_file_batch_files` | List files in a vector store file batch | [Documentation](docs/data-sources/vector_store_file_batch_files.md) |
| 55 | `openai_vector_store_file_content` | Retrieve content of a file in a vector store | [Documentation](docs/data-sources/vector_store_file_content.md) |
| 56 | `openai_vector_store_files` | List files in a vector store | [Documentation](docs/data-sources/vector_store_files.md) |
| 57 | `openai_vector_stores` | List all vector stores | [Documentation](docs/data-sources/vector_stores.md) |


## Usage Notes

1. All data sources require authentication via the `OPENAI_API_KEY` environment variable
2. Many examples demonstrate proper data source usage patterns
3. Consult the documentation for each data source for field descriptions and usage details
4. The example directories provide working configurations for most common use cases

