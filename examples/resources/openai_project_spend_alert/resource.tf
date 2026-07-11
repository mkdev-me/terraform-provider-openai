resource "openai_project_spend_alert" "monthly" {
  project_id       = "proj_abc123"
  threshold_amount = 100000

  notification_channel {
    recipients     = ["finance@example.com"]
    subject_prefix = "OpenAI spend alert"
  }
}
