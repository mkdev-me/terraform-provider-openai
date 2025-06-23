data "openai_models" "available" {}

output "available_models" {
  value = [for model in data.openai_models.available.models : model.id if startswith(model.id, "gpt-")]
}
