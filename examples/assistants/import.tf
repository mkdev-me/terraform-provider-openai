# Example of importing an existing assistant
# This resource definition is used for importing an existing assistant
# Replace "asst_abc123" with your actual assistant ID when importing

resource "openai_assistant" "imported_assistant" {
  # These values will be overwritten during import
  # They are required just to create a valid resource block
  name  = "Placeholder Name"
  model = "gpt-4o"

  # The actual values will be fetched from OpenAI when importing
  # Run: terraform import openai_assistant.imported_assistant asst_abc123

  # After import, you can modify properties as needed

  # After importing, you can use the resource in your configuration
  # like any other resource. For example, you could add or change metadata:
  # metadata = {
  #   "imported_by" = "terraform",
  #   "import_date" = "2023-06-01"
  # }

  # Or you could modify instructions:
  # instructions = "Updated instructions after import"

  # To prevent accidental modification of sensitive properties during apply:
  lifecycle {
    ignore_changes = [
      # List properties you don't want to modify here
      # For example, if you don't want to change the model:
      # model,
    ]
  }
} 