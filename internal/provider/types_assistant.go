package provider

import "encoding/json"

// AssistantResponse represents the API response for an OpenAI assistant.
// It contains all the fields returned by the OpenAI API when creating or retrieving an assistant.
type AssistantResponse struct {
	ID              string                 `json:"id"`
	Object          string                 `json:"object"`
	CreatedAt       int                    `json:"created_at"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Model           string                 `json:"model"`
	Instructions    string                 `json:"instructions"`
	Tools           []AssistantTool        `json:"tools"`
	FileIDs         []string               `json:"file_ids"`                 // Legacy v1
	ToolResources   *ToolResources         `json:"tool_resources,omitempty"` // v2
	Metadata        map[string]interface{} `json:"metadata"`
	ResponseFormat  interface{}            `json:"response_format,omitempty"`
	ReasoningEffort string                 `json:"reasoning_effort,omitempty"`
	Temperature     float64                `json:"temperature,omitempty"`
	TopP            float64                `json:"top_p,omitempty"`
}

// ListAssistantsResponse represents the API response for listing OpenAI assistants.
type ListAssistantsResponse struct {
	Object  string              `json:"object"`   // Object type, always "list"
	Data    []AssistantResponse `json:"data"`     // Array of assistant objects
	FirstID string              `json:"first_id"` // ID of the first assistant in the list
	LastID  string              `json:"last_id"`  // ID of the last assistant in the list
	HasMore bool                `json:"has_more"` // Whether there are more assistants to fetch
}

// ToolResources represents the v2 tool resources (e.g. file_search vector stores)
type ToolResources struct {
	FileSearch      *FileSearchResources      `json:"file_search,omitempty"`
	CodeInterpreter *CodeInterpreterResources `json:"code_interpreter,omitempty"`
}

type FileSearchResources struct {
	VectorStoreIDs []string `json:"vector_store_ids,omitempty"`
}

type CodeInterpreterResources struct {
	FileIDs []string `json:"file_ids,omitempty"`
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
	Model         string                 `json:"model"`
	Name          string                 `json:"name,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Instructions  string                 `json:"instructions,omitempty"`
	Tools         []AssistantTool        `json:"tools,omitempty"`
	FileIDs       []string               `json:"file_ids,omitempty"` // Legacy v1? Provide support for backward compat or strictly v2?
	ToolResources *ToolResources         `json:"tool_resources,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AssistantFileResponse represents the API response for an OpenAI assistant file
type AssistantFileResponse struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	CreatedAt   int    `json:"created_at"`
	AssistantID string `json:"assistant_id"`
	FileID      string `json:"file_id"`
}

// AssistantFileCreateRequest represents the request to create an assistant file
type AssistantFileCreateRequest struct {
	FileID string `json:"file_id"`
}
