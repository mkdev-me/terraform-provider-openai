package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// BatchResponse represents the API response for batch operations.
// It contains information about the batch job, including its status, timing, and results.
type BatchResponse struct {
	ID               string                 `json:"id"`                // Unique identifier for the batch job
	Object           string                 `json:"object"`            // Type of object (e.g., "batch")
	Endpoint         string                 `json:"endpoint"`          // Endpoint used for this batch
	Status           string                 `json:"status"`            // Current status of the batch job
	InputFileID      string                 `json:"input_file_id"`     // ID of the input file
	CompletionWindow string                 `json:"completion_window"` // Time window for completion
	OutputFileID     string                 `json:"output_file_id"`    // ID of the output file (if available)
	ErrorFileID      string                 `json:"error_file_id"`     // ID of the error file (if available)
	CreatedAt        int                    `json:"created_at"`        // Unix timestamp when the job was created
	InProgressAt     *int                   `json:"in_progress_at"`    // When processing started
	ExpiresAt        int                    `json:"expires_at"`        // Unix timestamp when the job expires
	FinalizingAt     *int                   `json:"finalizing_at"`     // When finalizing started
	CompletedAt      *int                   `json:"completed_at"`      // When processing completed
	FailedAt         *int                   `json:"failed_at"`         // When processing failed
	ExpiredAt        *int                   `json:"expired_at"`        // When the job expired
	CancellingAt     *int                   `json:"cancelling_at"`     // When cancellation started
	CancelledAt      *int                   `json:"cancelled_at"`      // When the job was cancelled
	RequestCounts    map[string]int         `json:"request_counts"`    // Statistics about request processing
	Errors           interface{}            `json:"errors,omitempty"`  // Errors that occurred (if any)
	Metadata         map[string]interface{} `json:"metadata"`          // Additional custom data
}

// BatchError represents an error that occurred during batch processing.
// It contains details about the error, including its code and message.
type BatchError struct {
	Code    string `json:"code"`            // Error code identifying the type of error
	Message string `json:"message"`         // Human-readable error message
	Param   string `json:"param,omitempty"` // Parameter that caused the error (if applicable)
}

// BatchCreateRequest represents the request payload for creating a new batch job.
// It specifies the input file, endpoint, and completion window for the batch operation.
type BatchCreateRequest struct {
	InputFileID      string                 `json:"input_file_id"`      // ID of the input file to process
	Endpoint         string                 `json:"endpoint"`           // API endpoint to use for processing
	CompletionWindow string                 `json:"completion_window"`  // Time window for job completion
	Metadata         map[string]interface{} `json:"metadata,omitempty"` // Optional metadata for the batch
}

// resourceOpenAIBatch defines the schema and CRUD operations for OpenAI batch jobs.
// This resource allows users to create and manage batch processing jobs for various OpenAI operations.
// It supports monitoring job status and retrieving results once processing is complete.
func resourceOpenAIBatch() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIBatchCreate,
		ReadContext:   resourceOpenAIBatchRead,
		DeleteContext: resourceOpenAIBatchDelete,
		Importer: &schema.ResourceImporter{
			StateContext: importBatchState,
		},
		Schema: map[string]*schema.Schema{
			"input_file_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the file containing the inputs for the batch",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The ID of the project to use for this batch. If not specified, the default project will be used.",
			},
			"output_file_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the output file",
			},
			"completion_window": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "24h",
				Description: "The maximum time to wait for the batch to complete (e.g., '24h')",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The endpoint to use for the batch request (e.g., 'chat/completions')",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the batch was created",
			},
			"expires_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the batch expires",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the batch (validating, validation_failed, processing, processing_failed, completed)",
			},
			"error": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Information about the error that occurred during processing, if any",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of 16 key-value pairs that can be attached to the batch object",
			},
		},
	}
}

