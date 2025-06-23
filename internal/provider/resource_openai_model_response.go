package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// resourceOpenAIModelResponse returns a schema.Resource for OpenAI model responses.
// This resource allows users to generate text completions from OpenAI models.
func resourceOpenAIModelResponse() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIModelResponseCreate,
		ReadContext:   resourceOpenAIModelResponseRead,
		UpdateContext: resourceOpenAIModelResponseUpdate, // Explicit update function
		DeleteContext: resourceOpenAIModelResponseDelete,
		CustomizeDiff: resourceOpenAIModelResponseCustomDiff,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIModelResponseImport,
		},
		Schema: map[string]*schema.Schema{
			"input": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true, // Mark as computed to handle imported resources
				Description:  "The input text to the model",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"model": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true, // Mark as computed to handle imported resources
				Description:  "ID of the model to use (e.g., 'gpt-4o', 'gpt-4-turbo').",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"max_output_tokens": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum number of tokens to generate.",
			},
			"temperature": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Default:      0.7,
				ValidateFunc: validation.FloatBetween(0.0, 2.0),
				Description:  "Sampling temperature between 0 and 2. Higher values mean more randomness.",
			},
			"top_p": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Computed:     true, // Mark as computed to handle imported resources where top_p is present
				ValidateFunc: validation.FloatBetween(0.0, 1.0),
				Description:  "Nucleus sampling parameter. Top probability mass to consider.",
			},
			"top_k": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Top-k sampling parameter. Only consider top k tokens.",
			},
			"include": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Optional fields to include in the response.",
			},
			"instructions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional instructions to guide the model.",
			},
			"stop_sequences": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Optional list of sequences where the API will stop generating further tokens.",
			},
			"frequency_penalty": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ValidateFunc: validation.FloatBetween(-2.0, 2.0),
				Description:  "Penalty for token frequency between -2.0 and 2.0.",
			},
			"presence_penalty": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ValidateFunc: validation.FloatBetween(-2.0, 2.0),
				Description:  "Penalty for token presence between -2.0 and 2.0.",
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A unique identifier representing the end-user, to help track and detect abuse.",
			},
			"preserve_on_change": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, preserves the existing resource when parameters change. This prevents recreation but shows drift.",
			},
			// New fields for import tracking
			"imported": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Internal field to track if this resource was imported",
			},
			"_imported_resource": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Internal field to prevent recreation of imported resources",
			},
			// Response attributes
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Unique identifier for this response.",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Object type (usually 'model_response').",
			},
			"created": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Unix timestamp when the response was created.",
			},
			"output": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The generated output containing text and token count.",
			},
			"usage": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Token usage statistics for the request.",
			},
			"finish_reason": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Reason why the response finished (e.g., stop, length, content).",
			},
		},
		Description: `Generates text from OpenAI models using the responses API endpoint.
Note: OpenAI response resources are immutable. Any change to input parameters
will cause the existing resource to be deleted and a new one created. This is
due to the nature of the OpenAI API where responses cannot be modified once 
created.

Set preserve_on_change = true to prevent resource recreation on parameter changes.
If preservation is enabled, the resource will show drift but will not be recreated.`,
	}
}

