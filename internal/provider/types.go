package provider

import "encoding/json"

// FileResponse represents the API response for an OpenAI file.
// FileResponse represents the API response for an OpenAI file.
type FileResponse struct {
	ID        string `json:"id"`         // Unique identifier for the file
	Object    string `json:"object"`     // Type of object (e.g., "file")
	Bytes     int64  `json:"bytes"`      // Size of the file in bytes
	CreatedAt int64  `json:"created_at"` // Unix timestamp of file creation
	Filename  string `json:"filename"`   // Original name of the uploaded file
	Purpose   string `json:"purpose"`    // Intended use of the file (e.g., "fine-tune", "assistants")
}

// ListFilesResponse represents the API response for listing OpenAI files
type ListFilesResponse struct {
	Data   []FileResponse `json:"data"`
	Object string         `json:"object"`
}

// ErrorResponse represents an error response from the OpenAI API.
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"` // Human-readable error message
		Type    string `json:"type"`    // Type of error (e.g., "invalid_request_error")
		Code    string `json:"code"`    // Error code for programmatic handling
	} `json:"error"`
}

// ModelInfoResponse represents the API response for an OpenAI model info endpoint
type ModelInfoResponse struct {
	ID         string            `json:"id"`
	Object     string            `json:"object"`
	Created    int               `json:"created"`
	OwnedBy    string            `json:"owned_by"`
	Permission []ModelPermission `json:"permission"`
}

// ModelPermission represents the permission details for a model
type ModelPermission struct {
	ID                 string      `json:"id"`
	Object             string      `json:"object"`
	Created            int         `json:"created"`
	AllowCreateEngine  bool        `json:"allow_create_engine"`
	AllowSampling      bool        `json:"allow_sampling"`
	AllowLogprobs      bool        `json:"allow_logprobs"`
	AllowSearchIndices bool        `json:"allow_search_indices"`
	AllowView          bool        `json:"allow_view"`
	AllowFineTuning    bool        `json:"allow_fine_tuning"`
	Organization       string      `json:"organization"`
	Group              interface{} `json:"group"`
	IsBlocking         bool        `json:"is_blocking"`
}

// EmbeddingResponse represents the API response for text embeddings.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

type EmbeddingData struct {
	Object    string          `json:"object"`
	Index     int             `json:"index"`
	Embedding json.RawMessage `json:"embedding"` // float array or base64
}

type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type EmbeddingRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"` // string or []string
	User           string      `json:"user,omitempty"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     int         `json:"dimensions,omitempty"`
}

// ModerationResponse represents the API response for content moderation.
type ModerationResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

type ModerationResult struct {
	Flagged        bool               `json:"flagged"`
	Categories     map[string]bool    `json:"categories"`
	CategoryScores map[string]float64 `json:"category_scores"`
}

type ModerationRequest struct {
	Input interface{} `json:"input"` // string or []string
	Model string      `json:"model,omitempty"`
}
