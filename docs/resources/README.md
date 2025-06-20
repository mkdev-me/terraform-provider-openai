# OpenAI Provider Resources

This directory contains documentation for all resources in the OpenAI Terraform Provider. Resources allow you to create, update, and delete resources in the OpenAI API.

## General Usage

Resources follow this general pattern:

```hcl
resource "openai_resource_name" "example" {
  # Required parameters
  parameter_one = "value"
  
  # Optional parameters
  parameter_two = "value"
}

output "resource_id" {
  value = openai_resource_name.example.id
}
```

