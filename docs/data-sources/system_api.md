# Admin API Key Data Source

The `openai_admin_api_key` data source allows you to retrieve information about a specific OpenAI admin API key.

## Example Usage

```hcl
data "openai_admin_api_key" "example" {
  api_key_id = "key_abc123"
}

output "api_key_name" {
  value = data.openai_admin_api_key.example.name
}

output "api_key_created_at" {
  value = data.openai_admin_api_key.example.created_at
}
```

### Using a Custom API Key for Authentication

```hcl
data "openai_admin_api_key" "custom_auth" {
  api_key_id = "key_abc123"
  api_key    = "sk-adm-your-admin-api-key"  # The admin API key to use for this operation
}
```

## Argument Reference

* `api_key_id` - (Required) The ID of the admin API key to retrieve.
* `api_key` - (Optional) A custom admin API key to use for this data source. If not provided, the provider's default API key will be used.

## Attribute Reference

* `id` - The ID of the admin API key.
* `name` - The name of the admin API key.
* `created_at` - A timestamp of when the API key was created, formatted as an RFC3339 string.
* `expires_at` - Unix timestamp representing when the key expires. Will be absent if the key has no expiration.
* `scopes` - List of scopes associated with the API key.
* `object` - The object type, which is typically "api_key".

## Special Notes

* **Permissions**: To retrieve admin API keys, you need an API key with administrative permissions (`api.management.read`).
* **Security**: This data source does not return the actual API key value, as that value cannot be retrieved after creation, even by the OpenAI API.
* **Timestamp Format**: The `created_at` field is returned as an RFC3339 formatted string for better readability.

## Related Resources

* [`openai_admin_api_key` Resource](../resources/system_api.md)
* [`openai_admin_api_keys` Data Source](../data-sources/system_apis.md)
* [Admin API Key Module](../../modules/admin_api_key/README.md) 