// resourceOpenAIModelResponseCreate generates a new response using the given model.
func resourceOpenAIModelResponseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fmt.Printf("\n\n[RESOURCE-DEBUG] ========== RESOURCE CREATE DEBUG ==========\n")
	fmt.Printf("[RESOURCE-DEBUG] Function called: resourceOpenAIModelResponseCreate\n")
	fmt.Printf("[RESOURCE-DEBUG] Function address: %p\n", resourceOpenAIModelResponseCreate)

	providerClient, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	fmt.Printf("[RESOURCE-DEBUG] Retrieved provider client\n")
	fmt.Printf("[RESOURCE-DEBUG] Initial API URL: %s\n", providerClient.APIURL)
	fmt.Printf("[RESOURCE-DEBUG] Initial API Key (first 8 chars): %s***\n", providerClient.APIKey[:8])

	// Add additional debug info
	fmt.Printf("[RESOURCE-DEBUG] Provider client type: %T\n", providerClient)
	fmt.Printf("[RESOURCE-DEBUG] Provider client pointer: %p\n", providerClient)

	// Extract parameters from the resource data
	model, ok := d.GetOk("model")
	if !ok {
		return diag.Errorf("model is required")
	}

	input, ok := d.GetOk("input")
	if !ok {
		return diag.Errorf("input is required")
	}

	fmt.Printf("[RESOURCE-DEBUG] Model: %s\n", model.(string))
	fmt.Printf("[RESOURCE-DEBUG] Input (first 50 chars): %s...\n", truncateString(input.(string), 50))

	// Create a plain request map instead of using client.ModelResponseRequest
	requestBody := map[string]interface{}{
		"model": model.(string),
		"input": input.(string),
	}

	// Add optional parameters if present
	if v, ok := d.GetOk("max_output_tokens"); ok {
		requestBody["max_output_tokens"] = v.(int)
	}
	if v, ok := d.GetOk("temperature"); ok {
		requestBody["temperature"] = v.(float64)
	}
	if v, ok := d.GetOk("top_p"); ok {
		requestBody["top_p"] = v.(float64)
	}
	if v, ok := d.GetOk("top_k"); ok {
		requestBody["top_k"] = v.(int)
	}
	if v, ok := d.GetOk("instructions"); ok {
		requestBody["instructions"] = v.(string)
	}
	if v, ok := d.GetOk("frequency_penalty"); ok {
		// OpenAI API doesn't support frequency_penalty for this endpoint
		// Just log that we're ignoring it but don't add it to request
		fmt.Printf("[RESOURCE-DEBUG] Ignoring frequency_penalty=%v because API doesn't support it\n", v)
	}
	if v, ok := d.GetOk("presence_penalty"); ok {
		// OpenAI API doesn't support presence_penalty for this endpoint
		// Just log that we're ignoring it but don't add it to request
		fmt.Printf("[RESOURCE-DEBUG] Ignoring presence_penalty=%v because API doesn't support it\n", v)
	}
	if v, ok := d.GetOk("user"); ok {
		requestBody["user"] = v.(string)
	}

	// Handle stop sequences if supported by the API
	if v, ok := d.GetOk("stop_sequences"); ok {
		stopList := v.([]interface{})
		stops := make([]string, len(stopList))
		for i, v := range stopList {
			stops[i] = v.(string)
		}
		// API expects 'store' not 'stop' (per error message)
		requestBody["store"] = true
		fmt.Printf("[RESOURCE-DEBUG] Added store=true because stop sequences were provided\n")
	}

	// Usar directamente HTTP
	fmt.Printf("[RESOURCE-DEBUG] Using direct HTTP approach\n")

	// Create complete URL safely
	baseURL := providerClient.APIURL
	fmt.Printf("[RESOURCE-DEBUG] Original baseURL: %s\n", baseURL)

	// Define the url variable
	var url string

	// Normalize: If base URL already contains /v1, don't add it again
	if strings.HasSuffix(baseURL, "/v1") {
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
		url = baseURL + "responses"
		fmt.Printf("[RESOURCE-DEBUG] URL with /v1 in base: %s\n", url)
	} else {
		// Standard case - need to add /v1
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
		url = baseURL + "v1/responses"
		fmt.Printf("[RESOURCE-DEBUG] Standard URL: %s\n", url)
	}

	// Marshal the request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("[RESOURCE-DEBUG] Error marshaling request: %s\n", err)
		return diag.Errorf("Error marshaling request: %s", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("[RESOURCE-DEBUG] Error creating request: %s\n", err)
		return diag.Errorf("Error creating request: %s", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", providerClient.APIKey))
	if providerClient.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", providerClient.OrganizationID)
	}

	// Make the HTTP request
	fmt.Printf("[RESOURCE-DEBUG] Sending request to: %s\n", req.URL.String())
	resp, err := providerClient.HTTPClient.Do(req)
	if err != nil {
		fmt.Printf("[RESOURCE-DEBUG] Error making request: %s\n", err)
		return diag.Errorf("Error making request: %s", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[RESOURCE-DEBUG] Error reading response: %s\n", err)
		return diag.Errorf("Error reading response: %s", err)
	}

	// Verify the status code
	if resp.StatusCode >= 400 {
		fmt.Printf("[RESOURCE-DEBUG] Error response: %d - %s\n", resp.StatusCode, string(respBody))
		return diag.Errorf("API error: status code %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response as a map instead of using client.ModelResponse
	var responseMap map[string]interface{}
	if err := json.Unmarshal(respBody, &responseMap); err != nil {
		fmt.Printf("[RESOURCE-DEBUG] Error parsing response: %s\n", err)
		return diag.Errorf("Error parsing response: %s", err)
	}

	// Get the ID from the response
	responseID, ok := responseMap["id"].(string)
	if !ok {
		return diag.Errorf("Response did not contain an ID")
	}
	fmt.Printf("[RESOURCE-DEBUG] Successfully got response with ID: %s\n", responseID)

	// Set the resource ID
	d.SetId(responseID)
	fmt.Printf("[RESOURCE-DEBUG] Set resource ID to: %s\n", responseID)

	// Set other resource data
	if object, ok := responseMap["object"].(string); ok {
		if err := d.Set("object", object); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle the created timestamp (could be "created" or "created_at")
	if createdAt, ok := responseMap["created_at"].(float64); ok {
		if err := d.Set("created", int(createdAt)); err != nil {
			return diag.FromErr(err)
		}
	} else if created, ok := responseMap["created"].(float64); ok {
		if err := d.Set("created", int(created)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set the model
	if model, ok := responseMap["model"].(string); ok {
		if err := d.Set("model", model); err != nil {
			return diag.FromErr(err)
		}
	}

	// Process the output - this handles the newer "output" array format
	if output, ok := responseMap["output"].([]interface{}); ok && len(output) > 0 {
		if msg, ok := output[0].(map[string]interface{}); ok {
			outputMap := make(map[string]string)

			// Try to extract content array (newer format)
			if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
				if contentItem, ok := content[0].(map[string]interface{}); ok {
					if text, ok := contentItem["text"].(string); ok {
						outputMap["text"] = text
					}
				}
			}

			// If there's a role, include it
			if role, ok := msg["role"].(string); ok {
				outputMap["role"] = role
			}

			if err := d.Set("output", outputMap); err != nil {
				return diag.FromErr(err)
			}
		}
	} else if choices, ok := responseMap["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			// Set finish reason
			if finishReason, ok := choice["finish_reason"].(string); ok {
				if err := d.Set("finish_reason", finishReason); err != nil {
					return diag.FromErr(err)
				}
			}

			// Extract message content
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					outputMap := map[string]string{"text": content}
					if err := d.Set("output", outputMap); err != nil {
						return diag.FromErr(err)
					}
				}
			}
		}
	}

	// Handle usage statistics
	if usage, ok := responseMap["usage"].(map[string]interface{}); ok {
		usageMap := make(map[string]string)

		// Handle both naming conventions (prompt_tokens/input_tokens)
		if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
			usageMap["prompt_tokens"] = strconv.Itoa(int(promptTokens))
		} else if inputTokens, ok := usage["input_tokens"].(float64); ok {
			usageMap["prompt_tokens"] = strconv.Itoa(int(inputTokens))
		}

		// Handle both naming conventions (completion_tokens/output_tokens)
		if completionTokens, ok := usage["completion_tokens"].(float64); ok {
			usageMap["completion_tokens"] = strconv.Itoa(int(completionTokens))
		} else if outputTokens, ok := usage["output_tokens"].(float64); ok {
			usageMap["completion_tokens"] = strconv.Itoa(int(outputTokens))
		}

		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			usageMap["total_tokens"] = strconv.Itoa(int(totalTokens))
		}

		if err := d.Set("usage", usageMap); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

// resourceOpenAIModelResponseRead reads the model response from the state.
func resourceOpenAIModelResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// For read operations, we need to try to fetch the response from the API
	// The resource ID should be the response ID
	responseID := d.Id()
	if responseID == "" {
		return diag.Errorf("No ID is set")
	}

	// Get the API client
	client, err := GetOpenAIClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build the URL
	var url string
	if strings.HasSuffix(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/responses/%s", client.APIURL, responseID)
	} else {
		url = fmt.Sprintf("%s/v1/responses/%s", client.APIURL, responseID)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %w", err))
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %w", err))
	}
	defer resp.Body.Close()

	// If response doesn't exist (404), mark it as gone
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %w", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody)))
		}
		return diag.FromErr(fmt.Errorf("API error: %s - %s (%s)", errorResponse.Error.Type, errorResponse.Error.Message, errorResponse.Error.Code))
	}

	// Parse the response
	var responseMap map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseMap); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %w", err))
	}

	// Check if this is an imported resource and needs full field population
	isImported := d.Get("imported").(bool)

	// Set response data - but only what we need for state
	if object, ok := responseMap["object"].(string); ok {
		if err := d.Set("object", object); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle the created timestamp (could be "created" or "created_at")
	if createdAt, ok := responseMap["created_at"].(float64); ok {
		if err := d.Set("created", int(createdAt)); err != nil {
			return diag.FromErr(err)
		}
	} else if created, ok := responseMap["created"].(float64); ok {
		if err := d.Set("created", int(created)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set the model
	if model, ok := responseMap["model"].(string); ok {
		if err := d.Set("model", model); err != nil {
			return diag.FromErr(err)
		}
	}

	// For imported resources, also retrieve and set the input parameters from the API
	if isImported {
		// Always set input field for imported resources
		// First check in responseMap["input"] for the input field
		if input, ok := responseMap["input"].(string); ok {
			if err := d.Set("input", input); err != nil {
				return diag.FromErr(err)
			}
			log.Printf("[INFO] Set input for imported resource: %s", input)
		} else {
			// Try to extract from input_items or other properties
			// Sometimes input is stored in request.messages[0].content
			if request, ok := responseMap["request"].(map[string]interface{}); ok {
				if input, ok := request["input"].(string); ok {
					if err := d.Set("input", input); err != nil {
						return diag.FromErr(err)
					}
					log.Printf("[INFO] Set input from request for imported resource: %s", input)
				} else if messages, ok := request["messages"].([]interface{}); ok && len(messages) > 0 {
					if msg, ok := messages[0].(map[string]interface{}); ok {
						if content, ok := msg["content"].(string); ok {
							if err := d.Set("input", content); err != nil {
								return diag.FromErr(err)
							}
							log.Printf("[INFO] Set input from messages for imported resource: %s", content)
						}
					}
				}
			}
		}

		// Set optional parameters if they exist in the API response
		if temperature, ok := responseMap["temperature"].(float64); ok {
			if err := d.Set("temperature", temperature); err != nil {
				return diag.FromErr(err)
			}
		}

		if topP, ok := responseMap["top_p"].(float64); ok {
			if err := d.Set("top_p", topP); err != nil {
				return diag.FromErr(err)
			}
		}

		if maxTokens, ok := responseMap["max_output_tokens"].(float64); ok {
			if err := d.Set("max_output_tokens", int(maxTokens)); err != nil {
				return diag.FromErr(err)
			}
		}

		if topK, ok := responseMap["top_k"].(float64); ok {
			if err := d.Set("top_k", int(topK)); err != nil {
				return diag.FromErr(err)
			}
		}

		if instructions, ok := responseMap["instructions"].(string); ok {
			if err := d.Set("instructions", instructions); err != nil {
				return diag.FromErr(err)
			}
		}

		// Handle user field if present
		if user, ok := responseMap["user"].(string); ok {
			if err := d.Set("user", user); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Process the output - this handles the newer "output" array format
	if output, ok := responseMap["output"].([]interface{}); ok && len(output) > 0 {
		if msg, ok := output[0].(map[string]interface{}); ok {
			outputMap := make(map[string]string)

			// Try to extract content array (newer format)
			if content, ok := msg["content"].([]interface{}); ok && len(content) > 0 {
				if contentItem, ok := content[0].(map[string]interface{}); ok {
					if text, ok := contentItem["text"].(string); ok {
						outputMap["text"] = text
					}
				}
			}

			// If there's a role, include it
			if role, ok := msg["role"].(string); ok {
				outputMap["role"] = role
			}

			if err := d.Set("output", outputMap); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Handle usage statistics
	if usage, ok := responseMap["usage"].(map[string]interface{}); ok {
		usageMap := make(map[string]string)

		// Handle both naming conventions (prompt_tokens/input_tokens)
		if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
			usageMap["prompt_tokens"] = strconv.Itoa(int(promptTokens))
		} else if inputTokens, ok := usage["input_tokens"].(float64); ok {
			usageMap["prompt_tokens"] = strconv.Itoa(int(inputTokens))
		}

		// Handle both naming conventions (completion_tokens/output_tokens)
		if completionTokens, ok := usage["completion_tokens"].(float64); ok {
			usageMap["completion_tokens"] = strconv.Itoa(int(completionTokens))
		} else if outputTokens, ok := usage["output_tokens"].(float64); ok {
			usageMap["completion_tokens"] = strconv.Itoa(int(outputTokens))
		}

		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			usageMap["total_tokens"] = strconv.Itoa(int(totalTokens))
		}

		if err := d.Set("usage", usageMap); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

// resourceOpenAIModelResponseDelete is a no-op since there's no external resource to delete.
func resourceOpenAIModelResponseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// No-op: Model responses don't need to be deleted from the API
	d.SetId("")
	return nil
}

// resourceOpenAIModelResponseImport imports a model response into Terraform state.
func resourceOpenAIModelResponseImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[INFO] Importing OpenAI model response with ID: %s", d.Id())

	// Mark as imported resource
	_ = d.Set("imported", true)

	// Also set preserve_on_change to true for imported resources to prevent recreation
	_ = d.Set("preserve_on_change", true)

	// Get OpenAI client
	client, err := GetOpenAIClient(meta)
	if err != nil {
		return nil, err
	}

	// Build the URL for fetching the model response
	responseID := d.Id()
	var url string
	if strings.HasSuffix(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/responses/%s", client.APIURL, responseID)
	} else {
		url = fmt.Sprintf("%s/v1/responses/%s", client.APIURL, responseID)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// If response doesn't exist (404), mark it as gone
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("model response with ID %s not found", responseID)
	}

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
			return nil, fmt.Errorf("API returned error: %s - %s", resp.Status, string(responseBody))
		}
		return nil, fmt.Errorf("API error: %s - %s (%s)", errorResponse.Error.Type, errorResponse.Error.Message, errorResponse.Error.Code)
	}

	// Parse the response
	var responseMap map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseMap); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// For imported resources, we need to ensure the input field is set
	// We'll check in multiple places where input might be found
	inputFound := false

	// Try direct input field
	if inputVal, ok := responseMap["input"].(string); ok && inputVal != "" {
		_ = d.Set("input", inputVal)
		log.Printf("[DEBUG] Import: Found input directly in response: %s", truncateString(inputVal, 50))
		inputFound = true
	} else if request, ok := responseMap["request"].(map[string]interface{}); ok {
		// Try request fields where input might be stored
		if prompt, exists := request["prompt"]; exists && prompt != nil && prompt.(string) != "" {
			_ = d.Set("input", prompt)
			log.Printf("[DEBUG] Import: Found input in request.prompt: %s", truncateString(prompt.(string), 50))
			inputFound = true
		} else if input, exists := request["input"]; exists && input != nil && input.(string) != "" {
			_ = d.Set("input", input)
			log.Printf("[DEBUG] Import: Found input in request.input: %s", truncateString(input.(string), 50))
			inputFound = true
		} else if messages, exists := request["messages"].([]interface{}); exists && len(messages) > 0 {
			// For chat models, the input might be in the messages array
			for _, msg := range messages {
				msgMap, ok := msg.(map[string]interface{})
				if ok {
					if content, exists := msgMap["content"]; exists && content != nil &&
						msgMap["role"] == "user" && content.(string) != "" {
						_ = d.Set("input", content)
						log.Printf("[DEBUG] Import: Found input in messages[user].content: %s",
							truncateString(content.(string), 50))
						inputFound = true
						break
					}
				}
			}
		}
	}

	if !inputFound {
		log.Printf("[WARN] Import: Could not find input in API response. User may need to set it manually.")
	}

	// Handle top_p specially for imported resources
	if topP, ok := responseMap["top_p"].(float64); ok {
		_ = d.Set("top_p", topP)
		log.Printf("[DEBUG] Import: Found top_p in response: %v", topP)
	} else if request, ok := responseMap["request"].(map[string]interface{}); ok {
		if topP, ok := request["top_p"].(float64); ok {
			_ = d.Set("top_p", topP)
			log.Printf("[DEBUG] Import: Found top_p in request: %v", topP)
		}
	}

	// Mark as a special imported resource to prevent recreation
	_ = d.Set("_imported_resource", "true")

	// Read the computed attributes from the API response
	diags := resourceOpenAIModelResponseRead(ctx, d, meta)
	if diags.HasError() {
		return nil, fmt.Errorf("error reading imported model response: %s", diags[0].Summary)
	}

	log.Printf("[INFO] Successfully imported OpenAI model response with ID: %s", responseID)
	return []*schema.ResourceData{d}, nil
}

// truncateString safely truncates a string to maxLen characters, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// resourceOpenAIModelResponseUpdate handles updates to an existing OpenAI model response.
// Since responses are immutable, updates effectively create a new response unless preservation is enabled.
func resourceOpenAIModelResponseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Check if this is an imported resource or if preserve_on_change is enabled
	if d.Get("imported").(bool) || d.Get("preserve_on_change").(bool) {
		// For imported resources, we always preserve
		if d.Get("imported").(bool) {
			log.Printf("[INFO] OpenAI model response %s is imported and will be preserved despite parameter changes", d.Id())
		} else {
			log.Printf("[INFO] OpenAI model response %s is preserved despite parameter changes (drift will be shown)", d.Id())
		}

		// Call read to make sure state is up-to-date, but don't trigger recreation
		return resourceOpenAIModelResponseRead(ctx, d, meta)
	}

	// Log why we're recreating
	log.Printf("[INFO] OpenAI model response %s is immutable, creating a new one due to parameter changes", d.Id())

	// Clear the ID to force recreation
	d.SetId("")

	// Call create to generate a new response
	return resourceOpenAIModelResponseCreate(ctx, d, meta)
}