// resourceOpenAIBatchCreate handles the creation of a new OpenAI batch job.
// It submits a batch processing request to OpenAI's API and initializes the job.
// The function supports various batch operations and provides options for job configuration.
func resourceOpenAIBatchCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.Errorf("error getting OpenAI client: %s", err)
	}

	// Obtener parámetros
	inputFileID := d.Get("input_file_id").(string)
	endpoint := d.Get("endpoint").(string)

	// Ensure endpoint includes /v1 prefix for the API
	// The API expects: /v1/chat/completions, /v1/completions, /v1/embeddings, etc.
	apiEndpoint := endpoint
	if !strings.HasPrefix(apiEndpoint, "/v1") {
		apiEndpoint = "/v1" + endpoint
	}

	completionWindow := d.Get("completion_window").(string)

	// Project ID will be included in the client request automatically

	// Preparar la petición
	createRequest := &BatchCreateRequest{
		InputFileID:      inputFileID,
		Endpoint:         apiEndpoint, // Use the endpoint with /v1 prefix
		CompletionWindow: completionWindow,
	}

	// Add metadata if provided
	if v, ok := d.GetOk("metadata"); ok {
		metadataMap := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			metadataMap[k] = v
		}
		createRequest.Metadata = metadataMap
	}

	// Make the API request
	responseBytes, err := client.DoRequest("POST", "/v1/batches", createRequest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating batch: %s", err))
	}

	// Parse the response
	var batchResponse BatchResponse
	if err := json.Unmarshal(responseBytes, &batchResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Guardar el ID y otros datos en el estado
	d.SetId(batchResponse.ID)
	if err := d.Set("created_at", batchResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("expires_at", batchResponse.ExpiresAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", batchResponse.Status); err != nil {
		return diag.FromErr(err)
	}

	// Si hay un error, guardarlo
	if batchResponse.ErrorFileID != "" {
		if err := d.Set("error", batchResponse.ErrorFileID); err != nil {
			return diag.FromErr(err)
		}
	}

	// Si hay un output_file_id, guardarlo
	if batchResponse.OutputFileID != "" {
		if err := d.Set("output_file_id", batchResponse.OutputFileID); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update the endpoint in state to ensure consistency
	// Store the normalized endpoint (without /v1 prefix) in the state
	// This helps prevent unnecessary diffs in the Terraform plan
	if err := d.Set("endpoint", endpoint); err != nil {
		return diag.FromErr(err)
	}

	// Store metadata if present
	if batchResponse.Metadata != nil && len(batchResponse.Metadata) > 0 {
		metadataMap := make(map[string]string)
		for k, v := range batchResponse.Metadata {
			switch val := v.(type) {
			case string:
				metadataMap[k] = val
			default:
				// Convert non-string values to JSON string
				jsonBytes, err := json.Marshal(val)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error marshaling metadata value: %s", err))
				}
				metadataMap[k] = string(jsonBytes)
			}
		}
		if err := d.Set("metadata", metadataMap); err != nil {
			return diag.FromErr(err)
		}
	}

	// Do not wait for completion by default - batch jobs can take hours or days
	// The user can check the status with a data source or output

	return diag.Diagnostics{}
}

// resourceOpenAIBatchRead retrieves the current state of an OpenAI batch job.
// It fetches the latest status and results from OpenAI's API and updates the Terraform state.
// This function is used to monitor job progress and retrieve results.
func resourceOpenAIBatchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.Errorf("error getting OpenAI client: %s", err)
	}

	// Obtener el ID del batch
	batchID := d.Id()
	if batchID == "" {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Project ID will be handled by the client automatically via headers

	// Make API call
	responseBytes, err := client.DoRequest("GET", fmt.Sprintf("/v1/batches/%s", batchID), nil)
	if err != nil {
		// If the error contains "404" or "not found", the batch might have been deleted
		if strings.Contains(strings.ToLower(err.Error()), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("error retrieving batch: %s", err))
	}

	// Parse the response
	var batchResponse BatchResponse
	if err := json.Unmarshal(responseBytes, &batchResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update the state
	d.Set("input_file_id", batchResponse.InputFileID)

	// Endpoint should be normalized to remove the /v1 prefix if present
	endpoint := batchResponse.Endpoint
	if strings.HasPrefix(endpoint, "/v1") {
		endpoint = strings.TrimPrefix(endpoint, "/v1")
	}
	d.Set("endpoint", endpoint)

	d.Set("completion_window", batchResponse.CompletionWindow)
	d.Set("created_at", batchResponse.CreatedAt)
	d.Set("expires_at", batchResponse.ExpiresAt)
	d.Set("status", batchResponse.Status)

	// Si hay un output_file_id, guardarlo
	if batchResponse.OutputFileID != "" {
		d.Set("output_file_id", batchResponse.OutputFileID)
	}

	// Si hay un error_file_id, guardarlo como error
	if batchResponse.ErrorFileID != "" {
		d.Set("error", batchResponse.ErrorFileID)
	}

	// Store metadata if present
	if batchResponse.Metadata != nil && len(batchResponse.Metadata) > 0 {
		metadataMap := make(map[string]string)
		for k, v := range batchResponse.Metadata {
			switch val := v.(type) {
			case string:
				metadataMap[k] = val
			default:
				// Convert non-string values to JSON string
				jsonBytes, err := json.Marshal(val)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error marshaling metadata value: %s", err))
				}
				metadataMap[k] = string(jsonBytes)
			}
		}
		if err := d.Set("metadata", metadataMap); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIBatchDelete handles the deletion of an OpenAI batch job.
// Since batch jobs cannot be deleted from the API, this function verifies the job exists and cleans up the Terraform state.
func resourceOpenAIBatchDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.Errorf("error getting OpenAI client: %s", err)
	}

	// Get the batch ID
	batchID := d.Id()
	if batchID == "" {
		return diag.Diagnostics{}
	}

	// Project ID will be handled by the client automatically via headers

	// Make a request to cancel the batch
	var responseBytes []byte
	var requestErr error

	// Try to cancel the batch
	responseBytes, requestErr = client.DoRequest("POST", fmt.Sprintf("/v1/batches/%s/cancel", batchID), nil)

	// If we got a 404, the batch is already gone, so we can remove it from the state
	if requestErr != nil {
		if strings.Contains(strings.ToLower(requestErr.Error()), "404") || strings.Contains(strings.ToLower(requestErr.Error()), "not found") {
			d.SetId("")
			return diag.Diagnostics{}
		}
		// For other errors, just log them but still remove from state
		log.Printf("[WARN] Error cancelling batch: %s", requestErr)
	} else {
		log.Printf("[INFO] Batch cancel response: %s", string(responseBytes))
	}

	// Remove the batch from the Terraform state
	d.SetId("")
	return diag.Diagnostics{}
}

