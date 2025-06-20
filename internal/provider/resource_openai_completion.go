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

// CompletionResponse represents the API response for text completions.
// It contains the generated completions and metadata about the generation process.
type CompletionResponse struct {
	ID      string             `json:"id"`      // Unique identifier for the completion
	Object  string             `json:"object"`  // Type of object (e.g., "text_completion")
	Created int                `json:"created"` // Unix timestamp of creation
	Model   string             `json:"model"`   // Model used for completion
	Choices []CompletionChoice `json:"choices"` // List of generated completions
	Usage   CompletionUsage    `json:"usage"`   // Token usage statistics
}

// CompletionChoice represents a single completion option.
// It contains the generated text and metadata about how the completion was generated.
type CompletionChoice struct {
	Text         string              `json:"text"`          // Generated completion text
	Index        int                 `json:"index"`         // Position in the list of choices
	Logprobs     *CompletionLogprobs `json:"logprobs"`      // Optional probability information
	FinishReason string              `json:"finish_reason"` // Why the completion stopped
}

// CompletionLogprobs represents probability information for a completion.
// It provides detailed token-level probability data for analyzing the completion.
type CompletionLogprobs struct {
	Tokens        []string             `json:"tokens"`         // Individual tokens in the completion
	TokenLogprobs []float64            `json:"token_logprobs"` // Log probabilities of tokens
	TopLogprobs   []map[string]float64 `json:"top_logprobs"`   // Top alternative tokens and their probabilities
	TextOffset    []int                `json:"text_offset"`    // Character offsets for tokens
}

