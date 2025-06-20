# OpenAI Project Management with Terraform

This example demonstrates how to manage OpenAI projects and retrieve project information using Terraform. It includes examples for project resources and data sources.

## Example Files

This directory contains two Terraform configuration files:

1. `main.tf` - Demonstrates how to create and manage OpenAI projects
2. `data_sources.tf` - Demonstrates how to use the OpenAI project data sources

## Resources and Data Sources

1. `openai_project` (Resource) - Creates and manages OpenAI projects
2. `openai_project` (Data Source) - Retrieves information about a specific OpenAI project
3. `openai_projects` (Data Source) - Retrieves information about all OpenAI projects in an organization

## Prerequisites

- An OpenAI account with administrator privileges
- An OpenAI Admin API key with appropriate permissions
- Terraform >= 0.13.0

## Setup

1. Create a `terraform.tfvars` file with your credentials:

```hcl
openai_admin_key = "sk-xxxx"  # Your OpenAI Admin API key
```

## Usage

1. Initialize Terraform:
```bash
terraform init
```

2. Apply the configuration:
```bash
terraform apply -var="openai_admin_key=$OPENAI_ADMIN_KEY"
```

## Example 1: Creating an OpenAI Project (main.tf)

This example demonstrates how to create and manage an OpenAI project.

```hcl
resource "openai_project" "example" {
  name = "Example Project"
}

output "project_id" {
  value = openai_project.example.id
}
```

## Example 2: Retrieving Project Information (data_sources.tf)

This example demonstrates how to retrieve information about existing OpenAI projects.

### Data Source: openai_project

This data source allows you to retrieve detailed information about a specific OpenAI project.

**Example:**
```hcl
data "openai_project" "project_info" {
  project_id = "proj_abc123xyz" # Replace with a real project ID
}

output "project_name" {
  value = data.openai_project.project_info.name
}

output "project_status" {
  value = data.openai_project.project_info.status
}

output "project_created_at" {
  value = data.openai_project.project_info.created_at
}

output "project_usage_limits" {
  value = data.openai_project.project_info.usage_limits
}
```

**Available Attributes:**
- `project_id` (Required): The ID of the project to retrieve information about
- `name` (Computed): The name of the project
- `status` (Computed): The current status of the project
- `created_at` (Computed): Timestamp when the project was created
- `usage_limits` (Computed): Usage limits for the project, including:
  - `max_monthly_dollars`: Maximum monthly spend in dollars
  - `max_parallel_requests`: Maximum number of parallel requests allowed
  - `max_tokens`: Maximum number of tokens per request

### Data Source: openai_projects

This data source allows you to retrieve a list of all OpenAI projects in your organization.

**Example:**
```hcl
data "openai_projects" "all_projects" {
  # No specific parameters required
}

output "all_projects_count" {
  value = length(data.openai_projects.all_projects.projects)
}

output "all_project_names" {
  value = [for p in data.openai_projects.all_projects.projects : p.name]
}

output "all_project_ids" {
  value = [for p in data.openai_projects.all_projects.projects : p.id]
}
```

**Available Attributes:**
- `projects` (Computed): A list of all projects in the organization, each containing:
  - `id`: The ID of the project
  - `name`: The name of the project
  - `status`: The current status of the project
  - `created_at`: Timestamp when the project was created

## Importing Existing Projects

You can import existing OpenAI projects into Terraform state to manage them with Terraform:

```bash
terraform import -var="openai_admin_key=$OPENAI_ADMIN_API" openai_project.example proj_abc123xyz
```

Where:
- `openai_project.example` is the resource ID in your Terraform configuration
- `proj_abc123xyz` is the ID of the existing OpenAI project you want to import

### Example Workflow

1. Define the project resource in your Terraform configuration:

```hcl
resource "openai_project" "imported" {
  name = "Project to be imported"  # This will be updated to match the actual project name
}
```

2. Import the existing project:

```bash
terraform import -var="openai_admin_key=$OPENAI_ADMIN_API" openai_project.imported proj_ECvwuJcFiayw3x7AcbkRE1Mn
```

3. Verify the import:

```bash
terraform state show openai_project.imported
```

After importing, Terraform will fetch the project's actual name, creation date, and status from the OpenAI API. You can then manage the project through Terraform.

Note: Import requires an API key with sufficient permissions to read project information.

## Technical Details

- The OpenAI project data sources retrieve information directly from the OpenAI API
- These data sources are read-only and don't make any changes to your OpenAI resources
- Make sure to use `depends_on` if you're retrieving information about a project you're also creating in the same Terraform configuration
- The `openai_projects` data source is useful for:
  - Auditing projects in your organization
  - Creating dynamic configurations based on existing projects
  - Generating reports on project usage and configuration

## Troubleshooting

### Common Issues

1. **404 Not Found errors**: 
   - Ensure the project ID is correct
   - Check that the project exists in your OpenAI organization

2. **Permission errors**:
   - Verify your Admin API key has sufficient permissions
   - Use environment variables to avoid hardcoding sensitive keys

3. **API connection issues**:
   - Check your network connectivity
   - Verify the OpenAI API is up and running 