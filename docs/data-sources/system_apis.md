# Admin API Keys Data Source

The `openai_admin_api_keys` data source allows you to retrieve a list of all OpenAI admin API keys associated with your organization.

## Example Usage

```hcl
data "openai_admin_api_keys" "all" {}

output "all_api_keys" {
  value = data.openai_admin_api_keys.all.api_keys
}

output "number_of_api_keys" {
  value = length(data.openai_admin_api_keys.all.api_keys)
}
```

### Filtering and Pagination

```hcl
data "openai_admin_api_keys" "paginated" {
  limit = 10
  after = "key_abc123"  # Optional: Start listing after this API key ID
}
```

### Using a Custom API Key for Authentication

```hcl
data "openai_admin_api_keys" "custom_auth" {
  api_key = "sk-adm-your-admin-api-key"  # The admin API key to use for this operation
  limit   = 20
}
```

## Argument Reference

* `api_key` - (Optional) A custom admin API key to use for this data source. If not provided, the provider's default API key will be used.
* `limit` - (Optional) Maximum number of API keys to retrieve. Defaults to 20 if not specified.
* `after` - (Optional) API key ID to start listing from (for pagination).

## Attribute Reference

* `api_keys` - List of admin API keys. Each API key contains:
  * `id` - The ID of the admin API key.
  * `name` - The name of the admin API key.
  * `created_at` - A timestamp of when the API key was created, formatted as an RFC3339 string.
  * `expires_at` - Unix timestamp representing when the key expires. Will be absent if the key has no expiration.
  * `last_used_at` - A timestamp of when the API key was last used, formatted as an RFC3339 string.
  * `scopes` - List of scopes associated with the API key.
  * `object` - The object type, which is typically "api_key".
* `has_more` - A boolean indicating whether there are more API keys available beyond the limit.

## Special Notes

* **Permissions**: To list admin API keys, you need an API key with administrative permissions (`api.management.read`).
* **Security**: This data source does not return the actual API key values, as those values cannot be retrieved after creation, even by the OpenAI API.
* **Pagination**: For organizations with many API keys, use the `limit` and `after` parameters to paginate the results.
* **Timestamp Format**: The `created_at` and `last_used_at` fields are returned as RFC3339 formatted strings for better readability.

## Related Resources

* [`openai_admin_api_key` Resource](../resources/system_api.md)
* [`openai_admin_api_key` Data Source](../data-sources/system_api.md)
* [Admin API Key Module](../../modules/admin_api_key/README.md) 