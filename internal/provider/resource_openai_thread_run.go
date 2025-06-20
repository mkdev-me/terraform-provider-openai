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

// ThreadRunRequest represents the request payload for creating a run and thread together in the OpenAI API.
// It contains all the configuration parameters needed to execute an assistant run,
// including thread configuration, model settings, tools, and execution parameters.
type ThreadRunRequest struct {
	AssistantID         string                   `json:"assistant_id"`                    // ID of the assistant to use for this run
	Thread              *ThreadCreateRequest     `json:"thread,omitempty"`                // Thread configuration
	Model               string                   `json:"model,omitempty"`                 // Optional model override for this run
	Instructions        string                   `json:"instructions,omitempty"`          // Optional instructions override for this run
	Tools               []map[string]interface{} `json:"tools,omitempty"`                 // Tools the assistant can use for this run
	Metadata            map[string]interface{}   `json:"metadata,omitempty"`              // Optional metadata for the run
	Temperature         *float64                 `json:"temperature,omitempty"`           // Sampling temperature (0-2)
	MaxCompletionTokens *int                     `json:"max_completion_tokens,omitempty"` // Maximum number of completion tokens to generate
	MaxPromptTokens     *int                     `json:"max_prompt_tokens,omitempty"`     // Maximum number of prompt tokens to use
	TopP                *float64                 `json:"top_p,omitempty"`                 // Nucleus sampling parameter (0-1)
	ResponseFormat      *ResponseFormat          `json:"response_format,omitempty"`       // Response format configuration
	Stream              *bool                    `json:"stream,omitempty"`                // Whether to stream the response
	ToolChoice          interface{}              `json:"tool_choice,omitempty"`           // Controls which tool is called
	TruncationStrategy  *TruncationStrategy      `json:"truncation_strategy,omitempty"`   // Controls how thread will be truncated
}

// ResponseFormat represents the format configuration for the assistant's response.
type ResponseFormat struct {
	Type       string      `json:"type,omitempty"`        // Format type (auto, json_object, etc.)
	JSONSchema interface{} `json:"json_schema,omitempty"` // JSON schema for structured output
}

// TruncationStrategy represents configuration for how a thread will be truncated.
type TruncationStrategy struct {
	Type          string `json:"type,omitempty"`            // Type of truncation strategy
	LastNMessages int    `json:"last_n_messages,omitempty"` // Number of messages to keep
}