// CompletionUsage represents token usage statistics for the request.
// It tracks the number of tokens used in both the input and output.
type CompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`     // Number of tokens in the prompt
	CompletionTokens int `json:"completion_tokens"` // Number of tokens in the completion
	TotalTokens      int `json:"total_tokens"`      // Total tokens used
}

// CompletionRequest represents the request payload for generating completions.
// It specifies the parameters that control the text generation process.
type CompletionRequest struct {
	Model            string             `json:"model"`                       // ID of the model to use
	Prompt           string             `json:"prompt"`                      // Input text to generate from
	MaxTokens        int                `json:"max_tokens,omitempty"`        // Maximum tokens to generate
	Temperature      float64            `json:"temperature,omitempty"`       // Sampling temperature (0-2)
	TopP             float64            `json:"top_p,omitempty"`             // Nucleus sampling parameter
	N                int                `json:"n,omitempty"`                 // Number of completions to generate
	Stream           bool               `json:"stream,omitempty"`            // Whether to stream responses
	Logprobs         *int               `json:"logprobs,omitempty"`          // Number of log probabilities to return
	Echo             bool               `json:"echo,omitempty"`              // Whether to include prompt in completion
	Stop             []string           `json:"stop,omitempty"`              // Sequences where completion should stop
	PresencePenalty  float64            `json:"presence_penalty,omitempty"`  // Penalty for new tokens
	FrequencyPenalty float64            `json:"frequency_penalty,omitempty"` // Penalty for frequent tokens
	BestOf           int                `json:"best_of,omitempty"`           // Number of completions to generate server-side
	LogitBias        map[string]float64 `json:"logit_bias,omitempty"`        // Modify likelihood of specific tokens
	User             string             `json:"user,omitempty"`              // Optional user identifier
	Suffix           string             `json:"suffix,omitempty"`            // Text to append to completion
}

// resourceOpenAICompletion defines the schema and CRUD operations for OpenAI text completions.
// This resource allows users to generate text completions using OpenAI's models.
// It provides comprehensive control over the generation process and supports various output formats.
func resourceOpenAICompletion() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAICompletionCreate,
		ReadContext:   resourceOpenAICompletionRead,
		DeleteContext: resourceOpenAICompletionDelete,
		Schema: map[string]*schema.Schema{
			"model": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the model to use for completion",
			},
			"prompt": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The prompt to generate completions for",
			},
			"max_tokens": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Default:     16,
				Description: "The maximum number of tokens to generate in the completion",
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
				Description: "How many completions to generate for each prompt",
			},
			"stream": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Whether to stream back partial progress",
			},
			"logprobs": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 5),
				Description:  "Include the log probabilities on the logprobs most likely tokens, between 0 and 5",
			},
			"echo": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Echo back the prompt in addition to the completion",
			},
			"stop": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    4,
				Description: "Up to 4 sequences where the API will stop generating further tokens",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"presence_penalty": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.FloatBetween(-2.0, 2.0),
				Description:  "Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far",
			},
			"frequency_penalty": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.FloatBetween(-2.0, 2.0),
				Description:  "Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far",
			},
			"best_of": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Default:     1,
				Description: "Generates best_of completions server-side and returns the 'best'",
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "A unique identifier representing your end-user",
			},
			"suffix": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The suffix that comes after a completion of inserted text",
			},
			"logit_bias": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Description: "Modify the likelihood of specified tokens appearing in the completion",
				Elem: &schema.Schema{
					Type: schema.TypeFloat,
				},
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
				Description: "The Unix timestamp (in seconds) of when the completion was created",
			},
			"completion_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the completion",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type, which is always 'text_completion'",
			},
			"model_used": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The model used for the completion",
			},
			"choices": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of completion choices the model generated for the input prompt",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"text": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The generated text",
						},
						"index": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The index of the choice in the list of choices",
						},
						"finish_reason": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The reason the model stopped generating text",
						},
						"logprobs": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Log probability information for the choice",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tokens": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The tokens generated",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"token_logprobs": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The log probabilities of the tokens",
										Elem: &schema.Schema{
											Type: schema.TypeFloat,
										},
									},
									"top_logprobs": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The top log probabilities for each token position",
										Elem: &schema.Schema{
											Type: schema.TypeMap,
											Elem: &schema.Schema{
												Type: schema.TypeFloat,
											},
										},
									},
									"text_offset": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The text offsets for each token",
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
									},
								},
							},
						},
					},
				},
			},
			"usage": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Usage statistics for the completion request",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

// resourceOpenAICompletionCreate initiates the generation of text completions.
// It processes the completion request, handles the API call, and manages the response.
// The function supports various generation options and provides control over the output format.
func resourceOpenAICompletionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the OpenAI client from the provider configuration
	client := meta.(*OpenAIClient)

	// Prepare the request with all fields
	request := &CompletionRequest{
		Model:  d.Get("model").(string),
		Prompt: d.Get("prompt").(string),
	}

	// Add the rest of the fields if they are present
	if v, ok := d.GetOk("max_tokens"); ok {
		request.MaxTokens = v.(int)
	}

	if v, ok := d.GetOk("temperature"); ok {
		request.Temperature = v.(float64)
	}

	if v, ok := d.GetOk("top_p"); ok {
		request.TopP = v.(float64)
	}

	if v, ok := d.GetOk("n"); ok {
		request.N = v.(int)
	}

	if v, ok := d.GetOk("stream"); ok {
		request.Stream = v.(bool)
	}

	if v, ok := d.GetOk("logprobs"); ok {
		logprobs := v.(int)
		request.Logprobs = &logprobs
	}

	if v, ok := d.GetOk("echo"); ok {
		request.Echo = v.(bool)
	}

	if v, ok := d.GetOk("stop"); ok {
		stopList := v.([]interface{})
		stop := make([]string, 0, len(stopList))

		for _, s := range stopList {
			stop = append(stop, s.(string))
		}

		request.Stop = stop
	}

	if v, ok := d.GetOk("presence_penalty"); ok {
		request.PresencePenalty = v.(float64)
	}

	if v, ok := d.GetOk("frequency_penalty"); ok {
		request.FrequencyPenalty = v.(float64)
	}

	if v, ok := d.GetOk("best_of"); ok {
		request.BestOf = v.(int)
	}

	if v, ok := d.GetOk("user"); ok {
		request.User = v.(string)
	}

	if v, ok := d.GetOk("suffix"); ok {
		request.Suffix = v.(string)
	}

	if v, ok := d.GetOk("logit_bias"); ok {
		logitBiasMap := v.(map[string]interface{})
		logitBias := make(map[string]float64, len(logitBiasMap))

		for k, v := range logitBiasMap {
			if fv, ok := v.(float64); ok {
				logitBias[k] = fv
			} else if sv, ok := v.(string); ok {
				if fv, err := readFloatCompletion(sv); err == nil {
					logitBias[k] = fv
				}
			}
		}

		request.LogitBias = logitBias
	}

	// Determine the API URL (considering the project_id if present)
	url := fmt.Sprintf("%s/completions", client.APIURL)

	// Create HTTP request
	reqBody, err := json.Marshal(request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing completion request: %s", err))
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
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %s, status code: %d, body: %s", err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error creating completion: %s - %s", errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse the response
	var completionResponse CompletionResponse
	if err := json.Unmarshal(respBody, &completionResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Update the state with the data from the response
	d.SetId(completionResponse.ID)
	if err := d.Set("completion_id", completionResponse.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", completionResponse.Created); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("object", completionResponse.Object); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("model_used", completionResponse.Model); err != nil {
		return diag.FromErr(err)
	}

	// Process the response options
	if len(completionResponse.Choices) > 0 {
		choices := make([]map[string]interface{}, 0, len(completionResponse.Choices))

		for _, choice := range completionResponse.Choices {
			choiceMap := map[string]interface{}{
				"text":          choice.Text,
				"index":         choice.Index,
				"finish_reason": choice.FinishReason,
			}

			// Process logprobs if they are present
			if choice.Logprobs != nil {
				logprobsMap := map[string]interface{}{
					"tokens":         choice.Logprobs.Tokens,
					"token_logprobs": choice.Logprobs.TokenLogprobs,
					"text_offset":    choice.Logprobs.TextOffset,
				}

				// Process top_logprobs
				if len(choice.Logprobs.TopLogprobs) > 0 {
					topLogprobsList := make([]map[string]interface{}, 0, len(choice.Logprobs.TopLogprobs))

					for _, topLogprobs := range choice.Logprobs.TopLogprobs {
						topLogprobsMap := make(map[string]interface{})

						for token, logprob := range topLogprobs {
							topLogprobsMap[token] = logprob
						}

						topLogprobsList = append(topLogprobsList, topLogprobsMap)
					}

					logprobsMap["top_logprobs"] = topLogprobsList
				}

				choiceMap["logprobs"] = []map[string]interface{}{logprobsMap}
			}

			choices = append(choices, choiceMap)
		}

		if err := d.Set("choices", choices); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update usage statistics
	usage := map[string]int{
		"prompt_tokens":     completionResponse.Usage.PromptTokens,
		"completion_tokens": completionResponse.Usage.CompletionTokens,
		"total_tokens":      completionResponse.Usage.TotalTokens,
	}
	if err := d.Set("usage", usage); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// readFloatCompletion is a helper function to read float values from schema data.
// It handles the conversion of interface{} values to float64, with proper error handling.
func readFloatCompletion(s string) (float64, error) {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return 0, err
	}
	return f, nil
}

// resourceOpenAICompletionRead retrieves the current state of a completion.
// It fetches the latest information about the completion and updates the Terraform state.
// Note: Completions are immutable, so this function only verifies existence.
func resourceOpenAICompletionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Completions are ephemeral and cannot be retrieved after creation
	// This function exists only to maintain the resource in the Terraform state
	return nil
}

// resourceOpenAICompletionDelete removes a completion.
// Note: Completions cannot be deleted through the API.
// This function only removes the resource from the Terraform state.
func resourceOpenAICompletionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Completions are ephemeral and cannot be deleted
	// Clear the ID to remove from Terraform state
	d.SetId("")
	return nil
}
