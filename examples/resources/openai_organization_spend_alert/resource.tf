resource "openai_organization_spend_alert" "finance" {
  threshold_amount = 50000

  notification_channel {
    recipients     = ["finance@example.com"]
    subject_prefix = "OpenAI organization spend alert"
  }
}
