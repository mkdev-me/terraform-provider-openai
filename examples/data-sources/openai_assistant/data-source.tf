# Fetch a specific assistant by ID
data "openai_assistant" "code_reviewer" {
  assistant_id = "asst_8sPATZ7dVbBL1m1Yve97j2BM"
}

# Output the assistant ID
output "assistant_id" {
  value = data.openai_assistant.code_reviewer.id
}
