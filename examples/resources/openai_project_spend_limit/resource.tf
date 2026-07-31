resource "openai_project_spend_limit" "production" {
  project_id       = "proj_abc123"
  threshold_amount = 10000
}
