# Admin API Key Resource

The `openai_admin_api_key` resource allows you to programmatically create and manage OpenAI admin API keys with specific permissions.

Unlike project API keys, admin API keys can be created programmatically and have broader permissions that can manage organization resources.

## Example Usage

```hcl
# Create a new admin API key
resource "openai_admin_api_key" "example" {
  name       = "terraform-managed-admin-key"
  expires_at = 1735689600  # Optional: Unix timestamp (Dec 31, 2024)
  scopes     = ["api.management.read", "api.management.write"]
}

output "api_key_id" {
  value = openai_admin_api_key.example.id
}

output "api_key_value" {
  value     = openai_admin_api_key.example.api_key_value
  sensitive = true
}
```

### Creating an Admin API Key with a Custom API Key

```hcl
# Create an admin API key using a specific admin API key for authentication
resource "openai_admin_api_key" "custom_auth" {
  name       = "created-with-custom-key"
  api_key    = "sk-adm-your-admin-api-key"  # The admin API key to use for this operation
  expires_at = 1735689600  # Optional
  scopes     = ["api.management.read"]
}
```

## Argument Reference

* `name` - (Required) The name of the admin API key.
* `api_key` - (Optional) A custom admin API key to use for this resource. If not provided, the provider's default API key will be used.
* `expires_at` - (Optional) Unix timestamp for when the key should expire. If not specified, the key will not expire.
* `scopes` - (Optional) List of scopes to assign to the API key. Available scopes include `api.management.read` and `api.management.write`.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the created admin API key.
* `api_key_value` - The actual API key value. This is only available when the key is first created and cannot be retrieved later.
* `object` - The object type, which is typically "api_key".
* `created_at` - Timestamp when the API key was created.

## Import

Admin API keys can be imported using the API key ID. However, note that you cannot import the actual API key value, as OpenAI does not allow retrieving this value after creation:

```
terraform import openai_admin_api_key.example key_abc123
```

## Special Notes

* **Security**: The API key value is sensitive and should be treated as a secret. Once created, the value can never be retrieved again, even by the OpenAI API.
* **Permissions**: To create admin API keys, you need an API key with administrative permissions (`api.management.read`, `api.management.write`).
* **Creating vs. Reading**: This resource is for creating new API keys. To read existing keys, use the `openai_admin_api_key` data source instead.
* **Scopes**: Assign appropriate scopes to limit the permissions of the generated keys following the principle of least privilege.
* **Expiration**: Consider setting an expiration date for security best practices.

## Related Resources

* [`openai_admin_api_key` Data Source](../data-sources/system_api.md)
* [`openai_admin_api_keys` Data Source](../data-sources/system_apis.md) 