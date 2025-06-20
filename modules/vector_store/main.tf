terraform {
  required_providers {
    openai = {
      source  = "fjcorp/openai"
      version = ">= 1.0.0"
    }
  }
}

# Vector Store resource
resource "openai_vector_store" "this" {
  name     = var.name
  file_ids = var.file_ids
  metadata = var.metadata

  # Use chunking_strategy as a block if defined
  dynamic "chunking_strategy" {
    for_each = var.chunking_strategy != null ? [var.chunking_strategy] : []
    content {
      type = chunking_strategy.value.type
    }
  }

  # Use expires_after as a block if defined
  dynamic "expires_after" {
    for_each = var.expires_after != null ? [var.expires_after] : []
    content {
      days   = lookup(expires_after.value, "days", null)
      anchor = lookup(expires_after.value, "anchor", "last_active_at")
    }
  }
}

# Vector Store File resource(s) - Only created when use_file_batches = false
resource "openai_vector_store_file" "individual" {
  for_each = var.use_file_batches ? toset([]) : toset(var.file_ids)

  vector_store_id = openai_vector_store.this.id
  file_id         = each.value

  # Use file_attributes map
  attributes = var.file_attributes

  # Use chunking_strategy as a block if defined
  dynamic "chunking_strategy" {
    for_each = var.chunking_strategy != null ? [var.chunking_strategy] : []
    content {
      type = chunking_strategy.value.type
    }
  }
}

# Vector Store File Batch resource - Only created when use_file_batches = true and file_ids is not empty
resource "openai_vector_store_file_batch" "batch" {
  count = var.use_file_batches && length(var.file_ids) > 0 ? 1 : 0

  vector_store_id = openai_vector_store.this.id
  file_ids        = var.file_ids

  # Convert file_attributes map to attributes map
  attributes = var.file_attributes

  # Use chunking_strategy as a block if defined
  dynamic "chunking_strategy" {
    for_each = var.chunking_strategy != null ? [var.chunking_strategy] : []
    content {
      type = chunking_strategy.value.type
    }
  }
} 