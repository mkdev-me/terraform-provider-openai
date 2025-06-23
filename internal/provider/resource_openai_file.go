package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// FileResponse represents the API response for an OpenAI file.
// It contains all the fields returned by the OpenAI API when creating or retrieving a file.
// This structure provides comprehensive information about the file's status, purpose, and metadata.
type FileResponse struct {
	ID        string `json:"id"`         // Unique identifier for the file
	Object    string `json:"object"`     // Type of object (e.g., "file")
	Bytes     int    `json:"bytes"`      // Size of the file in bytes
	CreatedAt int    `json:"created_at"` // Unix timestamp of file creation
	Filename  string `json:"filename"`   // Original name of the uploaded file
	Purpose   string `json:"purpose"`    // Intended use of the file (e.g., "fine-tune", "assistants")
}

// ErrorResponse represents an error response from the OpenAI API.
// It contains detailed information about any errors that occur during API operations,
// providing structured error information for proper error handling and debugging.
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"` // Human-readable error message
		Type    string `json:"type"`    // Type of error (e.g., "invalid_request_error")
		Code    string `json:"code"`    // Error code for programmatic handling
	} `json:"error"`
}

// resourceOpenAIFile defines the schema and CRUD operations for OpenAI files.
// This resource allows users to upload, read, and delete files through the OpenAI API.
// It supports various file purposes including fine-tuning, assistants, vision, and batch operations.
// The resource provides comprehensive file management capabilities with proper validation and error handling.
func resourceOpenAIFile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIFileCreate,
		ReadContext:   resourceOpenAIFileRead,
		DeleteContext: resourceOpenAIFileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIFileImport,
		},
		Schema: map[string]*schema.Schema{
			"file": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Path to the file to upload. Required for creation, ignored during import.",
			},
			"purpose": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"fine-tune", "assistants", "vision", "batch"}, false),
				Description:  "The purpose of the file. Can be 'fine-tune', 'assistants', 'vision', or 'batch'. Required for creation, computed for import.",
				Computed:     true,
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The project ID to associate this file with (for Terraform reference only, not sent to OpenAI API)",
			},
			"filename": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the file",
			},
			"bytes": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size of the file in bytes",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the file was created",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identifier of the file",
			},
		},
	}
}

// resourceOpenAIFileCreate handles the creation and upload of a new file to OpenAI.
// It processes the file upload, validates the file purpose, and manages the upload process.
// The function supports various file types and purposes, with appropriate error handling.
func resourceOpenAIFileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get parameters
	filePath, filePathOk := d.GetOk("file")
	if !filePathOk {
		return diag.FromErr(fmt.Errorf("file path is required for creating a new file"))
	}

	purpose := d.Get("purpose").(string)

	// Check if there's a project_id (we save it for state but don't send it to the API)
	var projectID string
	if v, ok := d.GetOk("project_id"); ok {
		projectID = v.(string)
	}

	// Read the file
	fileContent, err := os.ReadFile(filePath.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading file %s: %s", filePath, err))
	}

	// Prepare the OpenAI API request
	url := fmt.Sprintf("%s/v1/files", client.APIURL)

	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/files", client.APIURL)
	}

	// Create form-data for upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath.(string)))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating form file: %s", err))
	}
	_, err = part.Write(fileContent)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error writing file content: %s", err))
	}

	// Add the purpose
	err = writer.WriteField("purpose", purpose)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error writing purpose field: %s", err))
	}

	// Note: We no longer send project_id to the API as it's not supported
	// We keep project_id in the state for future use or reference

	err = writer.Close()
	if err != nil {
		return diag.FromErr(fmt.Errorf("error closing writer: %s", err))
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Add headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody)))
		}
		return diag.FromErr(fmt.Errorf("API error: %s - %s (%s)", errorResponse.Error.Type, errorResponse.Error.Message, errorResponse.Error.Code))
	}

	// Parse JSON response
	var fileResponse FileResponse
	err = json.Unmarshal(responseBody, &fileResponse)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Save data to state
	d.SetId(fileResponse.ID)
	if err := d.Set("filename", fileResponse.Filename); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("bytes", fileResponse.Bytes); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", fileResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("purpose", fileResponse.Purpose); err != nil {
		return diag.FromErr(err)
	}

	// If project_id was provided, save it to state as well
	// Note: This is for reference only as the API doesn't actually use this field
	if projectID != "" {
		if err := d.Set("project_id", projectID); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIFileRead retrieves the current state of an OpenAI file.
// It fetches the latest information about the file from OpenAI's API and updates
// the Terraform state with the current file metadata and status.
func resourceOpenAIFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the file ID
	fileID := d.Id()
	if fileID == "" {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Create URL to get file information
	url := fmt.Sprintf("%s/v1/files/%s", client.APIURL, fileID)

	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/files/%s", client.APIURL, fileID)
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// If file doesn't exist (404), mark as deleted
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody)))
		}
		return diag.FromErr(fmt.Errorf("API error: %s - %s (%s)", errorResponse.Error.Type, errorResponse.Error.Message, errorResponse.Error.Code))
	}

	// Parse JSON response
	var fileResponse FileResponse
	err = json.Unmarshal(responseBody, &fileResponse)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update state
	if err := d.Set("filename", fileResponse.Filename); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("bytes", fileResponse.Bytes); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", fileResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("purpose", fileResponse.Purpose); err != nil {
		return diag.FromErr(err)
	}

	// Keep project_id if it was set
	// The API doesn't return project_id, so we have to keep it from the state

	return diag.Diagnostics{}
}

