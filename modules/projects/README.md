# OpenAI Projects Module

This module provides reusable components for managing OpenAI projects using Terraform. It includes resources for creating projects and data sources for retrieving project information.

## Module Components

The module consists of:

1. `main.tf` - Resources for creating and managing OpenAI projects
2. `data_source.tf` - Data sources for retrieving information about existing projects
3. `variables.tf` - Input variable definitions
4. `outputs.tf` - Output definitions

## Usage

### Creating a New Project

```hcl
module "openai_project" {
  source = "../../modules/projects"

  name        = "My Project"

}

output "project_id" {
  value = module.openai_project.project_id
}
```

### Retrieving Information about an Existing Project

```hcl
module "project_info" {
  source = "../../modules/projects"

  # Set create_project to false to use the data source instead of creating a project
  create_project = false
  project_id     = "proj_abc123xyz"  # Replace with a real project ID

}

output "project_name" {
  value = module.project_info.project_name
}

output "project_status" {
  value = module.project_info.project_status
}

output "project_created_at" {
  value = module.project_info.project_created_at
}
```

### Retrieving a List of All Projects

```hcl
module "all_projects" {
  source = "../../modules/projects"

  # Set list_mode to true to retrieve all projects
  create_project = false
  list_mode      = true
}

output "project_count" {
  value = module.all_projects.project_count
}

output "project_names" {
  value = module.all_projects.project_names
}

output "project_ids" {
  value = module.all_projects.project_ids
}
```

## Input Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| `create_project` | Whether to create a new project or use data source mode | `bool` | `true` | no |
| `list_mode` | Whether to retrieve all projects instead of working with a single project | `bool` | `false` | no |
| `name` | Name of the project | `string` | `null` | yes, when create_project is true |
| `project_id` | ID of the project to use in data source mode | `string` | `null` | yes, when create_project is false and list_mode is false |
| `is_default` | Whether this project should be the default project | `bool` | `false` | no |
| `openai_admin_key` | OpenAI Admin API key | `string` | `null` | no |
| `rate_limits` | Rate limits for the project | `list(object)` | `[]` | no |
| `users` | Users to add to the project | `list(object)` | `[]` | no |

## Outputs

### Single Project Mode (list_mode = false)

| Name | Description |
|------|-------------|
| `project_id` | The ID of the project |
| `project_name` | The name of the project |
| `project_status` | The status of the project |
| `project_created_at` | When the project was created |
| `project_usage_limits` | Usage limits for the project |

### List Mode (list_mode = true)

| Name | Description |
|------|-------------|
| `projects` | List of all projects |
| `project_count` | Number of projects |
| `project_names` | Names of all projects |
| `project_ids` | IDs of all projects |

## Examples

See the [examples/projects](../../examples/projects) directory for examples of how to use this module.

## Technical Notes

- When `create_project` is set to `true`, this module will create a new OpenAI project
- When `create_project` is set to `false` and `list_mode` is set to `false`, this module will use the data source to retrieve information about an existing project
- When `list_mode` is set to `true`, this module will retrieve a list of all projects in the OpenAI account
- The OpenAI Admin API key must have permission to create and manage projects 