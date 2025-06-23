package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// AssistantResponse represents the API response for an OpenAI assistant.
// It contains all the fields returned by the OpenAI API when creating or retrieving an assistant.
type AssistantResponse struct {
	ID           string                 `json:"id"`
	Object       string                 `json:"object"`
	CreatedAt    int                    `json:"created_at"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Model        string                 `json:"model"`
	Instructions string                 `json:"instructions"`
	Tools        []AssistantTool        `json:"tools"`
	FileIDs      []string               `json:"file_ids"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// AssistantTool represents a tool that can be used by an assistant.
// Tools can be of different types such as code interpreter, retrieval, function, or file search.
type AssistantTool struct {
	Type     string                 `json:"type"`
	Function *AssistantToolFunction `json:"function,omitempty"`
}

// AssistantToolFunction represents a function definition for an assistant tool.
// It contains the name, description, and parameters of the function in JSON Schema format.
type AssistantToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
}

// AssistantCreateRequest represents the payload for creating an assistant in the OpenAI API.
// It contains all the fields that can be set when creating a new assistant.
type AssistantCreateRequest struct {
	Model        string                 `json:"model"`
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Instructions string                 `json:"instructions,omitempty"`
	Tools        []AssistantTool        `json:"tools,omitempty"`
	FileIDs      []string               `json:"file_ids,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// resourceOpenAIAssistant defines the schema and CRUD operations for OpenAI assistants.
// This resource allows users to create, read, update, and delete assistants through the OpenAI API.
func resourceOpenAIAssistant() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIAssistantCreate,
		ReadContext:   resourceOpenAIAssistantRead,
		UpdateContext: resourceOpenAIAssistantUpdate,
		DeleteContext: resourceOpenAIAssistantDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the assistant",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the assistant",
			},
			"model": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the model to use for this assistant",
			},
			"instructions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The system instructions that the assistant uses",
			},
			"tools": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"code_interpreter", "retrieval", "function", "file_search"}, false),
							Description:  "The type of tool being defined: code_interpreter, retrieval, function, or file_search",
						},
						"function": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the function",
									},
									"description": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The description of the function",
									},
									"parameters": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The parameters of the function in JSON Schema format (as a string)",
									},
								},
							},
						},
					},
				},
			},
			"file_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of file IDs attached to this assistant",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Metadata for the assistant",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the assistant was created",
			},
		},
	}
}

// resourceOpenAIAssistantCreate handles the creation of a new OpenAI assistant.
// It uses the configuration from Terraform to make an API request to OpenAI.
// The function processes all the provided fields and creates a new assistant with the specified configuration.
func resourceOpenAIAssistantCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Prepare the request
	createRequest := &AssistantCreateRequest{
		Model: d.Get("model").(string),
	}

	// Add name if present
	if name, ok := d.GetOk("name"); ok {
		createRequest.Name = name.(string)
	}

	// Add description if present
	if description, ok := d.GetOk("description"); ok {
		createRequest.Description = description.(string)
	}

	// Add instructions if present
	if instructions, ok := d.GetOk("instructions"); ok {
		createRequest.Instructions = instructions.(string)
	}

	// Add tools if present
	if toolsRaw, ok := d.GetOk("tools"); ok {
		toolsList := toolsRaw.([]interface{})
		tools := make([]AssistantTool, 0, len(toolsList))

		for _, toolRaw := range toolsList {
			toolMap := toolRaw.(map[string]interface{})

			tool := AssistantTool{
				Type: toolMap["type"].(string),
			}

			// If type is "function", add function details
			if tool.Type == "function" && toolMap["function"] != nil {
				functionList := toolMap["function"].([]interface{})
				if len(functionList) > 0 {
					functionMap := functionList[0].(map[string]interface{})

					// Convert parameters to JSON
					parametersStr := functionMap["parameters"].(string)
					parametersJSON := json.RawMessage(parametersStr)

					// Create the function
					tool.Function = &AssistantToolFunction{
						Name:       functionMap["name"].(string),
						Parameters: parametersJSON,
					}

					// Add description if present
					if description, ok := functionMap["description"]; ok && description.(string) != "" {
						tool.Function.Description = description.(string)
					}
				}
			}

			tools = append(tools, tool)
		}

		createRequest.Tools = tools
	}

	// Add file_ids if present
	if fileIDsRaw, ok := d.GetOk("file_ids"); ok {
		fileIDsList := fileIDsRaw.([]interface{})
		fileIDs := make([]string, 0, len(fileIDsList))

		for _, id := range fileIDsList {
			fileIDs = append(fileIDs, id.(string))
		}

		createRequest.FileIDs = fileIDs
	}

	// Add metadata if present
	if metadataRaw, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		for k, v := range metadataRaw.(map[string]interface{}) {
			metadata[k] = v.(string)
		}
		createRequest.Metadata = metadata
	}

	// Convert request to JSON
	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing assistant request: %s", err))
	}

	// Prepare HTTP request
	url := fmt.Sprintf("%s/assistants", client.APIURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Perform the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Verify if there was an error
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s", err))
		}
		return diag.FromErr(fmt.Errorf("error creating assistant: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var assistantResponse AssistantResponse
	if err := json.Unmarshal(respBody, &assistantResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Save the ID and other data in the state
	d.SetId(assistantResponse.ID)
	if err := d.Set("created_at", assistantResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("model", assistantResponse.Model); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", assistantResponse.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", assistantResponse.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("instructions", assistantResponse.Instructions); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("file_ids", assistantResponse.FileIDs); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceOpenAIAssistantRead fetches the current state of an OpenAI assistant
// and updates the Terraform state to match.
func resourceOpenAIAssistantRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the assistant ID
	assistantID := d.Id()
	if assistantID == "" {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Prepare the HTTP request
	url := fmt.Sprintf("%s/assistants/%s", client.APIURL, assistantID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// If the assistant doesn't exist, remove it from state
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check if there was an error
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s", err))
		}
		return diag.FromErr(fmt.Errorf("error reading assistant: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var assistantResponse AssistantResponse
	if err := json.Unmarshal(respBody, &assistantResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update the state
	if err := d.Set("model", assistantResponse.Model); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", assistantResponse.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", assistantResponse.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("instructions", assistantResponse.Instructions); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", assistantResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("file_ids", assistantResponse.FileIDs); err != nil {
		return diag.FromErr(err)
	}

	// Process tools if present
	if len(assistantResponse.Tools) > 0 {
		tools := make([]map[string]interface{}, 0, len(assistantResponse.Tools))

		for _, tool := range assistantResponse.Tools {
			toolMap := map[string]interface{}{
				"type": tool.Type,
			}

			// If type is "function", process the function details
			if tool.Type == "function" && tool.Function != nil {
				// Normalize JSON parameters to prevent whitespace differences
				var parametersObj interface{}
				if err := json.Unmarshal(tool.Function.Parameters, &parametersObj); err == nil {
					// Re-marshal with standard formatting
					normalizedJSON, err := json.Marshal(parametersObj)
					if err == nil {
						function := map[string]interface{}{
							"name":       tool.Function.Name,
							"parameters": string(normalizedJSON),
						}

						if tool.Function.Description != "" {
							function["description"] = tool.Function.Description
						}

						toolMap["function"] = []interface{}{function}
					} else {
						// Fallback to original if normalization fails
						function := map[string]interface{}{
							"name":       tool.Function.Name,
							"parameters": string(tool.Function.Parameters),
						}

						if tool.Function.Description != "" {
							function["description"] = tool.Function.Description
						}

						toolMap["function"] = []interface{}{function}
					}
				} else {
					// Fallback to original if JSON parsing fails
					function := map[string]interface{}{
						"name":       tool.Function.Name,
						"parameters": string(tool.Function.Parameters),
					}

					if tool.Function.Description != "" {
						function["description"] = tool.Function.Description
					}

					toolMap["function"] = []interface{}{function}
				}
			}

			tools = append(tools, toolMap)
		}

		if err := d.Set("tools", tools); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update metadata if present
	if len(assistantResponse.Metadata) > 0 {
		metadata := make(map[string]string)
		for k, v := range assistantResponse.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		if err := d.Set("metadata", metadata); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIAssistantUpdate applies any changes to an existing OpenAI assistant.
// After updating, it reads the assistant state to ensure the Terraform state is current.
func resourceOpenAIAssistantUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the assistant ID
	assistantID := d.Id()

	// Prepare the request
	updateRequest := make(map[string]interface{})

	// Add fields that have changed
	if d.HasChange("model") {
		updateRequest["model"] = d.Get("model").(string)
	}

	if d.HasChange("name") {
		updateRequest["name"] = d.Get("name").(string)
	}

	if d.HasChange("description") {
		updateRequest["description"] = d.Get("description").(string)
	}

	if d.HasChange("instructions") {
		updateRequest["instructions"] = d.Get("instructions").(string)
	}

	// Process tools if they have changed
	if d.HasChange("tools") {
		toolsRaw := d.Get("tools").([]interface{})
		tools := make([]AssistantTool, 0, len(toolsRaw))

		for _, toolRaw := range toolsRaw {
			toolMap := toolRaw.(map[string]interface{})

			tool := AssistantTool{
				Type: toolMap["type"].(string),
			}

			// If type is "function", add the function details
			if tool.Type == "function" && toolMap["function"] != nil {
				functionList := toolMap["function"].([]interface{})
				if len(functionList) > 0 {
					functionMap := functionList[0].(map[string]interface{})

					// Convert parameters to JSON
					parametersStr := functionMap["parameters"].(string)
					parametersJSON := json.RawMessage(parametersStr)

					// Create the function
					tool.Function = &AssistantToolFunction{
						Name:       functionMap["name"].(string),
						Parameters: parametersJSON,
					}

					// Add description if present
					if description, ok := functionMap["description"]; ok && description.(string) != "" {
						tool.Function.Description = description.(string)
					}
				}
			}

			tools = append(tools, tool)
		}

		updateRequest["tools"] = tools
	}

	// Process file_ids if they have changed
	if d.HasChange("file_ids") {
		fileIDsRaw := d.Get("file_ids").([]interface{})
		fileIDs := make([]string, 0, len(fileIDsRaw))

		for _, id := range fileIDsRaw {
			fileIDs = append(fileIDs, id.(string))
		}

		updateRequest["file_ids"] = fileIDs
	}

	// Process metadata if it has changed
	if d.HasChange("metadata") {
		metadataRaw := d.Get("metadata").(map[string]interface{})
		metadata := make(map[string]interface{})
		for k, v := range metadataRaw {
			metadata[k] = v.(string)
		}
		updateRequest["metadata"] = metadata
	}

	// Convert the request to JSON
	reqBody, err := json.Marshal(updateRequest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing update request: %s", err))
	}

	// Prepare the HTTP request
	url := fmt.Sprintf("%s/assistants/%s", client.APIURL, assistantID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
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
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check if there was an error
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s", err))
		}
		return diag.FromErr(fmt.Errorf("error updating assistant: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Call Read to update the state
	return resourceOpenAIAssistantRead(ctx, d, m)
}

// resourceOpenAIAssistantDelete removes an OpenAI assistant.
func resourceOpenAIAssistantDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the assistant ID
	assistantID := d.Id()

	// Prepare the HTTP request
	url := fmt.Sprintf("%s/assistants/%s", client.APIURL, assistantID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// If the assistant doesn't exist, it's not an error
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %s", err))
	}

	// Check if there was an error
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s", err))
		}
		return diag.FromErr(fmt.Errorf("error deleting assistant: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Clear the ID from state
	d.SetId("")

	return diag.Diagnostics{}
}