// importBatchState handles the import of a batch resource
// The ID is expected to be the batch ID from OpenAI
func importBatchState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return nil, fmt.Errorf("error getting OpenAI client: %s", err)
	}

	batchID := d.Id()
	if batchID == "" {
		return nil, fmt.Errorf("error: no batch ID provided for import")
	}

	// Make API call to get batch details
	responseBytes, err := client.DoRequest("GET", fmt.Sprintf("/v1/batches/%s", batchID), nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving batch: %s", err)
	}

	// Parse batch response
	var batchResponse BatchResponse
	if err := json.Unmarshal(responseBytes, &batchResponse); err != nil {
		return nil, fmt.Errorf("error parsing response: %s", err)
	}

	// Explicitly set the ID first to ensure it's preserved
	d.SetId(batchResponse.ID)

	// Set the required fields in the resource data
	if err := d.Set("input_file_id", batchResponse.InputFileID); err != nil {
		return nil, fmt.Errorf("error setting input_file_id: %s", err)
	}

	// Endpoint should be normalized to remove the /v1 prefix if present
	endpoint := batchResponse.Endpoint
	if strings.HasPrefix(endpoint, "/v1") {
		endpoint = strings.TrimPrefix(endpoint, "/v1")
	}
	if err := d.Set("endpoint", endpoint); err != nil {
		return nil, fmt.Errorf("error setting endpoint: %s", err)
	}

	// Set completion window if available
	if batchResponse.CompletionWindow != "" {
		if err := d.Set("completion_window", batchResponse.CompletionWindow); err != nil {
			return nil, fmt.Errorf("error setting completion_window: %s", err)
		}
	}

	// Set output_file_id if available
	if batchResponse.OutputFileID != "" {
		if err := d.Set("output_file_id", batchResponse.OutputFileID); err != nil {
			return nil, fmt.Errorf("error setting output_file_id: %s", err)
		}
	}

	// Set created_at
	if err := d.Set("created_at", batchResponse.CreatedAt); err != nil {
		return nil, fmt.Errorf("error setting created_at: %s", err)
	}

	// Set expires_at
	if err := d.Set("expires_at", batchResponse.ExpiresAt); err != nil {
		return nil, fmt.Errorf("error setting expires_at: %s", err)
	}

	// Set status
	if err := d.Set("status", batchResponse.Status); err != nil {
		return nil, fmt.Errorf("error setting status: %s", err)
	}

	// Set error_file_id if available
	if batchResponse.ErrorFileID != "" {
		if err := d.Set("error", batchResponse.ErrorFileID); err != nil {
			return nil, fmt.Errorf("error setting error field: %s", err)
		}
	}

	// Set metadata if available
	if batchResponse.Metadata != nil && len(batchResponse.Metadata) > 0 {
		// Convert the metadata to a string map as required by schema
		metadataMap := make(map[string]interface{})
		for k, v := range batchResponse.Metadata {
			switch val := v.(type) {
			case string:
				metadataMap[k] = val
			default:
				// Convert non-string values to JSON string
				jsonBytes, err := json.Marshal(val)
				if err != nil {
					return nil, fmt.Errorf("error marshaling metadata value: %s", err)
				}
				metadataMap[k] = string(jsonBytes)
			}
		}
		if err := d.Set("metadata", metadataMap); err != nil {
			return nil, fmt.Errorf("error setting metadata: %s", err)
		}
	}

	return []*schema.ResourceData{d}, nil
}
