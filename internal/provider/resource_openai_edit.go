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

// EditResponse represents the API response for text edits.
// It contains the edited text, model information, and usage statistics.
type EditResponse struct {
	ID      string       `json:"id"`      // Unique identifier for the edit
	Object  string       `json:"object"`  // Type of object (e.g., "edit")
	Created int          `json:"created"` // Unix timestamp when the edit was created
	Model   string       `json:"model"`   // Model used for the edit
	Choices []EditChoice `json:"choices"` // List of possible edits
	Usage   EditUsage    `json:"usage"`   // Token usage statistics
}

// EditChoice represents a single edit option from the model.
// It contains the edited text and its position in the list of choices.
type EditChoice struct {
	Text  string `json:"text"`  // The edited text
	Index int    `json:"index"` // Position of this choice in the list
}

// EditUsage represents token usage statistics for the edit request.
// It tracks the number of tokens used in the input and completion.
type EditUsage struct {
	PromptTokens     int `json:"prompt_tokens"`     // Number of tokens in the input
	CompletionTokens int `json:"completion_tokens"` // Number of tokens in the completion
	TotalTokens      int `json:"total_tokens"`      // Total number of tokens used
}

// EditRequest represents the request payload for creating a text edit.
// It specifies the model, input text, instruction, and various parameters to control the edit.
type EditRequest struct {
	Model       string  `json:"model"`                 // ID of the model to use
	Input       string  `json:"input,omitempty"`       // Text to be edited
	Instruction string  `json:"instruction"`           // Instructions for how to edit the text
	Temperature float64 `json:"temperature,omitempty"` // Sampling temperature
	TopP        float64 `json:"top_p,omitempty"`       // Nucleus sampling parameter
	N           int     `json:"n,omitempty"`           // Number of edits to generate
}

// resourceOpenAIEdit defines the schema and CRUD operations for OpenAI text edits.
// This resource allows users to edit text using OpenAI's language models.
// It supports various models and provides options for controlling the edit behavior.
func resourceOpenAIEdit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIEditCreate,
		ReadContext:   resourceOpenAIEditRead,
		DeleteContext: resourceOpenAIEditDelete,
		Schema: map[string]*schema.Schema{
			"model": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the model to use for the edit",
			},
			"input": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "The input text to edit",
			},
			"instruction": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The instruction that tells the model how to edit the input",
			},
			"temperature": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				Default:      1.0,
				ValidateFunc: validation.FloatBetween(0.0, 2.0),
				Description:  "Sampling temperature between 0 and 2. Higher values make output more random, lower values make it more deterministic",
			},
			"top_p": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				Default:      1.0,
				ValidateFunc: validation.FloatBetween(0.0, 1.0),
				Description:  "Nuclear sampling: consider the results of the tokens with top_p probability mass. Range from 0 to 1",
			},
			"n": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Default:     1,
				Description: "How many edits to generate for the input and instruction",
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The project to use for this request",
			},
			// Response fields
			"created": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The Unix timestamp (in seconds) of when the edit was created",
			},
			"edit_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the edit",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type, which is always 'edit'",
			},
			"model_used": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The model used for the edit",
			},
			"choices": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of edit choices the model generated",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"text": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The edited text",
						},
						"index": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The index of the choice in the list of choices",
						},
					},
				},
			},
			"usage": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Usage statistics for the edit request",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

// resourceOpenAIEditCreate handles the creation of a new OpenAI text edit.
// It sends the request to OpenAI's API and processes the response.
// The function supports various edit options and provides control over the editing process.
func resourceOpenAIEditCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the OpenAI client
	client := meta.(*OpenAIClient)

	// Prepare the request with all fields
	request := &EditRequest{
		Model:       d.Get("model").(string),
		Instruction: d.Get("instruction").(string),
	}

	// Add input if present
	if input, ok := d.GetOk("input"); ok {
		request.Input = input.(string)
	}

	// Add the rest of the fields if present
	if v, ok := d.GetOk("temperature"); ok {
		request.Temperature = v.(float64)
	}

	if v, ok := d.GetOk("top_p"); ok {
		request.TopP = v.(float64)
	}

	if v, ok := d.GetOk("n"); ok {
		request.N = v.(int)
	}

	// Determine the API URL (considering project_id if present)
	url := fmt.Sprintf("%s/edits", client.APIURL)

	// Create HTTP request
	reqBody, err := json.Marshal(request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing edit request: %s", err))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	// Add Organization ID if present
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Add Project ID if present
	if projectID, ok := d.GetOk("project_id"); ok {
		req.Header.Set("OpenAI-Project", projectID.(string))
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
			return diag.FromErr(fmt.Errorf("error parsing error response: %s, status code: %d, body: %s", err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error creating edit: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var editResponse EditResponse
	if err := json.Unmarshal(respBody, &editResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update the state with response data
	d.SetId(editResponse.ID)
	if err := d.Set("edit_id", editResponse.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", editResponse.Created); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("object", editResponse.Object); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("model_used", editResponse.Model); err != nil {
		return diag.FromErr(err)
	}

	// Process the response options
	if len(editResponse.Choices) > 0 {
		choices := make([]map[string]interface{}, 0, len(editResponse.Choices))

		for _, choice := range editResponse.Choices {
			choiceMap := map[string]interface{}{
				"text":  choice.Text,
				"index": choice.Index,
			}

			choices = append(choices, choiceMap)
		}

		if err := d.Set("choices", choices); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update the usage statistics
	usage := map[string]int{
		"prompt_tokens":     editResponse.Usage.PromptTokens,
		"completion_tokens": editResponse.Usage.CompletionTokens,
		"total_tokens":      editResponse.Usage.TotalTokens,
	}
	if err := d.Set("usage", usage); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceOpenAIEditRead retrieves the current state of an OpenAI text edit.
// It verifies that the edit exists and updates the Terraform state.
// Note: OpenAI edits are immutable, so this function only verifies existence.
func resourceOpenAIEditRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Edits are ephemeral and cannot be retrieved after creation
	// This function is basically a no-op, but we preserve the data we already have in state

	// If there is no ID, it means the resource does not exist
	if d.Id() == "" {
		return diag.Diagnostics{}
	}

	return diag.Diagnostics{}
}

// resourceOpenAIEditDelete removes an OpenAI text edit.
// Note: OpenAI edits are immutable and cannot be deleted through the API.
// This function only removes the resource from the Terraform state.
func resourceOpenAIEditDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Edits are ephemeral and cannot be deleted
	// Simply clear the ID from the state
	d.SetId("")
	return diag.Diagnostics{}
}
