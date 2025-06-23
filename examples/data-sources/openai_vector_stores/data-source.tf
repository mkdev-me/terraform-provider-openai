# List all vector stores
data "openai_vector_stores" "all" {}

# List vector stores with specific ordering
data "openai_vector_stores" "recent" {
  order = "desc" # Order by created_at descending
  limit = 10
}

# List vector stores created before a specific time
data "openai_vector_stores" "older" {
  before = "vs_xyz789" # List stores created before this ID
  limit  = 20
}

# Output total vector store count
output "total_vector_stores" {
  value = length(data.openai_vector_stores.all.vector_stores)
}