// resourceOpenAIModelResponseCustomDiff provides better visibility into resource recreation.
func resourceOpenAIModelResponseCustomDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// Only apply to existing resources
	if d.Id() == "" {
		return nil
	}

	// Check if preserve_on_change is true or this is an imported resource
	preserveOnChange := d.Get("preserve_on_change").(bool)
	imported := d.Get("imported").(bool)

	// If either preserve_on_change or imported is true, prevent recreation
	if preserveOnChange || imported {
		// For imported resources, ensure imported remains true
		if imported {
			_ = d.SetNew("imported", true)
			// Set a special flag to indicate this resource should not be recreated
			_ = d.SetNew("_imported_resource", "true")

			// For imported resources, ALWAYS preserve the input field from the state (from API)
			if d.HasChange("input") {
				oldVal, newVal := d.GetChange("input")
				log.Printf("[DEBUG] Input field changed for imported resource: old=%s, new=%s",
					truncateString(oldVal.(string), 30),
					truncateString(newVal.(string), 30))

				// If we have a non-empty value from the API (old value), always use that
				if oldVal != nil && oldVal.(string) != "" {
					_ = d.SetNew("input", oldVal)
					log.Printf("[INFO] Preserving original input from API for imported resource")
				} else if newVal != nil && newVal.(string) != "" {
					// Only use config value if we don't have one from API
					log.Printf("[INFO] Using input from configuration for imported resource")
				}
			}

			// Also handle top_p field specifically for imported resources
			if d.HasChange("top_p") {
				oldVal, newVal := d.GetChange("top_p")
				// Keep the old value if it's not nil and the new value is nil
				if oldVal != nil && newVal == nil {
					_ = d.SetNew("top_p", oldVal)
					log.Printf("[DEBUG] Preserving top_p value for imported resource: %v", oldVal)
				}
			}
		}

		// These are the fields that would normally cause recreation
		forceNewFields := []string{
			"model", "temperature", "max_output_tokens",
			"instructions", "top_k", "frequency_penalty",
			"presence_penalty", "user", "stop_sequences", "include",
		}

		// Collect changed parameters for logging
		changedParams := []string{}
		for _, param := range forceNewFields {
			if d.HasChange(param) {
				changedParams = append(changedParams, param)

				// Get the old value and keep it to prevent recreation
				oldValue, _ := d.GetChange(param)
				_ = d.SetNew(param, oldValue)
			}
		}

		if len(changedParams) > 0 {
			if imported {
				log.Printf("[INFO] Imported OpenAI model response will be preserved despite changes in: %s", strings.Join(changedParams, ", "))
			} else {
				log.Printf("[INFO] Preserving existing OpenAI model response despite changes in: %s (drift will be shown)", strings.Join(changedParams, ", "))
			}
		}
	} else {
		// Log parameters that will cause recreation if changed
		changedParams := []string{}
		recreationFields := []string{
			"input", "model", "temperature", "max_output_tokens",
			"instructions", "top_p", "top_k", "frequency_penalty",
			"presence_penalty", "user", "stop_sequences", "include",
		}

		for _, param := range recreationFields {
			if d.HasChange(param) {
				changedParams = append(changedParams, param)
			}
		}

		if len(changedParams) > 0 {
			log.Printf("[WARN] OpenAI model response will be recreated due to changes in: %s", strings.Join(changedParams, ", "))
		}
	}

	return nil
}
