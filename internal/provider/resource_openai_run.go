package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// RunRequest represents the request payload for creating a run in the OpenAI API.
// It contains all the configuration parameters needed to execute an assistant run,
// including model settings, tools, and execution parameters.
type RunRequest struct {
	AssistantID   string                   `json:"assistant_id"`              // ID of the assistant to use for this run
	Model         string                   `json:"model,omitempty"`           // Optional model override for this run
	Instructions  string                   `json:"instructions,omitempty"`    // Optional instructions override for this run
	Tools         []map[string]interface{} `json:"tools,omitempty"`           // Tools the assistant can use for this run
	Metadata      map[string]interface{}   `json:"metadata,omitempty"`        // Optional metadata for the run
	Temperature   *float64                 `json:"temperature,omitempty"`     // Sampling temperature (0-2)
	MaxTokens     *int                     `json:"max_tokens,omitempty"`      // Maximum number of tokens to generate
	TopP          *float64                 `json:"top_p,omitempty"`           // Nucleus sampling parameter (0-1)
	StreamForTool *bool                    `json:"stream_for_tool,omitempty"` // Whether to stream tool outputs
}

// RunResponse represents the API response for a run.
// It contains comprehensive information about a run's execution and status,
// including timing information, configuration, and results.
type RunResponse struct {
	ID           string                   `json:"id"`                     // Unique identifier for the run
	Object       string                   `json:"object"`                 // Object type, always "thread.run"
	CreatedAt    int64                    `json:"created_at"`             // Unix timestamp when the run was created
	ThreadID     string                   `json:"thread_id"`              // ID of the thread this run belongs to
	AssistantID  string                   `json:"assistant_id"`           // ID of the assistant used for this run
	Status       string                   `json:"status"`                 // Current status of the run
	StartedAt    *int64                   `json:"started_at,omitempty"`   // Unix timestamp when the run started
	CompletedAt  *int64                   `json:"completed_at,omitempty"` // Unix timestamp when the run completed
	Model        string                   `json:"model"`                  // Model used for the run
	Instructions string                   `json:"instructions"`           // Instructions used for the run
	Tools        []map[string]interface{} `json:"tools"`                  // Tools available to the assistant
	FileIDs      []string                 `json:"file_ids"`               // Files available to the assistant
	Metadata     map[string]interface{}   `json:"metadata"`               // User-provided metadata
	Usage        *RunUsage                `json:"usage,omitempty"`        // Token usage statistics
}

// RunUsage represents token usage statistics for a run.
// It tracks the number of tokens used for prompts and completions,
// providing detailed information about resource consumption.
type RunUsage struct {
	PromptTokens     int `json:"prompt_tokens"`     // Number of tokens in the prompt
	CompletionTokens int `json:"completion_tokens"` // Number of tokens in the completion
	TotalTokens      int `json:"total_tokens"`      // Total tokens used in the run
}

// RunStepResponse represents a single step in a run's execution.
// It provides detailed information about each action taken during the run,
// including timing, status, and step-specific details.
type RunStepResponse struct {
	ID        string                 `json:"id"`         // Unique identifier for the step
	Object    string                 `json:"object"`     // Object type, always "thread.run.step"
	CreatedAt int64                  `json:"created_at"` // Unix timestamp when the step was created
	RunID     string                 `json:"run_id"`     // ID of the run this step belongs to
	Type      string                 `json:"type"`       // Type of step (e.g., "message_creation", "tool_calls")
	Status    string                 `json:"status"`     // Current status of the step
	Details   map[string]interface{} `json:"details"`    // Additional details about the step
}

// ListRunStepsResponse represents the API response for listing run steps.
// It provides a paginated list of steps in a run, with metadata for navigation.
type ListRunStepsResponse struct {
	Object  string            `json:"object"`   // Object type, always "list"
	Data    []RunStepResponse `json:"data"`     // Array of run steps
	FirstID string            `json:"first_id"` // ID of the first item in the list
	LastID  string            `json:"last_id"`  // ID of the last item in the list
	HasMore bool              `json:"has_more"` // Whether there are more items to fetch
}

