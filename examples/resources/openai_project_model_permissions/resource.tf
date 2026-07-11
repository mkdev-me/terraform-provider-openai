resource "openai_project_model_permissions" "production" {
  project_id = "proj_abc123"
  mode       = "deny_list"
  models     = ["o3"]
}
