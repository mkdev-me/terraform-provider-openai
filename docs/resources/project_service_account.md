# Project Service Account Resource

The `openai_project_service_account` resource allows you to create and manage service accounts within OpenAI projects.

Service accounts are bot users that are not associated with individual human users. Unlike regular users, service accounts are not affected when a user leaves an organization - their API keys and project memberships continue to work uninterrupted. This makes them ideal for long-running applications and services.

## Example Usage

```hcl
# Create a project
resource "openai_project" "example" {
  name        = "My Project with Service Account"
  description = "A project that uses a service account"
}

# Create a service account in the project
resource "openai_project_service_account" "app_service" {
  project_id = openai_project.example.id
  name       = "Production App"
}

# Create an API key for the service account
resource "openai_project_api_key" "app_key" {
  project_id         = openai_project.example.id
  name               = "Production API Key"
  service_account_id = openai_project_service_account.app_service.service_account_id
}

# Output the API key for use in applications
output "api_key" {
  value     = openai_project_api_key.app_key.key
  sensitive = true
}
```

## Argument Reference

* `project_id` - (Required) The ID of the project to which the service account belongs. Cannot be changed after creation.
* `name` - (Required) The name of the service account. Cannot be changed after creation.
* `api_key` - (Optional) A custom API key to use for creating the service account. If not provided, the provider's default API key will be used.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - A composite ID in the format `{project_id}:{service_account_id}`.
* `service_account_id` - The unique identifier for the service account.
* `created_at` - The timestamp (in Unix time) when the service account was created.

## Import

Service accounts can be imported using a composite ID in the format `{project_id}:{service_account_id}`:

```
$ terraform import openai_project_service_account.example proj-123abc:svc-456def
```

## Limitations and Special Notes

### Immutability

Service accounts are largely immutable. Once created, you cannot change the name or other properties. If you need to change a service account's properties, you must delete it and create a new one.

### API Key Management

Service accounts don't automatically come with API keys. You must create API keys separately using the `openai_project_api_key` resource, specifying the service account ID.

### Deletion

When a service account is deleted:

1. All API keys associated with the service account will stop working immediately.
2. Any resources created by the service account will continue to exist but can no longer be managed using that service account's credentials.

### Permissions

To create and manage service accounts:

* You must use an API key with adequate permissions (typically an administrator key).
* Project-specific API keys typically do not have sufficient permissions to manage service accounts.

## Best Practices

1. **Use descriptive names**: Make service account names descriptive and include information about their purpose.

2. **Limit the number of service accounts**: Create only as many service accounts as needed to maintain clear separation of concerns.

3. **Regularly rotate API keys**: Create new API keys for service accounts periodically and update your applications to use the new keys.

4. **Document your service accounts**: Keep documentation about what each service account is used for and which systems depend on it.

5. **Service accounts vs. user accounts**:
   * Use service accounts for automated systems, CI/CD pipelines, and applications.
   * Use regular user accounts for human access.

6. **Set up monitoring**: Monitor the activity of service accounts to detect unusual patterns that might indicate compromise. 