// resourceOpenAIRun defines the schema and CRUD operations for OpenAI runs.
// This resource allows users to manage assistant runs through Terraform,
// providing control over execution, monitoring, and cleanup of runs.
func resourceOpenAIRun() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIRunCreate,
		ReadContext:   resourceOpenAIRunRead,
		DeleteContext: resourceOpenAIRunDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIRunImport,
		},
		CustomizeDiff: resourceOpenAIRunCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"thread_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the thread to run the assistant on.",
			},
			"assistant_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the assistant to use for the run.",
			},
			"model": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The model to use for the run. If not provided, the assistant's model will be used.",
			},
			"instructions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Override the default instructions of the assistant for the run.",
			},
			"tools": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Override the tools of the assistant for the run.",
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Description: "Metadata to associate with the run.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"temperature": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.FloatBetween(0, 2),
				Description:  "The sampling temperature to use. Higher values make the output more random, lower values make it more deterministic.",
			},
			"max_tokens": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "The maximum number of tokens to generate in the run. The default value is inf.",
			},
			"top_p": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.FloatBetween(0, 1),
				Description:  "An alternative to sampling with temperature. Defaults to 1.",
			},
			"stream_for_tool": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "Streaming for tool use is only available in the Chat Completions API.",
			},
			"completion_window": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "The maximum amount of time to wait for the run to complete, in seconds. If not provided, the run will be created but not waited for.",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type, which is always 'thread.run'.",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the run was created.",
			},
			"started_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the run was started.",
			},
			"completed_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the run was completed.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the run (queued, in_progress, completed, failed, etc.).",
			},
			"file_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The IDs of the files used in the run.",
			},
			"usage": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "Usage statistics for the run.",
			},
			"steps": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the run step.",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The object type, which is always 'thread.run.step'.",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The timestamp for when the run step was created.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the run step.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the run step.",
						},
						"details": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The details of the run step, JSON encoded.",
						},
					},
				},
				Description: "The steps of the run.",
			},
		},
	}
}

