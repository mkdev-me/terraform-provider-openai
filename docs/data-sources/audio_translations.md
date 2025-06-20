---
page_title: "OpenAI: openai_audio_translations Data Source"
subcategory: ""
description: |-
  Data source for OpenAI audio translations (Not supported by OpenAI API).
---

# Data Source: openai_audio_translations

> **Important Note:** This data source is included for documentation purposes, but it will result in an error when used. The OpenAI API does not currently support listing all audio translations. You can only retrieve individual translations using the `openai_audio_translation` (singular) data source with a specific `translation_id`.

This data source would theoretically allow you to list all audio translations, but the OpenAI API does not provide an endpoint for this operation.

## Example Usage (Will Result in Error)

```terraform
data "openai_audio_translations" "all" {
  # Optional filter by model (not supported)
  model = "whisper-1"
}

# This will fail with an error from the OpenAI API
output "all_translations" {
  value = data.openai_audio_translations.all.translations
}
```

## Supported Alternative

Instead of trying to list all translations, you should use the `openai_audio_translation` (singular) data source to retrieve a specific translation by ID:

```terraform
data "openai_audio_translation" "example" {
  translation_id = "translation-1234567890"
}

output "translation_text" {
  value = data.openai_audio_translation.example.text
}
```

## API Limitation Details

The OpenAI API does not provide endpoints for listing operations for audio resources. This is a limitation of the API itself, not the Terraform provider. If you need to track multiple translations, consider implementing your own tracking system outside of this provider.

For more information about working with individual audio translations, see the [openai_audio_translation](./audio_translation.md) data source documentation.

## Argument Reference

The following arguments are supported:

* `project_id` - (Optional) The ID of the project to retrieve audio translations from. If not specified, the API key's default project will be used.
* `api_key` - (Optional) Project-specific API key to use for authentication. If not provided, the provider's default API key will be used.
* `model` - (Optional) Filter by model. Options include 'whisper-1'.

## Attributes Reference

The following attributes are exported:

* `id` - A unique identifier for this data source.
* `translations` - A list of audio translations, each containing the following attributes:
  * `id` - The ID of the audio translation.
  * `created_at` - The timestamp when the translation was created.
  * `status` - The status of the translation.
  * `model` - The model used for translation.
  * `text` - The translated text.
  * `duration` - The duration of the audio in seconds.

## Import

This is a data source and does not support import. 