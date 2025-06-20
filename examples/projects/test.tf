# Test file for the projects data source

# Retrieve all projects using the projects data source
data "openai_projects" "test_all_projects" {
}

# Output results
output "test_projects_count" {
  value = length(data.openai_projects.test_all_projects.projects)
}

output "test_project_names" {
  value = [for p in data.openai_projects.test_all_projects.projects : p.name]
} 