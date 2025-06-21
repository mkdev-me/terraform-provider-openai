# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.1] - 2025-06-21

### Fixed
- Fixed pagination issue when fetching all projects - now retrieves all pages instead of just the first page
- Cleaned up hardcoded version references in example files
- Removed unnecessary API key mentions from examples
- Improved code organization by reordering imports

## [1.0.0] - 2025-06-20

### Added
- Initial release of the Terraform Provider for OpenAI
- Provider configuration with support for API keys and organization ID
- Resource: `openai_chat_completion` - Manage chat completions
- Resource: `openai_embedding` - Create embeddings
- Resource: `openai_file` - Manage files for fine-tuning and assistants
- Resource: `openai_fine_tuning_job` - Create and manage fine-tuning jobs
- Resource: `openai_image` - Generate and edit images
- Resource: `openai_audio_transcription` - Transcribe audio files
- Resource: `openai_audio_translation` - Translate audio files
- Resource: `openai_audio_speech` - Generate speech from text
- Resource: `openai_assistant` - Manage AI assistants
- Resource: `openai_assistant_file` - Attach files to assistants
- Resource: `openai_thread` - Manage conversation threads
- Resource: `openai_message` - Create messages in threads
- Resource: `openai_run` - Execute assistant runs
- Resource: `openai_vector_store` - Manage vector stores
- Resource: `openai_vector_store_file` - Manage files in vector stores
- Resource: `openai_vector_store_file_batch` - Batch operations for vector store files
- Resource: `openai_organization_invite` - Manage organization invitations
- Resource: `openai_organization_user` - Manage organization users
- Resource: `openai_project` - Manage projects
- Resource: `openai_project_rate_limit` - Configure project rate limits
- Resource: `openai_project_service_account` - Manage project service accounts
- Resource: `openai_project_user` - Manage project users
- Data Source: `openai_file` - Read file information
- Data Source: `openai_fine_tuning_job` - Read fine-tuning job information
- Data Source: `openai_model` - Get model information
- Data Source: `openai_models` - List available models
- Data Source: `openai_organization_audit_logs` - Read organization audit logs
- Data Source: `openai_organization_invites` - List organization invitations
- Data Source: `openai_organization_project` - Read project information
- Data Source: `openai_organization_projects` - List organization projects
- Data Source: `openai_organization_users` - List organization users
- Data Source: `openai_project_api_key` - Read project API key information
- Data Source: `openai_project_api_keys` - List project API keys
- Data Source: `openai_project_rate_limits` - List project rate limits
- Data Source: `openai_project_service_account` - Read service account information
- Data Source: `openai_project_service_accounts` - List project service accounts
- Data Source: `openai_project_user` - Read project user information
- Data Source: `openai_project_users` - List project users
- Comprehensive documentation for all resources and data sources
- Example configurations for common use cases
- Reusable Terraform modules for common patterns

[Unreleased]: https://github.com/mkdev-me/terraform-provider-openai/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mkdev-me/terraform-provider-openai/releases/tag/v0.1.0