# Example: Uploading file parts for large file uploads
# Upload parts are used with the Upload API to send large files in chunks
# Note: This is typically done programmatically, but shown here for completeness

# First, create an upload session
resource "openai_upload" "large_file" {
  filename  = "massive_dataset.jsonl"
  purpose   = "fine-tune"
  bytes     = 2147483648 # 2 GB
  mime_type = "application/jsonl"
}

# Upload the first part of the file
# In practice, you'd read file chunks programmatically
resource "openai_upload_part" "part_1" {
  # The upload session ID
  upload_id = openai_upload.large_file.id

  # Part number (starts from 1)
  part_number = 1

  # The actual data for this part (base64 encoded)
  # In real usage, this would be read from file chunks
  data = base64encode("First chunk of file data...")

  # Size of this part in bytes
  size = 67108864 # 64 MB
}

# Upload subsequent parts
resource "openai_upload_part" "part_2" {
  upload_id   = openai_upload.large_file.id
  part_number = 2
  data        = base64encode("Second chunk of file data...")
  size        = 67108864
}

resource "openai_upload_part" "part_3" {
  upload_id   = openai_upload.large_file.id
  part_number = 3
  data        = base64encode("Third chunk of file data...")
  size        = 67108864
}

# Example of the last part (might be smaller)
resource "openai_upload_part" "final_part" {
  upload_id   = openai_upload.large_file.id
  part_number = 32 # Assuming 32 parts total
  data        = base64encode("Final chunk of file data...")
  size        = 33554432 # 32 MB (last part can be smaller)
}

# After all parts are uploaded, complete the upload
resource "openai_upload_complete" "finalize" {
  upload_id = openai_upload.large_file.id

  # List all uploaded parts with their ETags
  parts = [
    {
      part_number = openai_upload_part.part_1.part_number
      etag        = openai_upload_part.part_1.etag
    },
    {
      part_number = openai_upload_part.part_2.part_number
      etag        = openai_upload_part.part_2.etag
    },
    {
      part_number = openai_upload_part.part_3.part_number
      etag        = openai_upload_part.part_3.etag
    },
    # ... include all parts ...
    {
      part_number = openai_upload_part.final_part.part_number
      etag        = openai_upload_part.final_part.etag
    }
  ]
}

# Output part ETag
output "part_1_etag" {
  value = openai_upload_part.part_1.etag
}

# Note: In production, you would typically:
# 1. Create the upload session with Terraform
# 2. Use a script or application to read the file in chunks
# 3. Upload each chunk using the OpenAI API directly
# 4. Complete the upload when all parts are uploaded

# Example script structure (not Terraform):
# ```python
# import openai
# 
# upload_id = terraform_output['upload_id']
# file_path = "massive_dataset.jsonl"
# chunk_size = 64 * 1024 * 1024  # 64 MB
# 
# with open(file_path, 'rb') as f:
#     part_number = 1
#     while chunk := f.read(chunk_size):
#         response = openai.uploads.parts.create(
#             upload_id=upload_id,
#             data=chunk,
#             part_number=part_number
#         )
#         part_number += 1
# 
# # Complete the upload
# openai.uploads.complete(upload_id=upload_id, parts=parts)
# ```
