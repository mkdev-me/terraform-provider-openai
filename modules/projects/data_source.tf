# Project Data Source - only used when create_project is false and list_mode is false
data "openai_project" "project" {
  count      = (!var.create_project && !var.list_mode) ? 1 : 0
  project_id = var.project_id
}

# Projects Data Source - only used when list_mode is true
data "openai_projects" "all" {
  count = var.list_mode ? 1 : 0
} 