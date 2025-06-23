resource "openai_embedding" "example" {
  model = "text-embedding-3-small"
  input = jsonencode(["The quick brown fox jumps over the lazy dog."])
}

output "embedding_vector" {
  value     = openai_embedding.example.embeddings[0].embedding
  sensitive = true
}