// resourceOpenAIRunCreate initiates a new assistant run.
// It processes the run configuration, starts the execution, and monitors progress.
// The function handles various run parameters and provides detailed status updates.
func resourceOpenAIRunCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Get the thread ID and assistant ID
	threadID := d.Get("thread_id").(string)
	assistantID := d.Get("assistant_id").(string)

	// Prepare the request
	request := &RunRequest{
		AssistantID: assistantID,
	}

	// Add optional fields if present
	if v, ok := d.GetOk("model"); ok {
		request.Model = v.(string)
	}

	if v, ok := d.GetOk("instructions"); ok {
		request.Instructions = v.(string)
	}

	// Process tools if present
	if toolsRaw, ok := d.GetOk("tools"); ok {
		toolsList := toolsRaw.([]interface{})
		tools := make([]map[string]interface{}, 0, len(toolsList))

		for _, toolRaw := range toolsList {
			tool := toolRaw.(map[string]interface{})
			tools = append(tools, tool)
		}

		request.Tools = tools
	}

	// Add metadata if present
	if metadataRaw, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		for k, v := range metadataRaw.(map[string]interface{}) {
			metadata[k] = v.(string)
		}
		request.Metadata = metadata
	}

	// Add temperature if present
	if v, ok := d.GetOk("temperature"); ok {
		temp := v.(float64)
		request.Temperature = &temp
	}

	// Add max_tokens if present
	if v, ok := d.GetOk("max_tokens"); ok {
		maxTokens := v.(int)
		request.MaxTokens = &maxTokens
	}

	// Add top_p if present
	if v, ok := d.GetOk("top_p"); ok {
		topP := v.(float64)
		request.TopP = &topP
	}

	// Add stream_for_tool if present
	if v, ok := d.GetOk("stream_for_tool"); ok {
		streamForTool := v.(bool)
		request.StreamForTool = &streamForTool
	}

	// Convert request to JSON
	reqBody, err := json.Marshal(request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing run request: %s", err))
	}

	// Prepare HTTP request
	url := fmt.Sprintf("%s/threads/%s/runs", client.APIURL, threadID)
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

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s", err))
		}
		return diag.FromErr(fmt.Errorf("error creating run: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var runResponse RunResponse
	if err := json.Unmarshal(respBody, &runResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Save ID and other data to state
	d.SetId(runResponse.ID)
	if err := d.Set("created_at", runResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", runResponse.Status); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("thread_id", runResponse.ThreadID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("assistant_id", runResponse.AssistantID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("model", runResponse.Model); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("instructions", runResponse.Instructions); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("file_ids", runResponse.FileIDs); err != nil {
		return diag.FromErr(err)
	}

	if runResponse.StartedAt != nil {
		if err := d.Set("started_at", *runResponse.StartedAt); err != nil {
			return diag.FromErr(err)
		}
	}

	if runResponse.CompletedAt != nil {
		if err := d.Set("completed_at", *runResponse.CompletedAt); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("tools", runResponse.Tools); err != nil {
		return diag.FromErr(err)
	}

	if runResponse.Metadata != nil {
		metadata := make(map[string]interface{})
		for k, v := range runResponse.Metadata {
			metadata[k] = v
		}
		if err := d.Set("metadata", metadata); err != nil {
			return diag.FromErr(err)
		}
	}

	if runResponse.Usage != nil {
		usage := map[string]interface{}{
			"prompt_tokens":     runResponse.Usage.PromptTokens,
			"completion_tokens": runResponse.Usage.CompletionTokens,
			"total_tokens":      runResponse.Usage.TotalTokens,
		}
		if err := d.Set("usage", usage); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("object", runResponse.Object); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", runResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceOpenAIRunRead retrieves the current state of a run.
// It fetches the latest run information from the API and updates the Terraform state
// with current status, results, and usage statistics.
func resourceOpenAIRunRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*OpenAIClient)

	// Get necessary parameters
	runID := d.Id()
	threadID := d.Get("thread_id").(string)

	// Prepare HTTP request
	url := fmt.Sprintf("%s/threads/%s/runs/%s", client.APIURL, threadID, runID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %v", err))
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	// Add Organization ID if present
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Perform request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %v", err))
	}
	defer resp.Body.Close()

	// Verify if run exists
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %v", err))
	}

	// Verify if there was an error
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %v, status code: %d, body: %s",
				err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error reading run: %s - %s",
			errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse response
	var runResponse RunResponse
	if err := json.Unmarshal(respBody, &runResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	// Update run attributes
	if err := d.Set("object", runResponse.Object); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", runResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("assistant_id", runResponse.AssistantID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("thread_id", runResponse.ThreadID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", runResponse.Status); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("model", runResponse.Model); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("instructions", runResponse.Instructions); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("file_ids", runResponse.FileIDs); err != nil {
		return diag.FromErr(err)
	}

	if runResponse.Usage != nil {
		usage := map[string]interface{}{
			"prompt_tokens":     runResponse.Usage.PromptTokens,
			"completion_tokens": runResponse.Usage.CompletionTokens,
			"total_tokens":      runResponse.Usage.TotalTokens,
		}
		if err := d.Set("usage", usage); err != nil {
			return diag.FromErr(err)
		}
	}

	// Get run steps if run has completed
	if runResponse.Status == "completed" {
		stepsURL := fmt.Sprintf("%s/threads/%s/runs/%s/steps",
			client.APIURL, threadID, runResponse.ID)

		stepsReq, err := http.NewRequestWithContext(ctx, "GET", stepsURL, nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating steps request: %v", err))
		}

		stepsReq.Header.Set("Authorization", "Bearer "+client.APIKey)
		stepsReq.Header.Set("OpenAI-Beta", "assistants=v2")

		if client.OrganizationID != "" {
			stepsReq.Header.Set("OpenAI-Organization", client.OrganizationID)
		}

		stepsResp, err := http.DefaultClient.Do(stepsReq)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error making steps request: %v", err))
		}
		defer stepsResp.Body.Close()

		stepsBody, err := io.ReadAll(stepsResp.Body)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error reading steps response: %v", err))
		}

		if stepsResp.StatusCode != http.StatusOK {
			var errorResponse ErrorResponse
			if err := json.Unmarshal(stepsBody, &errorResponse); err != nil {
				return diag.FromErr(fmt.Errorf("error parsing steps error response: %v, status code: %d, body: %s",
					err, stepsResp.StatusCode, string(stepsBody)))
			}
			return diag.FromErr(fmt.Errorf("error getting run steps: %s - %s",
				errorResponse.Error.Type, errorResponse.Error.Message))
		}

		var stepsResponse ListRunStepsResponse
		if err := json.Unmarshal(stepsBody, &stepsResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing steps response: %v", err))
		}

		// Convert steps to format for schema
		steps := make([]map[string]interface{}, len(stepsResponse.Data))
		for i, step := range stepsResponse.Data {
			stepMap := map[string]interface{}{
				"id":         step.ID,
				"object":     step.Object,
				"created_at": step.CreatedAt,
				"type":       step.Type,
				"status":     step.Status,
			}

			// Convert details to JSON
			if len(step.Details) > 0 {
				detailsJSON, err := json.Marshal(step.Details)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error serializing step details: %v", err))
				}
				stepMap["details"] = string(detailsJSON)
			} else {
				stepMap["details"] = "{}"
			}

			steps[i] = stepMap
		}

		if err := d.Set("steps", steps); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIRunDelete removes a run.
// It handles the cleanup of run resources and associated data,
// ensuring proper termination of any ongoing processes.
func resourceOpenAIRunDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*OpenAIClient)

	// Get necessary parameters
	runID := d.Id()
	threadID := d.Get("thread_id").(string)

	// Prepare HTTP request
	url := fmt.Sprintf("%s/threads/%s/runs/%s", client.APIURL, threadID, runID)
	req, err := http.NewRequestWithContext(ctx, "POST", url+"/cancel", nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %v", err))
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	// Add Organization ID if present
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Realizar la petición
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %v", err))
	}
	defer resp.Body.Close()

	// Si el run no existe o ya está completo, simplemente limpiar el estado
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Verificar si hubo un error
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %v, status code: %d, body: %s",
				err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error cancelling run: %s - %s",
			errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Limpiar el estado
	d.SetId("")

	return diag.Diagnostics{}
}

