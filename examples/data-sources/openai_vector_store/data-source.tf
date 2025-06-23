# Note: This example assumes you have a vector store ID
# In practice, you would either:
# 1. Create a vector store first, or
# 2. Use a known existing vector store ID from your OpenAI account

variable "vector_store_id" {
  description = "ID of an existing vector store"
  type        = string
  default     = "vs-example" # Replace with actual vector store ID
}

data "openai_vector_store" "knowledge_base" {
  id = var.vector_store_id
}

output "vector_store_id" {
  value = data.openai_vector_store.knowledge_base.id
}