// resourceOpenAIFileDelete removes a file from OpenAI.
// It handles the deletion of the file through OpenAI's API and ensures proper cleanup.
// Note: Some file types may have specific deletion requirements or restrictions.
func resourceOpenAIFileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the file ID
	fileID := d.Id()
	if fileID == "" {
		return diag.Diagnostics{}
	}

	// Create the URL to delete the file
	url := fmt.Sprintf("%s/v1/files/%s", client.APIURL, fileID)

	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/files/%s", client.APIURL, fileID)
	}

	// Create the HTTP request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the call
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// Si el archivo no existe, no es un error
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check if there were errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		if err != nil {
			return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody)))
		}
		return diag.FromErr(fmt.Errorf("API error: %s - %s (%s)", errorResponse.Error.Type, errorResponse.Error.Message, errorResponse.Error.Code))
	}

	// Remove the ID to mark that the resource no longer exists
	d.SetId("")

	return diag.Diagnostics{}
}

// resourceOpenAIFileImport handles the import of an existing file.
// It retrieves the file details directly from the OpenAI API and sets them in the Terraform state.
// If the file cannot be found or accessed, it returns an error rather than setting placeholder values.
func resourceOpenAIFileImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Get the OpenAI client
	client, err := GetOpenAIClient(m)
	if err != nil {
		return nil, fmt.Errorf("error getting OpenAI client: %s", err)
	}

	// Get the file ID from the import command
	fileID := d.Id()
	if fileID == "" {
		return nil, fmt.Errorf("error: no file ID provided for import")
	}

	// Construct the API URL for retrieving the file
	url := fmt.Sprintf("%s/v1/files/%s", client.APIURL, fileID)

	// If the APIURL already contains /v1, adjust the URL construction
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/files/%s", client.APIURL, fileID)
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %s", err)
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %s", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
			return nil, fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody))
		}
		return nil, fmt.Errorf("API error: %s - %s (%s)", errorResponse.Error.Type, errorResponse.Error.Message, errorResponse.Error.Code)
	}

	// Parse JSON response
	var fileResponse FileResponse
	if err := json.Unmarshal(responseBody, &fileResponse); err != nil {
		return nil, fmt.Errorf("error parsing response: %s", err)
	}

	// Set the ID first
	d.SetId(fileResponse.ID)

	// Use a more common path pattern that's likely to match existing configurations
	// This helps prevent unnecessary resource replacement after import
	filePath := fmt.Sprintf("./data/%s", fileResponse.Filename)

	// Check if the file exists at common locations
	_, errData := os.Stat(filePath)
	if errData != nil {
		// Try without the data subdirectory
		altPath := fmt.Sprintf("./%s", fileResponse.Filename)
		_, errAlt := os.Stat(altPath)
		if errAlt == nil {
			// Use the file path if found
			filePath = altPath
		} else {
			// Fallback to a placeholder that's less likely to cause replacement
			filePath = fmt.Sprintf("./data/%s", fileResponse.Filename)
		}
	}

	if err := d.Set("file", filePath); err != nil {
		return nil, fmt.Errorf("error setting file path: %s", err)
	}

	// Set all the computed fields from the API response
	if err := d.Set("filename", fileResponse.Filename); err != nil {
		return nil, fmt.Errorf("error setting filename: %s", err)
	}
	if err := d.Set("bytes", fileResponse.Bytes); err != nil {
		return nil, fmt.Errorf("error setting bytes: %s", err)
	}
	if err := d.Set("created_at", fileResponse.CreatedAt); err != nil {
		return nil, fmt.Errorf("error setting created_at: %s", err)
	}
	if err := d.Set("purpose", fileResponse.Purpose); err != nil {
		return nil, fmt.Errorf("error setting purpose: %s", err)
	}

	// Return the resourceData with properly populated fields
	return []*schema.ResourceData{d}, nil
}