// resourceOpenAIRunImport handles importing an existing OpenAI run into Terraform state.
// It uses a composite ID in the format "thread_id:run_id" to identify the run.
func resourceOpenAIRunImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Split the import ID on colon to get thread_id and run_id
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return nil, fmt.Errorf("invalid import format, expected 'thread_id:run_id'")
	}

	threadID := idParts[0]
	runID := idParts[1]

	// Set thread_id in the resource data
	if err := d.Set("thread_id", threadID); err != nil {
		return nil, fmt.Errorf("error setting thread_id: %s", err)
	}

	// Get the OpenAI client
	client := m.(*OpenAIClient)

	// Construct the API URL
	url := fmt.Sprintf("%s/threads/%s/runs/%s", client.APIURL, threadID, runID)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))
	req.Header.Add("OpenAI-Beta", "assistants=v2")
	if client.OrganizationID != "" {
		req.Header.Add("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check for error responses
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error - status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse the full response as a map to capture all fields
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &rawResponse); err != nil {
		return nil, fmt.Errorf("error parsing response body as map: %w", err)
	}

	// Set the resource ID (run_id)
	d.SetId(runID)

	// Map standard fields
	setIfPresent(d, rawResponse, "assistant_id", "assistant_id")
	setIfPresent(d, rawResponse, "model", "model")
	setIfPresent(d, rawResponse, "instructions", "instructions")
	setIfPresent(d, rawResponse, "object", "object")
	setIfPresent(d, rawResponse, "created_at", "created_at")
	setIfPresent(d, rawResponse, "status", "status")
	setIfPresent(d, rawResponse, "started_at", "started_at")
	setIfPresent(d, rawResponse, "completed_at", "completed_at")

	// Handle file_ids
	if fileIDs, ok := rawResponse["file_ids"].([]interface{}); ok {
		stringIDs := make([]string, len(fileIDs))
		for i, id := range fileIDs {
			if strID, ok := id.(string); ok {
				stringIDs[i] = strID
			}
		}
		_ = d.Set("file_ids", stringIDs)
	} else {
		_ = d.Set("file_ids", []string{})
	}

	// Handle tools
	if tools, ok := rawResponse["tools"].([]interface{}); ok && len(tools) > 0 {
		_ = d.Set("tools", tools)
	}

	// Handle metadata
	if metadata, ok := rawResponse["metadata"].(map[string]interface{}); ok {
		flatMetadata := make(map[string]string)
		for k, v := range metadata {
			if strValue, ok := v.(string); ok {
				flatMetadata[k] = strValue
			} else {
				// Convert non-string values to JSON string
				jsonValue, err := json.Marshal(v)
				if err == nil {
					flatMetadata[k] = string(jsonValue)
				}
			}
		}
		_ = d.Set("metadata", flatMetadata)
	}

	// Handle temperature
	if temp, ok := rawResponse["temperature"].(float64); ok {
		_ = d.Set("temperature", temp)
	}

	// Handle top_p
	if topP, ok := rawResponse["top_p"].(float64); ok {
		_ = d.Set("top_p", topP)
	}

	// Handle max_tokens
	if maxTokens, ok := rawResponse["max_completion_tokens"].(float64); ok {
		_ = d.Set("max_tokens", int(maxTokens))
	}

	// Handle stream_for_tool
	if streamForTool, ok := rawResponse["stream_for_tool"].(bool); ok {
		_ = d.Set("stream_for_tool", streamForTool)
	}

	// Handle steps
	if stepsRaw, ok := rawResponse["steps"].([]interface{}); ok && len(stepsRaw) > 0 {
		steps := make([]map[string]interface{}, 0, len(stepsRaw))
		for _, stepRaw := range stepsRaw {
			if stepMap, ok := stepRaw.(map[string]interface{}); ok {
				step := make(map[string]interface{})

				// Copy standard step fields
				if id, ok := stepMap["id"].(string); ok {
					step["id"] = id
				}
				if obj, ok := stepMap["object"].(string); ok {
					step["object"] = obj
				}
				if createdAt, ok := stepMap["created_at"].(float64); ok {
					step["created_at"] = int(createdAt)
				}
				if status, ok := stepMap["status"].(string); ok {
					step["status"] = status
				}
				if stepType, ok := stepMap["type"].(string); ok {
					step["type"] = stepType
				}

				// Handle details separately as it might be complex
				if details, ok := stepMap["details"].(map[string]interface{}); ok {
					detailsJSON, err := json.Marshal(details)
					if err == nil {
						step["details"] = string(detailsJSON)
					} else {
						step["details"] = "{}"
					}
				} else {
					step["details"] = "{}"
				}

				steps = append(steps, step)
			}
		}
		_ = d.Set("steps", steps)
	}

	// Handle usage
	if usageRaw, ok := rawResponse["usage"].(map[string]interface{}); ok {
		usage := make(map[string]interface{})

		if promptTokens, ok := usageRaw["prompt_tokens"].(float64); ok {
			usage["prompt_tokens"] = int(promptTokens)
		}
		if completionTokens, ok := usageRaw["completion_tokens"].(float64); ok {
			usage["completion_tokens"] = int(completionTokens)
		}
		if totalTokens, ok := usageRaw["total_tokens"].(float64); ok {
			usage["total_tokens"] = int(totalTokens)
		}

		_ = d.Set("usage", usage)
	}

	return []*schema.ResourceData{d}, nil
}

