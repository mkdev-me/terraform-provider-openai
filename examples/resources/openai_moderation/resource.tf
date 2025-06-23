resource "openai_moderation" "example" {
  input = "I want to create helpful content for everyone."
  model = "text-moderation-latest"
}

output "moderation_results" {
  value = openai_moderation.example.results[0].flagged
}
