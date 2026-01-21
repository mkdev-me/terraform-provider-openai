package provider

// ThreadResponse represents the API response for an OpenAI thread.
// It contains the thread's identifier, creation timestamp, and associated metadata.
type ThreadResponse struct {
	ID            string                 `json:"id"`
	Object        string                 `json:"object"`
	CreatedAt     int                    `json:"created_at"`
	Metadata      map[string]interface{} `json:"metadata"`
	ToolResources *ToolResources         `json:"tool_resources,omitempty"` // v2 support
}

// ThreadCreateRequest represents the request payload for creating a thread in the OpenAI API.
// It can include initial messages and metadata for the thread.
type ThreadCreateRequest struct {
	Messages      []ThreadMessage        `json:"messages,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ToolResources *ToolResources         `json:"tool_resources,omitempty"` // v2 support
}

// ThreadMessage represents a message within a thread.
// Each message has a role, content, optional file attachments, and metadata.
type ThreadMessage struct {
	Role        string                 `json:"role"`
	Content     string                 `json:"content"`
	Attachments []AttachmentRequest    `json:"attachments,omitempty"` // v2
	FileIDs     []string               `json:"file_ids,omitempty"`    // v1 legacy
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MessageResponse represents the API response for an OpenAI message.
// It contains all the fields returned by the OpenAI API when creating or retrieving a message.
type MessageResponse struct {
	ID          string                 `json:"id"`
	Object      string                 `json:"object"`
	CreatedAt   int                    `json:"created_at"`
	ThreadID    string                 `json:"thread_id"`
	Role        string                 `json:"role"`
	Content     []MessageContent       `json:"content"`
	AssistantID string                 `json:"assistant_id,omitempty"`
	RunID       string                 `json:"run_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Attachments []MessageAttachment    `json:"attachments,omitempty"`
}

// MessageContent represents the content of a message.
// It can contain text or other types of content with their respective annotations.
type MessageContent struct {
	Type string              `json:"type"`
	Text *MessageContentText `json:"text,omitempty"`
}

// MessageContentText represents the text content of a message.
// It includes the text value and any associated annotations.
type MessageContentText struct {
	Value       string        `json:"value"`
	Annotations []interface{} `json:"annotations,omitempty"`
}

// MessageAttachment represents an attachment in a message.
type MessageAttachment struct {
	ID          string        `json:"id"`
	FileID      string        `json:"file_id,omitempty"` // Added for completeness if API returns it, or if V2 uses it?
	Type        string        `json:"type,omitempty"`    // SDKv2 had Type
	AssistantID string        `json:"assistant_id,omitempty"`
	CreatedAt   int           `json:"created_at,omitempty"`
	Tools       []ToolRequest `json:"tools,omitempty"` // V2 might return tools?
}

// MessageCreateRequest represents the payload for creating a message in the OpenAI API.
// It contains all the fields that can be set when creating a new message.
type MessageCreateRequest struct {
	Role        string                 `json:"role"`
	Content     string                 `json:"content"`
	Attachments []AttachmentRequest    `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AttachmentRequest represents a file attachment in a message creation request.
// Note: In v2 API this seems to be same as MessageAttachment structure used in request
type AttachmentRequest struct {
	FileID string        `json:"file_id"`
	Tools  []ToolRequest `json:"tools"`
}

// ToolRequest represents a tool in an attachment
type ToolRequest struct {
	Type string `json:"type"`
}
