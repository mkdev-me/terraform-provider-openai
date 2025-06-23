data "openai_model" "gpt4" {
  model_id = "gpt-4o-mini"
}

output "model_details" {
  value = {
    id       = data.openai_model.gpt4.id
    created  = data.openai_model.gpt4.created
    owned_by = data.openai_model.gpt4.owned_by
  }
}
