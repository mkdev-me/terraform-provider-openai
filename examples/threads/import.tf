# Example of importing an existing thread
# This resource definition is used for importing an existing thread
# Replace "thread_abc123" with your actual thread ID when importing

resource "openai_thread" "imported_thread" {
  # No configuration is required for the import placeholder
  # The actual values will be fetched from OpenAI when importing

  # Run: terraform import openai_thread.imported_thread thread_abc123

  # After import, you can modify properties like metadata as needed
  # For example:
  # metadata = {
  #   "imported_by" = "terraform",
  #   "import_date" = "2023-06-01"
  # }

  # Note: You cannot add messages to an existing thread through Terraform
  # as the messages field is ForceNew and would create a new thread.
  # Instead, use the OpenAI API or openai_message resource to add messages.

  # To prevent accidental changes during apply:
  lifecycle {
    ignore_changes = [
      # List properties you don't want to modify here
      # For example, if you never want to modify metadata:
      # metadata,
    ]
  }
} 