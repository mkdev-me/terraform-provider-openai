package provider

// RunResponse represents the API response for a run.
type RunResponse struct {
	ID                  string                   `json:"id"`
	Object              string                   `json:"object"`
	CreatedAt           int64                    `json:"created_at"`
	ThreadID            string                   `json:"thread_id"`
	AssistantID         string                   `json:"assistant_id"`
	Status              string                   `json:"status"`
	StartedAt           *int64                   `json:"started_at,omitempty"`
	CompletedAt         *int64                   `json:"completed_at,omitempty"`
	Model               string                   `json:"model"`
	Instructions        string                   `json:"instructions"`
	Tools               []map[string]interface{} `json:"tools"` // Using generic map for tools representation
	FileIDs             []string                 `json:"file_ids"`
	Metadata            map[string]interface{}   `json:"metadata"`
	Usage               *RunUsage                `json:"usage,omitempty"`
	ExpiresAt           *int64                   `json:"expires_at,omitempty"`
	FailedAt            *int64                   `json:"failed_at,omitempty"`
	CancelledAt         *int64                   `json:"cancelled_at,omitempty"`
	RequiredAction      *RunRequiredAction       `json:"required_action,omitempty"`
	Temperature         *float64                 `json:"temperature,omitempty"`
	TopP                *float64                 `json:"top_p,omitempty"`
	MaxCompletionTokens *int                     `json:"max_completion_tokens,omitempty"`
	MaxPromptTokens     *int                     `json:"max_prompt_tokens,omitempty"`
	TruncationStrategy  *TruncationStrategy      `json:"truncation_strategy,omitempty"`
	ResponseFormat      *ResponseFormat          `json:"response_format,omitempty"` // simplified
	Stream              *bool                    `json:"stream,omitempty"`
}

// RunUsage represents token usage statistics for a run.
type RunUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// RunCreateRequest represents the request payload for creating a run.
type RunCreateRequest struct {
	AssistantID         string                   `json:"assistant_id"`
	Model               string                   `json:"model,omitempty"`
	Instructions        string                   `json:"instructions,omitempty"`
	Tools               []map[string]interface{} `json:"tools,omitempty"`
	Metadata            map[string]interface{}   `json:"metadata,omitempty"`
	Temperature         *float64                 `json:"temperature,omitempty"`
	TopP                *float64                 `json:"top_p,omitempty"`
	MaxCompletionTokens *int                     `json:"max_completion_tokens,omitempty"`
	MaxPromptTokens     *int                     `json:"max_prompt_tokens,omitempty"`
	TruncationStrategy  *TruncationStrategy      `json:"truncation_strategy,omitempty"`
	Stream              *bool                    `json:"stream,omitempty"`
	ResponseFormat      *ResponseFormat          `json:"response_format,omitempty"`
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
	MaxCompletionTokens *int                     `json:"max_completion_tokens,omitempty"`
	MaxPromptTokens     *int                     `json:"max_prompt_tokens,omitempty"`
	TruncationStrategy  *TruncationStrategy      `json:"truncation_strategy,omitempty"`
	ResponseFormat      *ResponseFormat          `json:"response_format,omitempty"`
}

// TruncationStrategy represents configuration for how a thread will be truncated.
type TruncationStrategy struct {
	Type          string `json:"type,omitempty"`
	LastNMessages int    `json:"last_n_messages,omitempty"`
}

// ResponseFormat represents the format configuration for the assistant's response.
// Note: This can be a string ("auto") or object. Using a custom unmarshal or handling both if possible.
// For now, assuming simple struct matching API usage in provider.
type ResponseFormat struct {
	Type       string      `json:"type,omitempty"`
	JSONSchema interface{} `json:"json_schema,omitempty"`
}

type RunRequiredAction struct {
	Type       string                 `json:"type"`
	SubmitTool *RunSubmitToolsRequest `json:"submit_tool_outputs,omitempty"`
}

type RunSubmitToolsRequest struct {
	ToolCalls []RunToolCall `json:"tool_calls"`
}

type RunToolCall struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Function *FunctionCall `json:"function,omitempty"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ListRunStepsResponse for generic step listing
type ListRunStepsResponse struct {
	Object  string            `json:"object"`
	Data    []RunStepResponse `json:"data"`
	HasMore bool              `json:"has_more"`
}

type RunStepResponse struct {
	ID        string                 `json:"id"`
	Object    string                 `json:"object"`
	CreatedAt int64                  `json:"created_at"`
	RunID     string                 `json:"run_id"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Details   map[string]interface{} `json:"details"`
}

// ListRunsResponse represents the API response for listing runs.
type ListRunsResponse struct {
	Object  string        `json:"object"`
	Data    []RunResponse `json:"data"`
	FirstID string        `json:"first_id"`
	LastID  string        `json:"last_id"`
	HasMore bool          `json:"has_more"`
}
