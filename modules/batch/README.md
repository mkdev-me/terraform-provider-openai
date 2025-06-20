# OpenAI Batch Processing Module (Legacy/Deprecated)

> **Note**: This module is now deprecated. For all projects, we recommend using the `openai_batch` resource directly instead of this module.

This module previously simulated batch processing jobs in OpenAI before the official batch API was fully supported. Now that the `openai_batch` resource is fully implemented, this module only exists for backward compatibility with older code.

## Usage Comparison

### Preferred: Direct resource usage (recommended for all code)

```hcl
resource "openai_batch" "example" {
  input_file_id     = openai_file.input_file.id
  endpoint          = "/v1/embeddings"
  model             = "text-embedding-ada-002"
  completion_window = "24h"
  
  # Optional metadata
  metadata = {
    environment = "production"
    project     = "document-embeddings" 
  }
}
```

### Legacy: Module usage (for backward compatibility only)

```hcl
module "batch" {
  source            = "../../modules/batch"
  input_file_id     = openai_file.input_file.id
  endpoint          = "/v1/embeddings"
  model             = "text-embedding-ada-002"
  completion_window = "24h"
}
```

## Migration Path

If you're using this module in existing code, we recommend migrating to the direct resource approach. The module outputs approximately match the attributes of the resource, but the actual batch API provides more features and real-time status:

```hcl
# Before
output "batch_id" {
  value = module.batch.batch_id
}

# After
output "batch_id" {
  value = openai_batch.example.id
}
```

## Module Variables

| Variable           | Type        | Description                                       | Default  |
|--------------------|-------------|---------------------------------------------------|----------|
| input_file_id      | string      | ID of the batch input file                        | Required |
| endpoint           | string      | API endpoint for batch processing                 | Required |
| model              | string      | Model to use for the batch                        | Required |
| completion_window  | string      | Time window for batch processing                  | "24h"    |
| project_id         | string      | OpenAI project ID                                 | ""       |
| metadata           | map(string) | Key-value pairs to attach to the batch            | {}       |

## Module Outputs

| Output        | Type   | Description                                  |
|---------------|--------|----------------------------------------------|
| batch_id      | string | Simulated batch job ID                       |
| status        | string | Always "in_progress" for simulated jobs      |
| created_at    | string | Static timestamp for job creation            |
| expires_at    | string | Static timestamp for job expiration          |
| output_file_id| string | Simulated output file ID                     |
| error         | string | Always null for simulated jobs               |

## How This Module Works

This module uses `openai_chat_completion` to create a simulated batch job. It generates deterministic IDs and static timestamps to provide consistent outputs between runs.

The simulation provides:
1. A deterministic batch ID
2. Static timestamps for creation and expiration
3. A simulated output file ID
4. A fixed "in_progress" status

## Limitations of the Module

Since this is a simulation:
1. Batch jobs don't actually process files
2. Status is always reported as "in_progress"
3. The simulation doesn't provide actual results
4. You cannot monitor progress or retrieve actual outputs

## Advantages of Using the Resource Instead

The official `openai_batch` resource provides:
1. Real batch processing through the OpenAI API
2. Accurate status tracking (validating, processing, completed, etc.)
3. Actual output file IDs when processing completes
4. Error handling and reporting
5. Metadata support for better organization
6. Request count statistics

## When to Use This Module

Only use this module when:
- You have existing code using this module that you're not ready to migrate
- You need to maintain compatibility with configurations written for earlier versions

For all other use cases, use the `openai_batch` resource directly. 