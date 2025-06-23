# List all assistants
data "openai_assistants" "all" {
  # Optional: Limit the number of assistants returned
  # limit = 20

  # Optional: Order by created_at
  # order = "desc"
}

# Output total number of assistants
output "assistant_count" {
  value = length(data.openai_assistants.all.assistants)
}

