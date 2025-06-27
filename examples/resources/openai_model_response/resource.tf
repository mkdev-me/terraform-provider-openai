resource "openai_model_response" "logo_prompt" {
  model = "gpt-4.1-2025-04-14"
  input = <<EOF
  Create a prompt to generate super fun logo for a new Terraform OpenAI provider.
  EOF
}