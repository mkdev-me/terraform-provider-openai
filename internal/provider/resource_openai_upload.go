package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fjcorp/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// UploadResponse represents the API response for an OpenAI upload.
// It contains information about the upload status and properties.
type UploadResponse struct {
	ID        string `json:"id"`         // Unique identifier for the upload
	Object    string `json:"object"`     // Type of object (e.g., "upload")
	Purpose   string `json:"purpose"`    // Intended use of the upload
	Filename  string `json:"filename"`   // Name of the uploaded file
	Bytes     int    `json:"bytes"`      // Size of the file in bytes
	CreatedAt int    `json:"created_at"` // Unix timestamp of upload creation
	Status    string `json:"status"`     // Current status of the upload (e.g., "pending", "completed")
}

// UploadCreateRequest represents the request payload for creating a new upload.
// It contains the required parameters for initiating an upload.
type UploadCreateRequest struct {
	Purpose  string `json:"purpose"`   // Intended use of the upload (e.g., "fine-tune")
	Filename string `json:"filename"`  // Name of the file being uploaded
	Bytes    int    `json:"bytes"`     // Size of the file in bytes
	MimeType string `json:"mime_type"` // MIME type of the file
}

// resourceOpenAIUpload defines the schema and CRUD operations for OpenAI uploads.
// This resource allows users to manage file uploads for various purposes including fine-tuning.
func resourceOpenAIUpload() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIUploadCreate,
		ReadContext:   resourceOpenAIUploadRead,
		DeleteContext: resourceOpenAIUploadDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"purpose": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"fine-tune", "assistants", "vision", "batch", "user_data", "evals"}, false),
				Description:  "The intended purpose of the uploaded file. Can be 'fine-tune', 'assistants', 'vision', 'batch', 'user_data', or 'evals'.",
			},
			"filename": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of the file to upload.",
			},
			"bytes": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Description: "The number of bytes in the file being uploaded.",
			},
			"mime_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The MIME type of the file. Must fall within supported MIME types for your file purpose.",
			},
			"file": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The path to the file to upload. Required for creation, not needed for import.",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The project ID to associate this upload with (for Terraform reference only, not sent to OpenAI API).",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the upload was created.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the upload (e.g., 'pending', 'completed').",
			},
		},
	}
}

// resourceOpenAIUploadCreate handles the creation of a new OpenAI upload.
// It sends the upload request to the OpenAI API and processes the response.
func resourceOpenAIUploadCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.OpenAIClient)

	// Get the values from the resource data
	purpose := d.Get("purpose").(string)

	// Check if we have a file path specified
	filePath, filePathOk := d.GetOk("file")
	if !filePathOk {
		return diag.FromErr(fmt.Errorf("file path is required for creating a new upload"))
	}

	// Use existing filename or get from file path
	var filename string
	if fn, ok := d.GetOk("filename"); ok {
		filename = fn.(string)
	} else {
		// Extract filename from file path
		parts := strings.Split(filePath.(string), "/")
		filename = parts[len(parts)-1]
	}

	// Use existing bytes or get file size
	var fileBytes int
	var err error
	if b, ok := d.GetOk("bytes"); ok {
		fileBytes = b.(int)
	} else {
		// Get file size from the file
		fileInfo, err := os.Stat(filePath.(string))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error getting file size: %v", err))
		}
		fileBytes = int(fileInfo.Size())
	}

	// Use existing mime_type or determine from file
	var mimeType string
	if mt, ok := d.GetOk("mime_type"); ok {
		mimeType = mt.(string)
	} else {
		// Determine mime type based on file extension
		ext := filepath.Ext(filename)
		switch strings.ToLower(ext) {
		case ".jsonl":
			mimeType = "application/jsonl"
		case ".json":
			mimeType = "application/json"
		case ".txt":
			mimeType = "text/plain"
		case ".csv":
			mimeType = "text/csv"
		case ".pdf":
			mimeType = "application/pdf"
		case ".md":
			mimeType = "text/markdown"
		default:
			mimeType = "application/octet-stream"
		}
	}

	// TODO: Implement file upload using multipart form

	// For now, we're using the API as before
	// Prepare the request body
	uploadReq := UploadCreateRequest{
		Purpose:  purpose,
		Filename: filename,
		Bytes:    fileBytes,
		MimeType: mimeType,
	}

	// Convert the request to JSON
	reqBody, err := json.Marshal(uploadReq)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshalling request: %v", err))
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/uploads", c.APIURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %v", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	if c.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", c.OrganizationID)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error sending request: %v", err))
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %v", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return diag.FromErr(fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error.Message))
		}
		return diag.FromErr(fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody)))
	}

	// Parse the response
	var uploadResp UploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	// Set the Terraform resource data
	d.SetId(uploadResp.ID)
	if err := d.Set("status", uploadResp.Status); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set status: %v", err))
	}
	if err := d.Set("created_at", uploadResp.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
	}

	return nil
}

// resourceOpenAIUploadRead retrieves information about an existing OpenAI upload.
// It queries the OpenAI API for the current state of the upload and updates the resource data.
func resourceOpenAIUploadRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.OpenAIClient)

	// Create the HTTP request to get upload details
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/uploads/%s", c.APIURL, d.Id()), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %v", err))
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	if c.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", c.OrganizationID)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error sending request: %v", err))
	}
	defer resp.Body.Close()

	// Handle not found (upload deleted or expired)
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %v", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return diag.FromErr(fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error.Message))
		}
		return diag.FromErr(fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody)))
	}

	// Parse the response
	var uploadResp UploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	// Set the computed attributes
	if err := d.Set("status", uploadResp.Status); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set status: %v", err))
	}
	if err := d.Set("created_at", uploadResp.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set created_at: %v", err))
	}

	// Set the purpose attribute
	if err := d.Set("purpose", uploadResp.Purpose); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set purpose: %v", err))
	}

	// Update the resource data
	if err := d.Set("filename", uploadResp.Filename); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set filename: %v", err))
	}
	if err := d.Set("bytes", uploadResp.Bytes); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set bytes: %v", err))
	}

	return nil
}

// resourceOpenAIUploadDelete handles the deletion of an OpenAI upload.
// It sends a delete request to the OpenAI API and processes the response.
func resourceOpenAIUploadDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.OpenAIClient)

	// Create the HTTP request to delete the upload
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/v1/uploads/%s", c.APIURL, d.Id()), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %v", err))
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	if c.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", c.OrganizationID)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error sending request: %v", err))
	}
	defer resp.Body.Close()

	// Handle not found (already deleted)
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %v", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return diag.FromErr(fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error.Message))
		}
		return diag.FromErr(fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody)))
	}

	// Clear the ID to mark the resource as deleted
	d.SetId("")

	return nil
}