// Helper function to set a field from the API response if it exists
func setIfPresent(d *schema.ResourceData, response map[string]interface{}, apiField, schemaField string) {
	if value, ok := response[apiField]; ok {
		_ = d.Set(schemaField, value)
	}
}

// resourceOpenAIRunCustomizeDiff is a custom diff function to prevent Terraform from recreating resources due to API-provided fields
func resourceOpenAIRunCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	// Don't run this logic during resource creation
	if d.Id() == "" {
		return nil
	}

	// Suppress changes to these fields when they are not set in the config but exist in the state
	// This prevents recreation of the resource when the API provides default values
	fieldsToSuppress := []string{
		"model",
		"instructions",
		"tools",
		"steps",
		"file_ids",
		"status",
		"object",
		"usage",
		"started_at",
		"completed_at",
	}

	for _, field := range fieldsToSuppress {
		// If the field is in old but not in new (not specified in config),
		// suppress the diff by setting new = old
		oldValue, newValue := d.GetChange(field)
		if oldValue != nil && !reflect.DeepEqual(oldValue, reflect.Zero(reflect.TypeOf(oldValue)).Interface()) {
			// If new value is empty/nil/default and old value exists
			if newValue == nil || reflect.DeepEqual(newValue, reflect.Zero(reflect.TypeOf(newValue)).Interface()) {
				_ = d.SetNew(field, oldValue)
			}
		}
	}

	return nil
}
