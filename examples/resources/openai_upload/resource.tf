# Example: Managing large file uploads using OpenAI's Upload API
# The Upload API is used for files larger than 512 MB that need to be uploaded in parts

# Create an upload session for a large file
resource "openai_upload" "large_dataset" {
  # The name of the file being uploaded
  filename = "training_dataset_10gb.jsonl"

  # Purpose of the file
  purpose = "fine-tune"

  # Total size of the file in bytes
  bytes = 10737418240 # 10 GB

  # MIME type of the file
  mime_type = "application/jsonl"
}

# Example: Upload session for a large video file
resource "openai_upload" "video_upload" {
  filename  = "product_demo_4k.mp4"
  purpose   = "assistants"
  bytes     = 5368709120 # 5 GB
  mime_type = "video/mp4"
}

# Example: Upload session for a large audio file
resource "openai_upload" "audiobook_upload" {
  filename  = "complete_audiobook.m4a"
  purpose   = "assistants"
  bytes     = 2147483648 # 2 GB
  mime_type = "audio/mp4"
}

# Example: Upload session for batch processing data
resource "openai_upload" "batch_data" {
  filename  = "batch_requests_large.jsonl"
  purpose   = "batch"
  bytes     = 1073741824 # 1 GB
  mime_type = "application/jsonl"
}

# After creating the upload session, you'll need to upload parts
# This is typically done programmatically outside of Terraform
# as it involves streaming file chunks

# Output upload ID
output "upload_id" {
  value = openai_upload.large_dataset.id
}