// ThreadRunResponse represents the API response for a thread run creation.
// It contains all the details of the created run, including thread ID, status, and configuration.
type ThreadRunResponse struct {
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

// ThreadRun represents a run in the OpenAI API.
type ThreadRun struct {
	ID                  string                   `json:"id"`
	Object              string                   `json:"object"`
	CreatedAt           int                      `json:"created_at"`
	AssistantID         string                   `json:"assistant_id"`
	ThreadID            string                   `json:"thread_id"`
	Status              string                   `json:"status"`
	StartedAt           int                      `json:"started_at"`
	CompletedAt         int                      `json:"completed_at,omitempty"`
	LastError           *RunError                `json:"last_error,omitempty"`
	Model               string                   `json:"model"`
	Instructions        string                   `json:"instructions,omitempty"`
	Tools               []map[string]interface{} `json:"tools,omitempty"`
	FileIDs             []string                 `json:"file_ids,omitempty"`
	Metadata            map[string]interface{}   `json:"metadata,omitempty"`
	Usage               *RunUsage                `json:"usage,omitempty"`
	ExpiresAt           int                      `json:"expires_at,omitempty"`
	FailedAt            int                      `json:"failed_at,omitempty"`
	CancelledAt         int                      `json:"cancelled_at,omitempty"`
	RequiredAction      *RunRequiredAction       `json:"required_action,omitempty"`
	Temperature         *float64                 `json:"temperature,omitempty"`
	TopP                *float64                 `json:"top_p,omitempty"`
	ResponseFormat      *ResponseFormat          `json:"response_format,omitempty"`
	Stream              *bool                    `json:"stream,omitempty"`
	MaxCompletionTokens *int                     `json:"max_completion_tokens,omitempty"`
	MaxPromptTokens     *int                     `json:"max_prompt_tokens,omitempty"`
	TruncationStrategy  *TruncationStrategy      `json:"truncation_strategy,omitempty"`
}

// ThreadRunCreateRequest represents a request to create a new thread and run.
type ThreadRunCreateRequest struct {
	AssistantID         string                   `json:"assistant_id"`
	Thread              *ThreadCreateRequest     `json:"thread,omitempty"`
	Model               string                   `json:"model,omitempty"`
	Instructions        string                   `json:"instructions,omitempty"`
	Tools               []map[string]interface{} `json:"tools,omitempty"`
	Metadata            map[string]interface{}   `json:"metadata,omitempty"`
	Stream              *bool                    `json:"stream,omitempty"`
	Temperature         *float64                 `json:"temperature,omitempty"`
	TopP                *float64                 `json:"top_p,omitempty"`
	ResponseFormat      *ResponseFormat          `json:"response_format,omitempty"`
	MaxCompletionTokens *int                     `json:"max_completion_tokens,omitempty"`
	MaxPromptTokens     *int                     `json:"max_prompt_tokens,omitempty"`
	TruncationStrategy  *TruncationStrategy      `json:"truncation_strategy,omitempty"`
}

// ThreadRunMessageRequest represents a message in a thread run request.
type ThreadRunMessageRequest struct {
	Role        string                 `json:"role"`
	Content     string                 `json:"content"`
	FileIDs     []string               `json:"file_ids,omitempty"`
	Attachments []AttachmentRequest    `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RunError represents an error that occurred during a run.
type RunError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RunRequiredAction represents an action that is required to continue a run.
type RunRequiredAction struct {
	Type       string                 `json:"type"`
	SubmitTool *RunSubmitToolsRequest `json:"submit_tool_outputs,omitempty"`
}

// RunSubmitToolsRequest represents a request to submit tool outputs for a run.
type RunSubmitToolsRequest struct {
	ToolCalls []RunToolCall `json:"tool_calls"`
}

// RunToolCall represents a tool call that was made during a run.
type RunToolCall struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Function *FunctionCall `json:"function,omitempty"`
}

// FunctionCall represents a function call made by a tool.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// resourceOpenAIThreadRun defines the schema and CRUD operations for OpenAI thread runs.
// This resource allows users to create a thread and start a run in a single operation,
// providing a streamlined way to interact with OpenAI's Assistants API.
func resourceOpenAIThreadRun() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages an OpenAI thread run, which allows for the execution of an Assistant on a given thread.",
		CreateContext: resourceOpenAIThreadRunCreate,
		ReadContext:   resourceOpenAIThreadRunRead,
		DeleteContext: resourceOpenAIThreadRunDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIThreadRunImport,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identifier of the run, which can be referenced in API endpoints.",
			},
			"thread_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the thread that was created and associated with this run.",
			},
			"existing_thread_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "The ID of an existing thread to use for this run.",
				ConflictsWith: []string{"thread"},
			},
			"thread": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ForceNew:      true,
				Description:   "Configuration for creating a new thread for this run.",
				ConflictsWith: []string{"existing_thread_id"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"messages": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    true,
							Description: "Messages to create on the new thread.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"role": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"user"}, false),
										Description:  "The role of the entity that is creating the message. Currently only 'user' is supported.",
									},
									"content": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The content of the message.",
									},
									"attachments": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "A list of attachments to include in the message.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"file_id": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The ID of the file to attach to the message.",
												},
											},
										},
									},
									"metadata": {
										Type:        schema.TypeMap,
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "Set of key-value pairs that can be attached to the message.",
									},
								},
							},
						},
						"metadata": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Set of key-value pairs that can be attached to the thread.",
						},
					},
				},
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
				Description: "The ID of the model to use for the run. If not provided, the assistant's default model will be used.",
			},
			"instructions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Instructions that override the assistant's instructions for this run only.",
			},
			"tools": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Override the tools the assistant can use for this run.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"code_interpreter", "retrieval", "function"}, false),
							Description:  "The type of tool: code_interpreter, retrieval, or function.",
						},
						"function": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Required when type is function. Defines a function that can be called by the assistant.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the function.",
									},
									"description": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "A description of what the function does.",
									},
									"parameters": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsJSON,
										Description:  "The parameters the function accepts, described as a JSON Schema object.",
									},
								},
							},
						},
					},
				},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of key-value pairs that can be attached to the run.",
			},
			"stream": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Whether to stream the run results. Not currently supported through the Terraform provider.",
			},
			"max_completion_tokens": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "The maximum number of tokens that can be generated in the run completion.",
			},
			"temperature": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.FloatBetween(0, 2),
				Description:  "What sampling temperature to use, between 0 and 2. Higher values make output more random, lower values more deterministic.",
			},
			"top_p": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.FloatBetween(0, 1),
				Description:  "An alternative to sampling with temperature, where the model considers the results of the tokens with top_p probability mass.",
			},
			"response_format": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "Specifies the format of the response.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"text", "json_object"}, false),
							Description:  "Must be one of 'text' or 'json_object'.",
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The Unix timestamp (in seconds) of when the run was created.",
			},
			"completed_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The Unix timestamp (in seconds) of when the run was completed.",
			},
			"started_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The Unix timestamp (in seconds) of when the run was started.",
			},
			"object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The object type, which is always 'thread.run'.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the run, which can be: queued, in_progress, requires_action, cancelling, cancelled, failed, completed, or expired.",
			},
			"file_ids": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A list of file IDs that the run has access to.",
			},
			"usage": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Usage statistics for the run, including prompt_tokens, completion_tokens, and total_tokens.",
			},
		},
		CustomizeDiff: customizeOpenAIThreadRunDiff,
	}
}

// customizeOpenAIThreadRunDiff customizes the diff for OpenAI thread run resources
// to handle API default values and prevent unnecessary recreation of resources.
func customizeOpenAIThreadRunDiff(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	// Only apply the customization for updates, not for new resources
	if d.Id() == "" {
		return nil
	}

	// List of fields that could be provided by the API with default values
	apiProvidedFields := []string{"model", "instructions", "tools"}

	for _, field := range apiProvidedFields {
		// Skip if the field is explicitly set in the config
		if d.Get(field) != nil && !d.NewValueKnown(field) {
			continue
		}

		// Check if field exists in state but not in config
		oldVal, newVal := d.GetChange(field)
		if oldVal != nil && oldVal != "" && !reflect.ValueOf(oldVal).IsZero() &&
			(newVal == nil || newVal == "" || reflect.ValueOf(newVal).IsZero()) {
			// Suppress diff for this field
			if err := d.SetNew(field, oldVal); err != nil {
				return fmt.Errorf("error setting new value for %s: %w", field, err)
			}
		}
	}

	return nil
}

// resourceOpenAIThreadRunCreate creates a new OpenAI thread and starts a run.
// It processes the configuration from Terraform, constructs the API request,
// and handles the response to create a thread run.
func resourceOpenAIThreadRunCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*OpenAIClient)

	// Prepare the thread run request
	createRequest := &ThreadRunCreateRequest{
		AssistantID: d.Get("assistant_id").(string),
	}

	// Check if we're using an existing thread or creating a new one
	if existingThreadID, ok := d.GetOk("existing_thread_id"); ok {
		// Use existing thread - make API call directly to /threads/{thread_id}/runs
		threadID := existingThreadID.(string)
		return createRunOnExistingThread(ctx, d, m, threadID, createRequest)
	}

	// Add thread configuration if present
	if threadConfig, ok := d.GetOk("thread"); ok {
		threadList := threadConfig.([]interface{})
		if len(threadList) > 0 {
			threadMap := threadList[0].(map[string]interface{})
			threadRequest := &ThreadCreateRequest{}

			// Add messages if present
			if messagesConfig, ok := threadMap["messages"]; ok {
				messagesList := messagesConfig.([]interface{})
				if len(messagesList) > 0 {
					messages := make([]ThreadMessage, 0, len(messagesList))
					for _, msgConfig := range messagesList {
						msgMap := msgConfig.(map[string]interface{})
						message := ThreadMessage{
							Role:    msgMap["role"].(string),
							Content: msgMap["content"].(string),
						}

						// Add attachments if present
						if attachmentsConfig, ok := msgMap["attachments"]; ok {
							attachmentsList := attachmentsConfig.([]interface{})
							if len(attachmentsList) > 0 {
								// Extract the file_ids and add them to the message
								fileIDs := make([]string, 0, len(attachmentsList))
								for _, attachmentConfig := range attachmentsList {
									attachmentMap := attachmentConfig.(map[string]interface{})
									fileIDs = append(fileIDs, attachmentMap["file_id"].(string))
								}
								message.FileIDs = fileIDs
							}
						}

						// Add metadata if present
						if msgMetadata, ok := msgMap["metadata"]; ok {
							metadataMap := msgMetadata.(map[string]interface{})
							if len(metadataMap) > 0 {
								message.Metadata = metadataMap
							}
						}

						messages = append(messages, message)
					}
					threadRequest.Messages = messages
				}
			}

			// Add thread metadata if present
			if threadMetadata, ok := threadMap["metadata"]; ok {
				metadataMap := threadMetadata.(map[string]interface{})
				if len(metadataMap) > 0 {
					threadRequest.Metadata = metadataMap
				}
			}

			createRequest.Thread = threadRequest
		}
	}

	// Add model if present
	if model, ok := d.GetOk("model"); ok {
		createRequest.Model = model.(string)
	}

	// Add instructions if present
	if instructions, ok := d.GetOk("instructions"); ok {
		createRequest.Instructions = instructions.(string)
	}

	// Add tools if present
	if toolsConfig, ok := d.GetOk("tools"); ok {
		toolsList := toolsConfig.([]interface{})
		toolsRequests := make([]map[string]interface{}, 0, len(toolsList))

		for _, toolConfig := range toolsList {
			tool := make(map[string]interface{})
			for k, v := range toolConfig.(map[string]interface{}) {
				tool[k] = v
			}
			toolsRequests = append(toolsRequests, tool)
		}

		createRequest.Tools = toolsRequests
	}

	// Add metadata if present
	if metadataConfig, ok := d.GetOk("metadata"); ok {
		metadataMap := metadataConfig.(map[string]interface{})
		if len(metadataMap) > 0 {
			createRequest.Metadata = metadataMap
		}
	}

	// Add temperature if present
	if temperature, ok := d.GetOk("temperature"); ok {
		temp := temperature.(float64)
		createRequest.Temperature = &temp
	}

	// Add max_completion_tokens if present
	if maxCompletionTokens, ok := d.GetOk("max_completion_tokens"); ok {
		tokens := maxCompletionTokens.(int)
		createRequest.MaxCompletionTokens = &tokens
	}

	// Add max_prompt_tokens if present
	if maxPromptTokens, ok := d.GetOk("max_prompt_tokens"); ok {
		tokens := maxPromptTokens.(int)
		createRequest.MaxPromptTokens = &tokens
	}

	// Add top_p if present
	if topP, ok := d.GetOk("top_p"); ok {
		tp := topP.(float64)
		createRequest.TopP = &tp
	}

	// Add response_format if present
	if formatConfig, ok := d.GetOk("response_format"); ok {
		formatList := formatConfig.([]interface{})
		if len(formatList) > 0 {
			formatMap := formatList[0].(map[string]interface{})
			format := &ResponseFormat{
				Type: formatMap["type"].(string),
			}

			// Add json_schema if present
			if jsonSchema, ok := formatMap["json_schema"]; ok && jsonSchema.(string) != "" {
				var schemaObj interface{}
				if err := json.Unmarshal([]byte(jsonSchema.(string)), &schemaObj); err == nil {
					format.JSONSchema = schemaObj
				}
			}

			createRequest.ResponseFormat = format
		}
	}

	// Add stream if present
	if stream, ok := d.GetOk("stream"); ok {
		s := stream.(bool)
		createRequest.Stream = &s
	}

	// Add truncation_strategy if present
	if strategyConfig, ok := d.GetOk("truncation_strategy"); ok {
		strategyList := strategyConfig.([]interface{})
		if len(strategyList) > 0 {
			strategyMap := strategyList[0].(map[string]interface{})
			strategy := &TruncationStrategy{
				Type: strategyMap["type"].(string),
			}

			// Add last_n_messages if present
			if lastN, ok := strategyMap["last_n_messages"]; ok {
				strategy.LastNMessages = lastN.(int)
			}

			createRequest.TruncationStrategy = strategy
		}
	}

	// Convert the request to JSON
	requestBody, err := json.Marshal(createRequest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshalling request: %w", err))
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/threads/runs", client.APIURL)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %w", err))
	}

	// Add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("OpenAI-Beta", "assistants=v2")

	// Send the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %w", err))
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %w", err))
	}

	// Check for error responses
	if resp.StatusCode != http.StatusOK {
		return diag.FromErr(fmt.Errorf("API returned error - status code: %d, body: %s", resp.StatusCode, string(respBody)))
	}

	// Parse the response
	var runResponse ThreadRunResponse
	if err := json.Unmarshal(respBody, &runResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response body: %w", err))
	}

	// Set the resource ID
	d.SetId(runResponse.ID)

	// Store the thread ID
	if err := d.Set("thread_id", runResponse.ThreadID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting thread_id: %w", err))
	}

	// Read the current state to populate computed attributes
	return resourceOpenAIThreadRunRead(ctx, d, m)
}

// Helper function to create a run on an existing thread
func createRunOnExistingThread(ctx context.Context, d *schema.ResourceData, m interface{}, threadID string, createRequest *ThreadRunCreateRequest) diag.Diagnostics {
	client := m.(*OpenAIClient)

	// Create request body without the thread configuration
	createRequest.Thread = nil

	// Add all the standard fields (model, instructions, tools, etc.)
	if model, ok := d.GetOk("model"); ok {
		createRequest.Model = model.(string)
	}

	if instructions, ok := d.GetOk("instructions"); ok {
		createRequest.Instructions = instructions.(string)
	}

	// Process tools if present
	if toolsConfig, ok := d.GetOk("tools"); ok {
		toolsList := toolsConfig.([]interface{})
		toolsRequests := make([]map[string]interface{}, 0, len(toolsList))

		for _, toolConfig := range toolsList {
			tool := make(map[string]interface{})
			for k, v := range toolConfig.(map[string]interface{}) {
				tool[k] = v
			}
			toolsRequests = append(toolsRequests, tool)
		}

		createRequest.Tools = toolsRequests
	}

	// Add metadata if present
	if metadataConfig, ok := d.GetOk("metadata"); ok {
		metadataMap := metadataConfig.(map[string]interface{})
		if len(metadataMap) > 0 {
			createRequest.Metadata = metadataMap
		}
	}

	// Add temperature if present
	if temperature, ok := d.GetOk("temperature"); ok {
		temp := temperature.(float64)
		createRequest.Temperature = &temp
	}

	// Add max_completion_tokens if present
	if maxCompletionTokens, ok := d.GetOk("max_completion_tokens"); ok {
		tokens := maxCompletionTokens.(int)
		createRequest.MaxCompletionTokens = &tokens
	}

	// Add max_prompt_tokens if present
	if maxPromptTokens, ok := d.GetOk("max_prompt_tokens"); ok {
		tokens := maxPromptTokens.(int)
		createRequest.MaxPromptTokens = &tokens
	}

	// Add top_p if present
	if topP, ok := d.GetOk("top_p"); ok {
		tp := topP.(float64)
		createRequest.TopP = &tp
	}

	// Add response_format if present
	if formatConfig, ok := d.GetOk("response_format"); ok {
		formatList := formatConfig.([]interface{})
		if len(formatList) > 0 {
			formatMap := formatList[0].(map[string]interface{})
			format := &ResponseFormat{
				Type: formatMap["type"].(string),
			}

			// Add json_schema if present
			if jsonSchema, ok := formatMap["json_schema"]; ok && jsonSchema.(string) != "" {
				var schemaObj interface{}
				if err := json.Unmarshal([]byte(jsonSchema.(string)), &schemaObj); err == nil {
					format.JSONSchema = schemaObj
				}
			}

			createRequest.ResponseFormat = format
		}
	}

	// Add stream if present
	if stream, ok := d.GetOk("stream"); ok {
		s := stream.(bool)
		createRequest.Stream = &s
	}

	// Add truncation_strategy if present
	if strategyConfig, ok := d.GetOk("truncation_strategy"); ok {
		strategyList := strategyConfig.([]interface{})
		if len(strategyList) > 0 {
			strategyMap := strategyList[0].(map[string]interface{})
			strategy := &TruncationStrategy{
				Type: strategyMap["type"].(string),
			}

			// Add last_n_messages if present
			if lastN, ok := strategyMap["last_n_messages"]; ok {
				strategy.LastNMessages = lastN.(int)
			}

			createRequest.TruncationStrategy = strategy
		}
	}

	// Convert the request to JSON
	requestBody, err := json.Marshal(createRequest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshalling request: %w", err))
	}

	// Construct the API URL for existing thread
	url := fmt.Sprintf("%s/threads/%s/runs", client.APIURL, threadID)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %w", err))
	}

	// Add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("OpenAI-Beta", "assistants=v2")

	// Send the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %w", err))
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %w", err))
	}

	// Check for error responses
	if resp.StatusCode != http.StatusOK {
		return diag.FromErr(fmt.Errorf("API returned error - status code: %d, body: %s", resp.StatusCode, string(respBody)))
	}

	// Parse the response
	var runResponse ThreadRunResponse
	if err := json.Unmarshal(respBody, &runResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response body: %w", err))
	}

	// Set the resource ID
	d.SetId(runResponse.ID)

	// Store the thread ID
	if err := d.Set("thread_id", threadID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting thread_id: %w", err))
	}

	// Read the current state to populate computed attributes
	return resourceOpenAIThreadRunRead(ctx, d, m)
}

// resourceOpenAIThreadRunRead fetches the current state of an OpenAI thread run.
// It makes an API request to retrieve the run details and updates the Terraform state.
func resourceOpenAIThreadRunRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*OpenAIClient)

	// Get the run ID and thread ID
	runID := d.Id()
	threadID := d.Get("thread_id").(string)
	if threadID == "" {
		return diag.FromErr(fmt.Errorf("thread_id is required but was not set in the state"))
	}

	// Construct the API URL
	url := fmt.Sprintf("%s/threads/%s/runs/%s", client.APIURL, threadID, runID)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %w", err))
	}

	// Add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("OpenAI-Beta", "assistants=v2")

	// Send the request
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %w", err))
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %w", err))
	}

	// Check for error responses
	if resp.StatusCode == http.StatusNotFound {
		// The run no longer exists or was deleted
		d.SetId("")
		return nil
	} else if resp.StatusCode != http.StatusOK {
		return diag.FromErr(fmt.Errorf("API returned error - status code: %d, body: %s", resp.StatusCode, string(respBody)))
	}

	// Parse the response
	var runResponse ThreadRunResponse
	if err := json.Unmarshal(respBody, &runResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response body: %w", err))
	}

	// Set computed fields in the state
	if err := d.Set("assistant_id", runResponse.AssistantID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting assistant_id: %w", err))
	}
	if err := d.Set("model", runResponse.Model); err != nil {
		return diag.FromErr(fmt.Errorf("error setting model: %w", err))
	}
	if err := d.Set("instructions", runResponse.Instructions); err != nil {
		return diag.FromErr(fmt.Errorf("error setting instructions: %w", err))
	}
	if err := d.Set("status", runResponse.Status); err != nil {
		return diag.FromErr(fmt.Errorf("error setting status: %w", err))
	}
	if err := d.Set("object", runResponse.Object); err != nil {
		return diag.FromErr(fmt.Errorf("error setting object: %w", err))
	}
	if err := d.Set("created_at", runResponse.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created_at: %w", err))
	}
	if err := d.Set("file_ids", runResponse.FileIDs); err != nil {
		return diag.FromErr(fmt.Errorf("error setting file_ids: %w", err))
	}

	// Set optional fields
	if runResponse.StartedAt != nil {
		if err := d.Set("started_at", *runResponse.StartedAt); err != nil {
			return diag.FromErr(fmt.Errorf("error setting started_at: %w", err))
		}
	}
	if runResponse.CompletedAt != nil {
		if err := d.Set("completed_at", *runResponse.CompletedAt); err != nil {
			return diag.FromErr(fmt.Errorf("error setting completed_at: %w", err))
		}
	}

	// Set usage data if available
	if runResponse.Usage != nil {
		usageData := map[string]interface{}{
			"prompt_tokens":     runResponse.Usage.PromptTokens,
			"completion_tokens": runResponse.Usage.CompletionTokens,
			"total_tokens":      runResponse.Usage.TotalTokens,
		}
		if err := d.Set("usage", usageData); err != nil {
			return diag.FromErr(fmt.Errorf("error setting usage: %w", err))
		}
	}

	// Set tools data
	if len(runResponse.Tools) > 0 {
		if err := d.Set("tools", runResponse.Tools); err != nil {
			return diag.FromErr(fmt.Errorf("error setting tools: %w", err))
		}
	}

	// Set metadata if present
	if runResponse.Metadata != nil {
		metadata := make(map[string]string)
		for k, v := range runResponse.Metadata {
			if strVal, ok := v.(string); ok {
				metadata[k] = strVal
			} else {
				// Convert non-string values to JSON strings
				jsonData, err := json.Marshal(v)
				if err == nil {
					metadata[k] = string(jsonData)
				}
			}
		}
		if err := d.Set("metadata", metadata); err != nil {
			return diag.FromErr(fmt.Errorf("error setting metadata: %w", err))
		}
	}

	return nil
}

// resourceOpenAIThreadRunDelete manages the deletion of an OpenAI run.
// This function does not actually delete the run from OpenAI (as that's not supported),
// but it removes the resource from Terraform state.
func resourceOpenAIThreadRunDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Remove the resource from Terraform state without actually deleting it from OpenAI
	// since OpenAI doesn't support run deletion.
	return nil
}

// resourceOpenAIThreadRunImport imports an existing OpenAI thread run into Terraform state.
// It requires both the thread ID and run ID in the format "thread_id:run_id".
func resourceOpenAIThreadRunImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// The import ID should be in the format "thread_id:run_id"
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import format, expected 'thread_id:run_id'")
	}

	threadID := parts[0]
	runID := parts[1]

	// Set the thread ID in the resource
	if err := d.Set("thread_id", threadID); err != nil {
		return nil, fmt.Errorf("error setting thread_id: %w", err)
	}

	// Set the ID to just the run ID
	d.SetId(runID)

	// Read the run to populate the rest of the state
	if diags := resourceOpenAIThreadRunRead(ctx, d, m); diags.HasError() {
		return nil, fmt.Errorf("error reading imported run: %v", diags[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
