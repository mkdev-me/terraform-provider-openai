# OpenAI Terraform Provider Documentation

Welcome to the comprehensive documentation for the OpenAI Terraform Provider. This documentation covers all aspects of using and contributing to the provider.

## Documentation Structure

### API Reference
- **[`resources/`](resources/)**: Complete documentation for all Terraform resources
- **[`data-sources/`](data-sources/)**: Complete documentation for all data sources

### Guides and Tutorials
- **[`TROUBLESHOOTING.md`](TROUBLESHOOTING.md)**: Solutions for common issues and error messages
- **[`PROJECTS.md`](PROJECTS.md)**: Guide to working with OpenAI projects
- **[`ORGANIZATION_USERS.md`](ORGANIZATION_USERS.md)**: Managing organization users and permissions
- **[`openai_fine_tuning_resources.md`](openai_fine_tuning_resources.md)**: Comprehensive guide to fine-tuning resources
- **[`IMPORT_LIMITATIONS.md`](IMPORT_LIMITATIONS.md)**: Important information about resource import limitations

### Developer Documentation
- **[`DEVELOPMENT.md`](DEVELOPMENT.md)**: Contributing guide for developers
- **[`DEPENDENCY_MANAGEMENT.md`](DEPENDENCY_MANAGEMENT.md)**: Managing provider dependencies
- **[`IMPLEMENTATION_SUMMARY.md`](IMPLEMENTATION_SUMMARY.md)**: Technical implementation details

## Quick Start

If you're new to the OpenAI Terraform Provider, start here:

1. **[Installation Guide](../README.md#installation)**: Set up the provider
2. **[Basic Usage](../README.md#basic-usage)**: Your first configuration
3. **[Examples](../examples/)**: Working examples for all features
4. **[Authentication](../README.md#authentication-and-api-key-requirements)**: API key setup

## Finding Information

### By Feature
- **Chat & Completions**: See [chat_completion](resources/chat_completion.md), [completion](resources/completion.md)
- **Assistants**: See [assistant](resources/assistant.md), [thread](resources/thread.md), [run](resources/run.md)
- **Fine-Tuning**: See [fine_tuning_job](resources/fine_tuning_job.md), [Fine-Tuning Guide](openai_fine_tuning_resources.md)
- **Files & Storage**: See [file](resources/file.md), [vector_store](resources/vector_store.md)
- **Images**: See [image_generation](resources/image_generation.md), [image_edit](resources/image_edit.md)
- **Audio**: See [audio_transcription](resources/audio_transcription.md), [text_to_speech](resources/text_to_speech.md)
- **Administration**: See [project](resources/project.md), [project_user](resources/project_user.md)

### By Task
- **Import existing resources**: See [Import Limitations](IMPORT_LIMITATIONS.md)
- **Manage API keys**: See [project_api_key](data-sources/project_api_key.md)
- **Set rate limits**: See [rate_limit](resources/rate_limit.md)
- **Work with organizations**: See [Organization Users Guide](ORGANIZATION_USERS.md)

## Best Practices

1. **Security**: Never commit API keys to version control
2. **State Management**: Use remote state for team collaboration
3. **Resource Organization**: Group related resources in modules
4. **Error Handling**: Implement proper error handling and retries
5. **Cost Management**: Monitor token usage and set appropriate limits

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/mkdev-me/terraform-provider-openai/issues)
- **Examples**: [Example Configurations](../examples/)
- **API Reference**: [OpenAI API Documentation](https://platform.openai.com/docs/api-reference)

## Contributing

We welcome contributions to improve this documentation:

1. **Format**: Use clear, consistent Markdown formatting
2. **Examples**: Include practical, working examples
3. **Accuracy**: Ensure technical accuracy and test all examples
4. **Clarity**: Write for both beginners and experienced users

See [DEVELOPMENT.md](DEVELOPMENT.md) for the complete contribution guide. 