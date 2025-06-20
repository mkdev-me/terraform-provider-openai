package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// OpenAIClient is a client for interacting with the OpenAI API
type OpenAIClient struct {
	APIKey         string
	OrganizationID string
	APIURL         string
	HTTPClient     *http.Client
}

// NewClient creates a new instance of the OpenAI client
func NewClient(apiKey, organizationID, apiURL string) *OpenAIClient {
	// Set default API URL if not provided
	if apiURL == "" {
		apiURL = "https://api.openai.com"
	}

	// Ensure the URL doesn't end with a slash
	if strings.HasSuffix(apiURL, "/") {
		apiURL = apiURL[:len(apiURL)-1]
	}

	// Debug: Print the client configuration
	fmt.Printf("DEBUG: Creating new OpenAI client with API URL: %s\n", apiURL)
	fmt.Printf("DEBUG: Organization ID: %s\n", organizationID)
	fmt.Printf("DEBUG: API key provided: %v\n", apiKey != "")

	// Create a custom transport with specific timeouts and DNS configuration
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   10,
	}

	return &OpenAIClient{
		APIKey:         apiKey,
		OrganizationID: organizationID,
		APIURL:         apiURL,
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second, // Increased timeout for API operations
		},
	}
}

// Project represents a project in OpenAI
type Project struct {
	Object         string        `json:"object"`
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description,omitempty"`
	OrganizationID string        `json:"organization_id,omitempty"`
	CreatedAt      *int64        `json:"created_at"`
	ArchivedAt     *int64        `json:"archived_at"`
	Status         string        `json:"status"`
	IsDefault      bool          `json:"is_default,omitempty"`
	BillingMode    string        `json:"billing_mode,omitempty"`
	APIKeys        []APIKey      `json:"api_keys,omitempty"`
	RateLimits     []RateLimit   `json:"rate_limits,omitempty"`
	Users          []ProjectUser `json:"users,omitempty"`
}

// ProjectUser represents a user associated with a project
type ProjectUser struct {
	Object  string `json:"object"`
	ID      string `json:"id"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	AddedAt int64  `json:"added_at"`
}

// RateLimit represents a rate limit configuration for a project
type RateLimit struct {
	ID                          string `json:"id"`
	Object                      string `json:"object"`
	Model                       string `json:"model"`
	MaxRequestsPer1Minute       int    `json:"max_requests_per_1_minute"`
	MaxTokensPer1Minute         int    `json:"max_tokens_per_1_minute"`
	MaxImagesPer1Minute         int    `json:"max_images_per_1_minute"`
	Batch1DayMaxInputTokens     int    `json:"batch_1_day_max_input_tokens"`
	MaxAudioMegabytesPer1Minute int    `json:"max_audio_megabytes_per_1_minute"`
	MaxRequestsPer1Day          int    `json:"max_requests_per_1_day"`
}

// RateLimitListResponse represents the response from the API when listing rate limits
type RateLimitListResponse struct {
	Object  string      `json:"object"`
	Data    []RateLimit `json:"data"`
	FirstID string      `json:"first_id"`
	LastID  string      `json:"last_id"`
	HasMore bool        `json:"has_more"`
}

// APIKey represents an API key in OpenAI
type APIKey struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreatedAt  *int64 `json:"created_at"`
	LastUsedAt *int64 `json:"last_used_at,omitempty"`
}

// CreateProjectRequest represents the request to create a project
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsDefault   bool   `json:"is_default,omitempty"`
}

// Error represents an error from the OpenAI API
type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// ErrorResponse represents an error response from the OpenAI API
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

// ListProjectsResponse represents the response from the API when listing projects
type ListProjectsResponse struct {
	Object  string    `json:"object"`
	Data    []Project `json:"data"`
	HasMore bool      `json:"has_more"`
}

// AdminAPIKey represents an API key in the OpenAI admin context
type AdminAPIKey struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	CreatedAt  int64    `json:"created_at"`
	ExpiresAt  *int64   `json:"expires_at,omitempty"`
	LastUsedAt *int64   `json:"last_used_at,omitempty"`
	Scopes     []string `json:"scopes"`
	Object     string   `json:"object"`
}

// AdminAPIKeyResponse represents the API response when creating an API key
type AdminAPIKeyResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	CreatedAt int64    `json:"created_at"`
	ExpiresAt *int64   `json:"expires_at,omitempty"`
	Object    string   `json:"object"`
	Scopes    []string `json:"scopes"`
	Key       string   `json:"key"`
}

// ListAPIKeysResponse represents the API response when listing API keys
type ListAPIKeysResponse struct {
	Object  string        `json:"object"`
	Data    []AdminAPIKey `json:"data"`
	HasMore bool          `json:"has_more"`
}

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name      string   `json:"name"`
	ExpiresAt *int64   `json:"expires_at,omitempty"`
	Scopes    []string `json:"scopes,omitempty"`
}

// CreateRateLimitRequest represents the request to create a rate limit
type CreateRateLimitRequest struct {
	ResourceType string `json:"resource_type"` // "request" or "token"
	LimitType    string `json:"limit_type"`    // "rpm" or "tpm"
	Value        int    `json:"value"`         // The limit value
}

// UpdateRateLimitRequest represents the request to update a rate limit
type UpdateRateLimitRequest struct {
	MaxRequestsPer1Minute       *int `json:"max_requests_per_1_minute,omitempty"`
	MaxTokensPer1Minute         *int `json:"max_tokens_per_1_minute,omitempty"`
	MaxImagesPer1Minute         *int `json:"max_images_per_1_minute,omitempty"`
	Batch1DayMaxInputTokens     *int `json:"batch_1_day_max_input_tokens,omitempty"`
	MaxAudioMegabytesPer1Minute *int `json:"max_audio_megabytes_per_1_minute,omitempty"`
	MaxRequestsPer1Day          *int `json:"max_requests_per_1_day,omitempty"`
}

// AddProjectUserRequest represents the request to add a user to a project
type AddProjectUserRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// ProjectUserList represents a list of users in a project
type ProjectUserList struct {
	Object  string        `json:"object"`
	Data    []ProjectUser `json:"data"`
	FirstID string        `json:"first_id"`
	LastID  string        `json:"last_id"`
	HasMore bool          `json:"has_more"`
}

// ProjectServiceAccount represents a service account in a project
type ProjectServiceAccount struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	Role      string `json:"role,omitempty"`
	APIKey    *struct {
		Object    string `json:"object"`
		Value     string `json:"value,omitempty"`
		Name      string `json:"name"`
		CreatedAt int64  `json:"created_at"`
		ID        string `json:"id"`
	} `json:"api_key,omitempty"`
}

// ProjectServiceAccountList represents a list of service accounts in a project
type ProjectServiceAccountList struct {
	Object  string                  `json:"object"`
	Data    []ProjectServiceAccount `json:"data"`
	FirstID string                  `json:"first_id"`
	LastID  string                  `json:"last_id"`
	HasMore bool                    `json:"has_more"`
}

// CreateProjectServiceAccountRequest represents the request to create a service account
type CreateProjectServiceAccountRequest struct {
	Name string `json:"name"`
}

// ChatCompletionRequest represents a request to the chat completion API
type ChatCompletionRequest struct {
	Model            string                  `json:"model"`                       // ID of the model to use
	Messages         []ChatCompletionMessage `json:"messages"`                    // List of messages in the conversation
	Functions        []ChatFunction          `json:"functions,omitempty"`         // Optional list of available functions
	FunctionCall     interface{}             `json:"function_call,omitempty"`     // Optional function call configuration
	Temperature      float64                 `json:"temperature,omitempty"`       // Sampling temperature
	TopP             float64                 `json:"top_p,omitempty"`             // Nucleus sampling parameter
	N                int                     `json:"n,omitempty"`                 // Number of completions to generate
	Stream           bool                    `json:"stream,omitempty"`            // Whether to stream the response
	Stop             []string                `json:"stop,omitempty"`              // Optional stop sequences
	MaxTokens        int                     `json:"max_tokens,omitempty"`        // Maximum tokens to generate
	PresencePenalty  float64                 `json:"presence_penalty,omitempty"`  // Presence penalty parameter
	FrequencyPenalty float64                 `json:"frequency_penalty,omitempty"` // Frequency penalty parameter
	LogitBias        map[string]float64      `json:"logit_bias,omitempty"`        // Optional token bias
	User             string                  `json:"user,omitempty"`              // Optional user identifier
}

// ChatCompletionResponse represents the API response for chat completions
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`      // Unique identifier for the completion
	Object  string                 `json:"object"`  // Type of object (e.g., "chat.completion")
	Created int                    `json:"created"` // Unix timestamp when the completion was created
	Model   string                 `json:"model"`   // Model used for the completion
	Choices []ChatCompletionChoice `json:"choices"` // List of possible completions
	Usage   ChatCompletionUsage    `json:"usage"`   // Token usage statistics
}

// ChatCompletionChoice represents a single completion option from the model
type ChatCompletionChoice struct {
	Index        int                   `json:"index"`         // Index of the choice in the list
	Message      ChatCompletionMessage `json:"message"`       // The generated message
	FinishReason string                `json:"finish_reason"` // Reason why the completion finished
}

// ChatCompletionMessage represents a message in the chat completion
type ChatCompletionMessage struct {
	Role         string            `json:"role"`                    // Role of the message sender (system, user, assistant)
	Content      string            `json:"content"`                 // Content of the message
	FunctionCall *ChatFunctionCall `json:"function_call,omitempty"` // Optional function call
	Name         string            `json:"name,omitempty"`          // Optional name of the message sender
}

// ChatFunctionCall represents a function call generated by the model
type ChatFunctionCall struct {
	Name      string `json:"name"`      // Name of the function to call
	Arguments string `json:"arguments"` // JSON string containing function arguments
}

// ChatCompletionUsage represents token usage statistics for the completion request
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`     // Number of tokens in the prompt
	CompletionTokens int `json:"completion_tokens"` // Number of tokens in the completion
	TotalTokens      int `json:"total_tokens"`      // Total number of tokens used
}

// ChatFunction represents a function that can be called by the model
type ChatFunction struct {
	Name        string          `json:"name"`                  // Name of the function
	Description string          `json:"description,omitempty"` // Optional function description
	Parameters  json.RawMessage `json:"parameters"`            // JSON schema for function parameters
}

// Vector Store related structs

// ExpiresAfter represents the expiration policy for a vector store
type ExpiresAfter struct {
	Days  *int64 `json:"days,omitempty"`
	Never *bool  `json:"never,omitempty"`
}

// ChunkingStrategy represents the chunking strategy for files in a vector store
type ChunkingStrategy struct {
	Type      string `json:"type"`
	Size      *int64 `json:"size,omitempty"`
	MaxTokens *int64 `json:"max_tokens,omitempty"`
}

// VectorStore represents an OpenAI Vector Store
type VectorStore struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	FileIDs   []string          `json:"file_ids"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt int64             `json:"created_at"`
	FileCount int               `json:"file_count"`
	Object    string            `json:"object"`
	Status    string            `json:"status"`
}

// VectorStoreFile represents a file in an OpenAI Vector Store
type VectorStoreFile struct {
	ID         string            `json:"id"`
	FileID     string            `json:"file_id"`
	Attributes map[string]string `json:"attributes,omitempty"`
	CreatedAt  int64             `json:"created_at"`
	Object     string            `json:"object"`
	Status     string            `json:"status"`
}

// VectorStoreFileBatch represents a batch operation for files in an OpenAI Vector Store
type VectorStoreFileBatch struct {
	ID         string                 `json:"id"`
	FileIDs    []string               `json:"file_ids"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	CreatedAt  int64                  `json:"created_at"`
	Object     string                 `json:"object"`
	Status     string                 `json:"status"`
}

// VectorStoreCreateParams contains parameters for creating a vector store
type VectorStoreCreateParams struct {
	Name             string            `json:"name"`
	FileIDs          []string          `json:"file_ids,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	ExpiresAfter     *ExpiresAfter     `json:"expires_after,omitempty"`
	ChunkingStrategy *ChunkingStrategy `json:"chunking_strategy,omitempty"`
}

// VectorStoreUpdateParams contains parameters for updating a vector store
type VectorStoreUpdateParams struct {
	ID           string            `json:"-"`
	Name         string            `json:"name,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	ExpiresAfter *ExpiresAfter     `json:"expires_after,omitempty"`
}

// VectorStoreFileCreateParams contains parameters for adding a file to a vector store
type VectorStoreFileCreateParams struct {
	VectorStoreID    string            `json:"-"`
	FileID           string            `json:"file_id"`
	Attributes       map[string]string `json:"attributes,omitempty"`
	ChunkingStrategy *ChunkingStrategy `json:"chunking_strategy,omitempty"`
}

// VectorStoreFileUpdateParams contains parameters for updating a file in a vector store
type VectorStoreFileUpdateParams struct {
	VectorStoreID string            `json:"-"`
	FileID        string            `json:"-"`
	Attributes    map[string]string `json:"attributes,omitempty"`
}

// VectorStoreFileBatchCreateParams contains parameters for adding a batch of files to a vector store
type VectorStoreFileBatchCreateParams struct {
	VectorStoreID    string                 `json:"-"`
	FileIDs          []string               `json:"file_ids"`
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
	ChunkingStrategy *ChunkingStrategy      `json:"chunking_strategy,omitempty"`
}

// VectorStoreFileBatchUpdateParams contains parameters for updating a batch of files in a vector store
type VectorStoreFileBatchUpdateParams struct {
	VectorStoreID string                 `json:"-"`
	BatchID       string                 `json:"-"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// ModelResponseOutputContent represents the content part of a message
type ModelResponseOutputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ModelResponseOutputMessage represents a message in the model response
type ModelResponseOutputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ModelResponseTokenUsage represents token usage in a model response
type ModelResponseTokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	InputTokens      int `json:"input_tokens,omitempty"`  // Alternative naming in some API versions
	OutputTokens     int `json:"output_tokens,omitempty"` // Alternative naming in some API versions
}

// ModelResponseFinishReason string representation of finish reason
type ModelResponseFinishReason string

// ModelResponse represents a response from the OpenAI model API
type ModelResponse struct {
	ID        string                     `json:"id"`
	Object    string                     `json:"object"`
	CreatedAt int                        `json:"created"`
	Model     string                     `json:"model"`
	Choices   []ModelResponseChoice      `json:"choices"`
	Usage     ModelResponseTokenUsage    `json:"usage"`
	System    ModelResponseSystemDetails `json:"system,omitempty"`
}

// ModelResponseSystemDetails contains system details
type ModelResponseSystemDetails struct {
	Version string `json:"version"`
}

// ModelResponseChoice represents a choice in the model response
type ModelResponseChoice struct {
	Index        int                        `json:"index"`
	Message      ModelResponseOutputMessage `json:"message"`
	FinishReason ModelResponseFinishReason  `json:"finish_reason"`
}

// ModelResponseRequest represents a request to create a model response
type ModelResponseRequest struct {
	Model       string   `json:"model"`
	Input       string   `json:"input"`
	MaxTokens   int      `json:"max_output_tokens,omitempty"`
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	TopK        int      `json:"top_k,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	User        string   `json:"user,omitempty"`
}

// AssistantResponse represents an individual assistant in the API response.
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

// AssistantTool represents a tool configuration for an assistant.
type AssistantTool struct {
	Type     string             `json:"type"`
	Function *AssistantFunction `json:"function,omitempty"`
}

// AssistantFunction represents a function configuration for an assistant tool.
type AssistantFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// User represents an OpenAI user
type User struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	AddedAt int64  `json:"added_at"`
}

// UsersResponse represents the response from the list users API
type UsersResponse struct {
	Object  string `json:"object"`
	Data    []User `json:"data"`
	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
	HasMore bool   `json:"has_more"`
}

// ListUsers retrieves a list of users in the organization
//
// Parameters:
//   - after: Cursor for pagination
//   - limit: Maximum number of users to return
//   - emails: Filter users by specific email addresses
//   - customAPIKey: Optional API key to use for this request
//
// Returns:
//   - A UsersResponse object containing the list of users
//   - An error if the operation failed
func (c *OpenAIClient) ListUsers(after string, limit int, emails []string, customAPIKey string) (*UsersResponse, error) {
	// Build query parameters
	queryParams := url.Values{}
	if after != "" {
		queryParams.Add("after", after)
	}
	if limit > 0 {
		queryParams.Add("limit", fmt.Sprintf("%d", limit))
	}
	if len(emails) > 0 {
		for _, email := range emails {
			queryParams.Add("emails", email)
		}
	}

	// Construct the URL for the request
	url := "/v1/organization/users"
	if len(queryParams) > 0 {
		url = url + "?" + queryParams.Encode()
	}

	// Log the request for debugging
	fmt.Printf("[DEBUG] Listing organization users\n")

	// Make the request with custom API key if provided
	var respBody []byte
	var err error
	if customAPIKey != "" {
		// Use custom API key for this request
		respBody, err = c.doRequestWithCustomAPIKey(http.MethodGet, url, nil, customAPIKey)
	} else {
		// Use default API key
		respBody, err = c.doRequest(http.MethodGet, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	// Parse the response
	var usersResponse UsersResponse
	if err := json.Unmarshal(respBody, &usersResponse); err != nil {
		return nil, fmt.Errorf("error decoding users response: %w", err)
	}

	return &usersResponse, nil
}

// FindUserByEmail finds a user in the organization by their email address
//
// Parameters:
//   - email: The email address of the user to find
//   - customAPIKey: Optional API key to use for this request
//
// Returns:
//   - The found User if it exists
//   - A boolean indicating if the user was found
//   - An error if the operation failed
func (c *OpenAIClient) FindUserByEmail(email string, customAPIKey string) (*User, bool, error) {
	// Create a list with just the one email we're looking for
	emails := []string{email}

	// Use the ListUsers function to filter by email
	usersResponse, err := c.ListUsers("", 1, emails, customAPIKey)
	if err != nil {
		return nil, false, fmt.Errorf("error finding user by email: %w", err)
	}

	// Check if any users were returned
	if len(usersResponse.Data) == 0 {
		return nil, false, nil
	}

	// Return the first matching user
	user := usersResponse.Data[0]

	// Double check that the email matches exactly (case insensitive)
	if strings.EqualFold(user.Email, email) {
		return &user, true, nil
	}

	// No exact match found
	return nil, false, nil
}

// GetUser retrieves a user by ID
func (c *OpenAIClient) GetUser(userID string, apiKey string) (*User, bool, error) {
	// Construct the correct URL using the API format
	url := fmt.Sprintf("/v1/organization/users/%s", userID)

	// Debug the request
	fmt.Printf("[DEBUG] Getting user with ID: %s\n", userID)

	// Make the request with custom API key if provided
	var respBody []byte
	var err error
	if apiKey != "" {
		// Use custom API key for this request
		respBody, err = c.doRequestWithCustomAPIKey("GET", url, nil, apiKey)
	} else {
		// Use default API key
		respBody, err = c.DoRequest("GET", url, nil)
	}

	// Handle 404 errors to indicate user not found
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error getting user: %w", err)
	}

	// Parse the response
	var user User
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, false, fmt.Errorf("error decoding user response: %w", err)
	}

	return &user, true, nil
}

// UpdateUserRole updates a user's role
func (c *OpenAIClient) UpdateUserRole(userID string, role string, apiKey string) (*User, error) {
	// Prepare request body
	body := map[string]string{
		"role": role,
	}

	// Construct the correct URL using the API format
	url := fmt.Sprintf("/v1/organization/users/%s", userID)

	// Debug the request
	fmt.Printf("[DEBUG] Updating user %s to role %s\n", userID, role)

	// Make the request with custom API key if provided
	var respBody []byte
	var err error
	if apiKey != "" {
		// Use custom API key for this request
		respBody, err = c.doRequestWithCustomAPIKey("POST", url, body, apiKey)
	} else {
		// Use default API key
		respBody, err = c.DoRequest("POST", url, body)
	}

	if err != nil {
		return nil, fmt.Errorf("error updating user role: %w", err)
	}

	// Parse the response
	var user User
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("error decoding user response: %w", err)
	}

	return &user, nil
}

// DoRequest performs an HTTP request with the given method, path, and body, and returns the response.
func (c *OpenAIClient) DoRequest(method, path string, body interface{}) ([]byte, error) {
	var jsonBody []byte
	var err error

	// Print base configuration for debugging
	fmt.Printf("OpenAI client config: API URL=%s, Organization ID=%s\n", c.APIURL, c.OrganizationID)

	// If body is provided, marshal it to JSON
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request: %w", err)
		}
	}

	// Construct the URL correctly, handling the case where path starts with "/v1"
	// and APIURL ends with "/v1" to avoid duplication
	var u string
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// Path is already a full URL
		u = path
	} else {
		// Check if we would have a duplicate /v1 in the path
		if strings.HasSuffix(c.APIURL, "/v1") && strings.HasPrefix(path, "/v1") {
			// Remove the /v1 prefix from the path to avoid duplication
			path = strings.TrimPrefix(path, "/v1")
		}
		u = SafeJoinURL(c.APIURL, path)
	}

	// Log the request details for debugging
	fmt.Printf("Making API request: %s %s\n", method, u)
	if body != nil {
		fmt.Printf("Request body: %s\n", string(jsonBody))
	}

	// Create the HTTP request
	req, err := http.NewRequest(method, u, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// SIEMPRE añadir el ID de organización como cabecera, independientemente de la URL
	// Este es el cambio principal para asegurarnos de que coincide con test_projects_api.go
	if c.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", c.OrganizationID)
		fmt.Printf("Setting OpenAI-Organization header: %s\n", c.OrganizationID)
	}

	// Make the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(responseBody, &errorResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(responseBody))
		}
		return nil, fmt.Errorf("API error: %s", errorResp.Error.Message)
	}

	return responseBody, nil
}

// doRequest is kept for backward compatibility but delegates to doRequestWithCustomAPIKey
func (c *OpenAIClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	return c.doRequestWithCustomAPIKey(method, path, body, c.APIKey)
}

// ListProjects retrieves a list of projects
func (c *OpenAIClient) ListProjects(limit int, includeArchived bool, after string) (*ListProjectsResponse, error) {
	// Construct query parameters
	var params []string
	if limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}
	if includeArchived {
		params = append(params, "include_archived=true")
	}
	if after != "" {
		params = append(params, fmt.Sprintf("after=%s", after))
	}

	// Combine parameters into query string
	queryString := ""
	if len(params) > 0 {
		queryString = "?" + strings.Join(params, "&")
	}

	// Use the exact endpoint structure consistent with the curl command
	url := fmt.Sprintf("/v1/organization/projects%s", queryString)

	// Debug info
	fmt.Printf("Listing projects\n")
	fmt.Printf("Using URL: %s\n", url)

	respBody, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var resp ListProjectsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal projects list response: %v", err)
	}

	return &resp, nil
}

// CreateProject creates a new project with the given name and description
func (c *OpenAIClient) CreateProject(name, description string, isDefault bool) (*Project, error) {
	// Create the request body
	requestBody := map[string]interface{}{
		"name": name,
	}

	// Only include description if it's not empty
	if description != "" {
		requestBody["description"] = description
	}

	// Only include is_default if it's true (don't send false)
	if isDefault {
		requestBody["is_default"] = isDefault
	}

	// Debug information
	fmt.Printf("Creating project with name: %s\n", name)
	fmt.Printf("Request body: %+v\n", requestBody)

	// Use the exact endpoint from the curl command that works
	url := "/v1/organization/projects"

	// Debug the URL
	fmt.Printf("Using URL for project creation: %s\n", url)

	// Make the API request
	responseBody, err := c.doRequest("POST", url, requestBody)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var project Project
	if err := json.Unmarshal(responseBody, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project response: %v", err)
	}

	return &project, nil
}

// GetProject retrieves a project by its ID
func (c *OpenAIClient) GetProject(id string) (*Project, error) {
	// Use the exact endpoint structure consistent with the curl command
	url := fmt.Sprintf("/v1/organization/projects/%s", id)

	// Debug info
	fmt.Printf("Getting project with ID: %s\n", id)
	fmt.Printf("Using URL: %s\n", url)

	responseBody, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var project Project
	if err := json.Unmarshal(responseBody, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project response: %v", err)
	}

	return &project, nil
}

// UpdateProject updates an existing project with the given details
func (c *OpenAIClient) UpdateProject(id, name, description string, isDefault bool) (*Project, error) {
	// Create the request body
	requestBody := map[string]interface{}{
		"name": name,
	}

	// Only include description if it's not empty
	if description != "" {
		requestBody["description"] = description
	}

	// Only include is_default if it's true (don't send false)
	if isDefault {
		requestBody["is_default"] = isDefault
	}

	// Use the exact endpoint structure consistent with the curl command
	url := fmt.Sprintf("/v1/organization/projects/%s", id)

	// Debug info
	fmt.Printf("Updating project with ID: %s\n", id)
	fmt.Printf("Using URL: %s\n", url)
	fmt.Printf("Request body: %+v\n", requestBody)

	responseBody, err := c.doRequest("POST", url, requestBody)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var project Project
	if err := json.Unmarshal(responseBody, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project response: %v", err)
	}

	return &project, nil
}

// DeleteProject deletes (archives) a project by its ID
func (c *OpenAIClient) DeleteProject(id string) error {
	// First get the project to retrieve its name
	project, err := c.GetProject(id)
	if err != nil {
		return fmt.Errorf("failed to get project before archiving: %v", err)
	}

	// Create a request body to archive the project, including the required 'name' field
	requestBody := map[string]interface{}{
		"name":   project.Name,
		"status": "archived",
	}

	// Use the exact endpoint structure consistent with the curl command
	url := fmt.Sprintf("/v1/organization/projects/%s", id)

	// Debug info
	fmt.Printf("Archiving project with ID: %s\n", id)
	fmt.Printf("Using URL: %s\n", url)
	fmt.Printf("Request body: %+v\n", requestBody)

	_, err = c.doRequest("POST", url, requestBody)
	if err != nil {
		return err
	}

	return nil
}

// ListAPIKeys retrieves the list of API keys for the organization
func (c *OpenAIClient) ListAPIKeys(limit int, after string) (*ListAPIKeysResponse, error) {
	// Construct the URL for the request with query params
	url := "/v1/organization/admin_api_keys"

	// Add query parameters if present
	queryParams := make([]string, 0)
	if limit > 0 {
		queryParams = append(queryParams, fmt.Sprintf("limit=%d", limit))
	}
	if after != "" {
		queryParams = append(queryParams, fmt.Sprintf("after=%s", after))
	}

	// Add the parameters to the URL
	if len(queryParams) > 0 {
		url += "?"
		for i, param := range queryParams {
			if i > 0 {
				url += "&"
			}
			url += param
		}
	}

	// Make the request
	respBody, err := c.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var listResponse ListAPIKeysResponse
	if err := json.Unmarshal(respBody, &listResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API keys list response: %v", err)
	}

	return &listResponse, nil
}

// GetAPIKey retrieves information about a specific API key
func (c *OpenAIClient) GetAPIKey(apiKeyID string) (*AdminAPIKey, error) {
	// Construct the URL for the request
	url := fmt.Sprintf("/v1/organization/admin_api_keys/%s", apiKeyID)

	// Make the request
	respBody, err := c.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var apiKey AdminAPIKey
	if err := json.Unmarshal(respBody, &apiKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API key response: %v", err)
	}

	return &apiKey, nil
}

// CreateAPIKey creates a new API key
func (c *OpenAIClient) CreateAPIKey(name string, expiresAt *int64, scopes []string) (*AdminAPIKeyResponse, error) {
	// Construct the URL for the request
	url := "/v1/organization/admin_api_keys"

	// Create the request body
	req := CreateAPIKeyRequest{
		Name:      name,
		ExpiresAt: expiresAt,
		Scopes:    scopes,
	}

	// Make the request
	respBody, err := c.doRequest(http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var apiKeyResp AdminAPIKeyResponse
	if err := json.Unmarshal(respBody, &apiKeyResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API key creation response: %v", err)
	}

	return &apiKeyResp, nil
}

// DeleteAPIKey deletes an API key
func (c *OpenAIClient) DeleteAPIKey(apiKeyID string) error {
	// Construct the URL for the request
	url := fmt.Sprintf("/v1/organization/admin_api_keys/%s", apiKeyID)

	// Make the request
	_, err := c.doRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	return nil
}

// CreateRateLimit creates a new rate limit for a project.
// It allows you to set restrictions on API usage based on requests or tokens per minute.
//
// Parameters:
//   - projectID: The ID of the project to apply the rate limit to
//   - resourceType: The type of resource to limit ("request" or "token")
//   - limitType: The type of limit to apply ("rpm" for requests per minute or "tpm" for tokens per minute)
//   - value: The numeric limit value
//
// Returns:
//   - A RateLimit object with details about the created rate limit
//   - An error if the operation failed
func (c *OpenAIClient) CreateRateLimit(projectID, resourceType, limitType string, value int) (*RateLimit, error) {
	// Create the request body
	req := CreateRateLimitRequest{
		ResourceType: resourceType,
		LimitType:    limitType,
		Value:        value,
	}

	// Construct the URL for the request
	url := fmt.Sprintf("/v1/organization/projects/%s/rate_limits", projectID)

	// Make the request
	respBody, err := c.doRequest(http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var rateLimit RateLimit
	if err := json.Unmarshal(respBody, &rateLimit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rate limit response: %v", err)
	}

	return &rateLimit, nil
}

// GetRateLimit retrieves information about a specific rate limit.
//
// Parameters:
//   - projectID: The ID of the project the rate limit belongs to
//   - rateLimitID: The ID of the rate limit to retrieve
//
// Returns:
//   - A RateLimit object with details about the requested rate limit
//   - An error if the operation failed or the rate limit doesn't exist
func (c *OpenAIClient) GetRateLimit(projectID, rateLimitID string) (*RateLimit, error) {
	return c.GetRateLimitWithKey(projectID, rateLimitID, "")
}

// GetRateLimitWithKey retrieves a rate limit from a project using a custom API key.
//
// Parameters:
//   - projectID: The ID of the project the rate limit belongs to
//   - modelOrRateLimitID: Either a model name (e.g. "gpt-4") or a rate limit ID (e.g. "rl-gpt-4" or "rl-gpt-4-projectsuffix")
//   - customAPIKey: Optional API key to use for this request instead of the client's default API key
//
// Returns:
//   - A RateLimit object with details about the rate limit
//   - An error if the operation failed or the rate limit doesn't exist
func (c *OpenAIClient) GetRateLimitWithKey(projectID, modelOrRateLimitID, customAPIKey string) (*RateLimit, error) {
	// Debug: Input parameters
	fmt.Printf("[DEBUG] GetRateLimitWithKey called with:\n")
	fmt.Printf("  - Project ID: %s\n", projectID)
	fmt.Printf("  - Model or Rate Limit ID: %s\n", modelOrRateLimitID)

	// Step 1: List all rate limits for the project
	fmt.Printf("[DEBUG] Listing all rate limits for project %s\n", projectID)
	rateLimits, err := c.ListRateLimitsWithKey(projectID, customAPIKey)
	if err != nil {
		fmt.Printf("[ERROR] Failed to list rate limits: %v\n", err)
		return nil, fmt.Errorf("failed to list rate limits: %w", err)
	}

	// Log the rate limits we got for debugging
	fmt.Printf("[DEBUG] Found %d rate limits for project\n", len(rateLimits.Data))
	for i, rl := range rateLimits.Data {
		fmt.Printf("[DEBUG] Rate limit #%d: ID=%s, Model=%s\n", i+1, rl.ID, rl.Model)
	}

	// Determine what we're looking for - model or rate limit ID
	model := ""
	searchID := modelOrRateLimitID

	if !strings.HasPrefix(modelOrRateLimitID, "rl-") {
		// If it doesn't have "rl-" prefix, it's a model name
		model = modelOrRateLimitID
		searchID = "rl-" + modelOrRateLimitID
		fmt.Printf("[DEBUG] Searching by MODEL: %s (ID would be %s)\n", model, searchID)
	} else {
		// It has the prefix, treat it as an ID but also extract the model
		// Try to extract model name from the ID
		parts := strings.Split(modelOrRateLimitID[3:], "-")
		if len(parts) >= 1 {
			model = parts[0]
			if len(parts) > 1 {
				// For multi-part model names like "gpt-4o-mini"
				if !containsProjectSuffix(parts[len(parts)-1]) {
					model = strings.Join(parts, "-")
				} else {
					model = strings.Join(parts[:len(parts)-1], "-")
				}
			}
		}
		fmt.Printf("[DEBUG] Searching by ID: %s (extracted model: %s)\n", searchID, model)
	}

	// SIMPLE MATCHING STRATEGY:
	// 1. Try exact ID match first
	for _, rl := range rateLimits.Data {
		if rl.ID == searchID {
			fmt.Printf("[DEBUG] Found exact ID match: %s (Model: %s)\n", rl.ID, rl.Model)
			return &rl, nil
		}
	}

	// 2. Try ID prefix match (accounts for project-specific suffixes)
	for _, rl := range rateLimits.Data {
		if strings.HasPrefix(rl.ID, searchID) || rl.ID == searchID {
			fmt.Printf("[DEBUG] Found ID prefix match: %s (Model: %s)\n", rl.ID, rl.Model)
			return &rl, nil
		}
	}

	// 3. Try model name match as fallback
	if model != "" {
		for _, rl := range rateLimits.Data {
			if rl.Model == model {
				fmt.Printf("[DEBUG] Found by model name match: %s (ID: %s)\n", rl.Model, rl.ID)
				return &rl, nil
			}
		}
	}

	// 4. Last resort: try partial model matching for compound model names
	if model != "" && strings.Contains(model, "-") {
		modelBase := strings.Split(model, "-")[0]
		for _, rl := range rateLimits.Data {
			if strings.HasPrefix(rl.Model, modelBase) {
				fmt.Printf("[DEBUG] Found by partial model name match: %s matches base %s (ID: %s)\n",
					rl.Model, modelBase, rl.ID)
				return &rl, nil
			}
		}
	}

	// Not found after all attempts
	fmt.Printf("[ERROR] Rate limit not found for ID/model: %s\n", modelOrRateLimitID)
	return nil, fmt.Errorf("API error: Project with ID '%s' not found or rate limit '%s' does not exist",
		projectID, modelOrRateLimitID)
}

// Helper function to check if a string looks like a project suffix
// (typically 8 or fewer characters at the end of a rate limit ID)
func containsProjectSuffix(s string) bool {
	return len(s) <= 8 && len(s) > 0
}

// UpdateRateLimit modifies an existing rate limit for a project.
func (c *OpenAIClient) UpdateRateLimit(projectID, rateLimitID string, maxRequestsPerMinute, maxTokensPerMinute, maxImagesPerMinute, batch1DayMaxInputTokens, maxAudioMegabytesPer1Minute, maxRequestsPer1Day *int) (*RateLimit, error) {
	return c.UpdateRateLimitWithKey(projectID, rateLimitID, maxRequestsPerMinute, maxTokensPerMinute, maxImagesPerMinute, batch1DayMaxInputTokens, maxAudioMegabytesPer1Minute, maxRequestsPer1Day, "")
}

// UpdateRateLimitWithKey modifies an existing rate limit for a project using a custom API key.
func (c *OpenAIClient) UpdateRateLimitWithKey(projectID, modelOrRateLimitID string, maxRequestsPerMinute, maxTokensPerMinute, maxImagesPerMinute, batch1DayMaxInputTokens, maxAudioMegabytesPer1Minute, maxRequestsPer1Day *int, customAPIKey string) (*RateLimit, error) {
	// Debug: Input parameters
	fmt.Printf("[UPDATE-RL-DEBUG] ========== UpdateRateLimitWithKey DEBUG ==========\n")
	fmt.Printf("[UPDATE-RL-DEBUG] ProjectID: %s\n", projectID)
	fmt.Printf("[UPDATE-RL-DEBUG] ModelOrRateLimitID: %s\n", modelOrRateLimitID)

	// Step 1: List all rate limits and find the one we want to update
	fmt.Printf("[UPDATE-RL-DEBUG] Listing all rate limits to find target rate limit\n")
	rateLimits, err := c.ListRateLimitsWithKey(projectID, customAPIKey)
	if err != nil {
		fmt.Printf("[UPDATE-RL-DEBUG] Failed to list rate limits: %v\n", err)
		return nil, fmt.Errorf("failed to list rate limits: %w", err)
	}

	// Search logic similar to GetRateLimitWithKey
	model := ""
	searchID := modelOrRateLimitID

	if !strings.HasPrefix(modelOrRateLimitID, "rl-") {
		// If it doesn't have "rl-" prefix, it's a model name
		model = modelOrRateLimitID
		searchID = "rl-" + modelOrRateLimitID
		fmt.Printf("[UPDATE-RL-DEBUG] Searching by MODEL: %s (ID would be %s)\n", model, searchID)
	} else {
		// It has the prefix, treat it as an ID but also extract the model
		// Try to extract model name from the ID
		parts := strings.Split(modelOrRateLimitID[3:], "-")
		if len(parts) >= 1 {
			model = parts[0]
			if len(parts) > 1 {
				// For multi-part model names like "gpt-4o-mini"
				if !containsProjectSuffix(parts[len(parts)-1]) {
					model = strings.Join(parts, "-")
				} else {
					model = strings.Join(parts[:len(parts)-1], "-")
				}
			}
		}
		fmt.Printf("[UPDATE-RL-DEBUG] Searching by ID: %s (extracted model: %s)\n", searchID, model)
	}

	// Find the target rate limit
	var targetRateLimit *RateLimit

	// 1. Try exact ID match first
	for _, rl := range rateLimits.Data {
		if rl.ID == searchID {
			fmt.Printf("[UPDATE-RL-DEBUG] Found exact ID match: %s (Model: %s)\n", rl.ID, rl.Model)
			targetRateLimit = &rl
			break
		}
	}

	// 2. Try ID prefix match (accounts for project-specific suffixes)
	if targetRateLimit == nil {
		for _, rl := range rateLimits.Data {
			if strings.HasPrefix(rl.ID, searchID+"-") || rl.ID == searchID {
				fmt.Printf("[UPDATE-RL-DEBUG] Found ID prefix match: %s (Model: %s)\n", rl.ID, rl.Model)
				targetRateLimit = &rl
				break
			}
		}
	}

	// 3. Try model name match as fallback
	if targetRateLimit == nil && model != "" {
		for _, rl := range rateLimits.Data {
			if rl.Model == model {
				fmt.Printf("[UPDATE-RL-DEBUG] Found by model name match: %s (ID: %s)\n", rl.Model, rl.ID)
				targetRateLimit = &rl
				break
			}
		}
	}

	// 4. Last resort: try partial model matching for compound model names
	if targetRateLimit == nil && model != "" && strings.Contains(model, "-") {
		modelBase := strings.Split(model, "-")[0]
		for _, rl := range rateLimits.Data {
			if strings.HasPrefix(rl.Model, modelBase) {
				fmt.Printf("[UPDATE-RL-DEBUG] Found by partial model name match: %s matches base %s (ID: %s)\n",
					rl.Model, modelBase, rl.ID)
				targetRateLimit = &rl
				break
			}
		}
	}

	if targetRateLimit == nil {
		fmt.Printf("[UPDATE-RL-DEBUG] Rate limit not found for ID/model: %s\n", modelOrRateLimitID)
		fmt.Printf("[UPDATE-RL-DEBUG] ========== END UpdateRateLimitWithKey DEBUG ==========\n\n")
		return nil, fmt.Errorf("API error: Project with ID '%s' not found or rate limit '%s' does not exist",
			projectID, modelOrRateLimitID)
	}

	// Now we have the target rate limit, construct the URL for update using the CORRECT ID
	effectiveRateLimitID := targetRateLimit.ID
	fmt.Printf("[UPDATE-RL-DEBUG] Found rate limit to update: %s (Model: %s)\n", effectiveRateLimitID, targetRateLimit.Model)

	// Construct the direct path with the rate limit ID included
	path := fmt.Sprintf("/organization/projects/%s/rate_limits/%s", projectID, effectiveRateLimitID)
	fmt.Printf("[UPDATE-RL-DEBUG] Using path for update: %s\n", path)

	// Create the request body with only non-nil fields
	req := make(map[string]interface{})

	// Add the required "name" field - use the model name from the target rate limit
	req["name"] = targetRateLimit.Model
	fmt.Printf("[UPDATE-RL-DEBUG] Setting name: %s\n", targetRateLimit.Model)

	if maxRequestsPerMinute != nil {
		req["max_requests_per_1_minute"] = *maxRequestsPerMinute
		fmt.Printf("[UPDATE-RL-DEBUG] Setting max_requests_per_1_minute: %d\n", *maxRequestsPerMinute)
	}
	if maxTokensPerMinute != nil {
		req["max_tokens_per_1_minute"] = *maxTokensPerMinute
		fmt.Printf("[UPDATE-RL-DEBUG] Setting max_tokens_per_1_minute: %d\n", *maxTokensPerMinute)
	}
	if maxImagesPerMinute != nil {
		req["max_images_per_1_minute"] = *maxImagesPerMinute
		fmt.Printf("[UPDATE-RL-DEBUG] Setting max_images_per_1_minute: %d\n", *maxImagesPerMinute)
	}
	if batch1DayMaxInputTokens != nil {
		req["batch_1_day_max_input_tokens"] = *batch1DayMaxInputTokens
		fmt.Printf("[UPDATE-RL-DEBUG] Setting batch_1_day_max_input_tokens: %d\n", *batch1DayMaxInputTokens)
	}
	if maxAudioMegabytesPer1Minute != nil {
		req["max_audio_megabytes_per_1_minute"] = *maxAudioMegabytesPer1Minute
		fmt.Printf("[UPDATE-RL-DEBUG] Setting max_audio_megabytes_per_1_minute: %d\n", *maxAudioMegabytesPer1Minute)
	}
	if maxRequestsPer1Day != nil {
		req["max_requests_per_1_day"] = *maxRequestsPer1Day
		fmt.Printf("[UPDATE-RL-DEBUG] Setting max_requests_per_1_day: %d\n", *maxRequestsPer1Day)
	}

	// Log the request for debugging
	reqJson, _ := json.Marshal(req)
	fmt.Printf("[UPDATE-RL-DEBUG] Request body: %s\n", string(reqJson))
	fmt.Printf("[UPDATE-RL-DEBUG] Rate limit ID after processing: %s\n", effectiveRateLimitID)

	// Debug: Output the actual API key being used (first/last 4 chars)
	apiKeyToUse := c.APIKey
	if customAPIKey != "" {
		apiKeyToUse = customAPIKey
	}
	maskedKey := ""
	if len(apiKeyToUse) > 8 {
		maskedKey = apiKeyToUse[:4] + "..." + apiKeyToUse[len(apiKeyToUse)-4:]
	} else {
		maskedKey = "***"
	}
	fmt.Printf("[UPDATE-RL-DEBUG] Using API key: %s\n", maskedKey)

	// Send POST request to update the rate limit - OpenAI API requires POST, not PUT for this endpoint
	var body []byte
	if customAPIKey != "" {
		fmt.Printf("[UPDATE-RL-DEBUG] Using custom API key for request\n")
		body, err = c.doRequestWithCustomAPIKey(http.MethodPost, path, req, customAPIKey)
	} else {
		fmt.Printf("[UPDATE-RL-DEBUG] Using default API key for request\n")
		body, err = c.doRequest(http.MethodPost, path, req)
	}

	if err != nil {
		fmt.Printf("[UPDATE-RL-DEBUG] Update rate limit request failed: %v\n", err)
		fmt.Printf("[UPDATE-RL-DEBUG] ========== END UpdateRateLimitWithKey DEBUG ==========\n\n")
		return nil, err
	}

	// Debug: Log the response
	fmt.Printf("[UPDATE-RL-DEBUG] Response from OpenAI API: %s\n", string(body))

	// Parse the response
	var rateLimit RateLimit
	if err := json.Unmarshal(body, &rateLimit); err != nil {
		fmt.Printf("[UPDATE-RL-DEBUG] Failed to unmarshal rate limit response: %v\n", err)
		fmt.Printf("[UPDATE-RL-DEBUG] ========== END UpdateRateLimitWithKey DEBUG ==========\n\n")
		return nil, fmt.Errorf("failed to unmarshal rate limit response: %v", err)
	}

	// Set the requested values for fields that might not be included in the response
	if maxRequestsPerMinute != nil {
		rateLimit.MaxRequestsPer1Minute = *maxRequestsPerMinute
	}
	if maxTokensPerMinute != nil {
		rateLimit.MaxTokensPer1Minute = *maxTokensPerMinute
	}
	if maxImagesPerMinute != nil {
		rateLimit.MaxImagesPer1Minute = *maxImagesPerMinute
	}
	if batch1DayMaxInputTokens != nil {
		rateLimit.Batch1DayMaxInputTokens = *batch1DayMaxInputTokens
	}
	if maxAudioMegabytesPer1Minute != nil {
		rateLimit.MaxAudioMegabytesPer1Minute = *maxAudioMegabytesPer1Minute
	}
	if maxRequestsPer1Day != nil {
		rateLimit.MaxRequestsPer1Day = *maxRequestsPer1Day
	}

	fmt.Printf("[UPDATE-RL-DEBUG] Final rate limit object: %+v\n", rateLimit)
	fmt.Printf("[UPDATE-RL-DEBUG] ========== END UpdateRateLimitWithKey DEBUG ==========\n\n")
	return &rateLimit, nil
}

// DeleteRateLimit removes a rate limit from a project.
// This will allow the project to operate without this specific limitation.
//
// Parameters:
//   - projectID: The ID of the project the rate limit belongs to
//   - rateLimitID: The ID of the rate limit to delete
//
// Returns:
//   - An error if the operation failed
func (c *OpenAIClient) DeleteRateLimit(projectID, rateLimitID string) error {
	return c.DeleteRateLimitWithKey(projectID, rateLimitID, "")
}

// AddProjectUser adds a user to a project.
// Users must already be members of the organization to be added to a project.
//
// Parameters:
//   - projectID: The ID of the project to add the user to
//   - userID: The ID of the user to add
//   - role: The role to assign to the user ("owner" or "member")
//   - customAPIKey: Optional API key to use for this request instead of the client's default API key
//
// Returns:
//   - A ProjectUser object with details about the added user
//   - An error if the operation failed
func (c *OpenAIClient) AddProjectUser(projectID, userID, role string, customAPIKey string) (*ProjectUser, error) {
	// Validate role
	if role != "owner" && role != "member" {
		return nil, fmt.Errorf("invalid role: %s (must be 'owner' or 'member')", role)
	}

	// Create the request body
	req := AddProjectUserRequest{
		UserID: userID,
		Role:   role,
	}

	// Construct the URL for the request
	url := fmt.Sprintf("/v1/organization/projects/%s/users", projectID)

	// Log the request for debugging
	fmt.Printf("[DEBUG] Adding user %s to project %s with role %s\n", userID, projectID, role)

	// Make the request with custom API key if provided
	var respBody []byte
	var err error
	if customAPIKey != "" {
		// Use custom API key for this request
		respBody, err = c.doRequestWithCustomAPIKey(http.MethodPost, url, req, customAPIKey)
	} else {
		// Use default API key
		respBody, err = c.doRequest(http.MethodPost, url, req)
	}

	if err != nil {
		return nil, err
	}

	// Parse the response
	var projectUser ProjectUser
	if err := json.Unmarshal(respBody, &projectUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project user response: %v", err)
	}

	return &projectUser, nil
}

// DoRequestWithCustomAPIKey performs a request with a custom API key
func (c *OpenAIClient) DoRequestWithCustomAPIKey(method, path string, body interface{}, customAPIKey string) ([]byte, error) {
	fmt.Printf("[REQUEST-DEBUG] ========== HTTP REQUEST DEBUG ==========\n")
	fmt.Printf("[REQUEST-DEBUG] Method: %s, Path: %s\n", method, path)
	fmt.Printf("[REQUEST-DEBUG] API URL: %s\n", c.APIURL)
	fmt.Printf("[REQUEST-DEBUG] Organization ID: %s\n", c.OrganizationID)

	// DEBUG: Check what the API key looks like
	if customAPIKey != "" {
		if len(customAPIKey) > 15 {
			fmt.Printf("[API-KEY-DEBUG] Key prefix: %s...\n", customAPIKey[:15])
			if !strings.HasPrefix(customAPIKey, "sk-") {
				fmt.Printf("[API-KEY-DEBUG] WARNING: API key doesn't start with 'sk-'!\n")
			}
		} else {
			fmt.Printf("[API-KEY-DEBUG] Key is too short: %d chars\n", len(customAPIKey))
		}
	} else {
		fmt.Printf("[API-KEY-DEBUG] No custom API key provided\n")
	}

	// Test network connectivity first
	connectivityErr := c.TestNetworkConnectivity()
	if connectivityErr != nil {
		fmt.Printf("[REQUEST-DEBUG] Network connectivity test failed: %v\n", connectivityErr)
		// Continue anyway, but log the warning
		fmt.Printf("[REQUEST-DEBUG] Proceeding with request despite connectivity test failure\n")
	} else {
		fmt.Printf("[REQUEST-DEBUG] Network connectivity test passed\n")
	}

	// Network environment debugging
	fmt.Printf("[NETWORK-DEBUG] Go Version: %s\n", runtime.Version())
	fmt.Printf("[NETWORK-DEBUG] GODEBUG env: %s\n", os.Getenv("GODEBUG"))

	// Check if we can resolve api.openai.com directly
	ips, resolveErr := net.LookupIP("api.openai.com")
	if resolveErr != nil {
		fmt.Printf("[NETWORK-DEBUG] DNS resolution error: %v\n", resolveErr)
	} else {
		fmt.Printf("[NETWORK-DEBUG] Resolved IPs for api.openai.com: %v\n", ips)
	}

	// Construct the full URL using SafeJoinURL for proper path handling
	fullURL := SafeJoinURL(c.APIURL, path)
	fmt.Printf("[REQUEST-DEBUG] Final full URL: %s\n", fullURL)

	// Parse the URL to check its components
	parsedURL, parseErr := url.Parse(fullURL)
	if parseErr != nil {
		fmt.Printf("[NETWORK-DEBUG] URL parse error: %v\n", parseErr)
	} else {
		fmt.Printf("[NETWORK-DEBUG] URL scheme: %s, host: %s, path: %s\n",
			parsedURL.Scheme, parsedURL.Host, parsedURL.Path)
	}

	// Create a buffer for the body if provided
	var bodyBuffer io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			fmt.Printf("[REQUEST-DEBUG] Error marshaling body: %v\n", err)
			return nil, fmt.Errorf("error marshaling request body: %v", err)
		}
		bodyBuffer = bytes.NewBuffer(bodyJSON)
		fmt.Printf("[REQUEST-DEBUG] Request body: %s\n", string(bodyJSON))
	} else {
		fmt.Printf("[REQUEST-DEBUG] No request body provided\n")
	}

	// Create the request
	req, err := http.NewRequest(method, fullURL, bodyBuffer)
	if err != nil {
		fmt.Printf("[REQUEST-DEBUG] Error creating request: %v\n", err)
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	maskedKey := "*****"
	if len(customAPIKey) > 5 {
		maskedKey = customAPIKey[:5] + "*****"
	}
	fmt.Printf("[REQUEST-DEBUG] Using API key (masked): %s\n", maskedKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", customAPIKey))

	if c.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", c.OrganizationID)
		fmt.Printf("[REQUEST-DEBUG] Set OpenAI-Organization header: %s\n", c.OrganizationID)
	}

	// Additional useful headers for debugging
	req.Header.Set("User-Agent", "Terraform-Provider-OpenAI/1.0")

	// Print all headers for debugging (excluding auth token)
	fmt.Printf("[REQUEST-DEBUG] Request headers:\n")
	for key, values := range req.Header {
		if key != "Authorization" {
			fmt.Printf("[REQUEST-DEBUG]   %s: %s\n", key, values)
		} else {
			// For Authorization, print just the Bearer prefix and first few chars
			authValue := values[0]
			if len(authValue) > 15 {
				fmt.Printf("[REQUEST-DEBUG]   %s: Bearer %s...\n", key, authValue[7:15])
			} else {
				fmt.Printf("[REQUEST-DEBUG]   %s: [REDACTED]\n", key)
			}
		}
	}

	// Make the request
	fmt.Printf("[REQUEST-DEBUG] Sending HTTP request...\n")

	// Check HTTP client configuration
	if c.HTTPClient == nil {
		fmt.Printf("[NETWORK-DEBUG] HTTPClient is nil, creating default client\n")
		c.HTTPClient = &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout: 10 * time.Second,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
			},
		}
	} else {
		fmt.Printf("[NETWORK-DEBUG] Using existing HTTPClient with timeout: %v\n", c.HTTPClient.Timeout)
		if transport, ok := c.HTTPClient.Transport.(*http.Transport); ok {
			fmt.Printf("[NETWORK-DEBUG] Transport: MaxIdleConns=%d, IdleConnTimeout=%v\n",
				transport.MaxIdleConns, transport.IdleConnTimeout)
		}
	}

	// Start a timer to measure request duration
	startTime := time.Now()

	// Do the HTTP request with more error context
	resp, err := c.HTTPClient.Do(req)
	requestDuration := time.Since(startTime)
	fmt.Printf("[NETWORK-DEBUG] Request took %v\n", requestDuration)

	if err != nil {
		fmt.Printf("[NETWORK-DEBUG] HTTP request error type: %T\n", err)
		fmt.Printf("[NETWORK-DEBUG] Error details: %v\n", err)

		// Try to determine if it's a DNS error
		if urlErr, ok := err.(*url.Error); ok {
			fmt.Printf("[NETWORK-DEBUG] URL error: %v\n", urlErr)
			if dnsErr, ok := urlErr.Err.(*net.DNSError); ok {
				fmt.Printf("[NETWORK-DEBUG] DNS error: %v, Name: %s, Server: %s, IsTimeout: %v, IsTemporary: %v\n",
					dnsErr, dnsErr.Name, dnsErr.Server, dnsErr.IsTimeout, dnsErr.IsTemporary)
			}
		}

		fmt.Printf("[REQUEST-DEBUG] Error making request: %v\n", err)
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[REQUEST-DEBUG] Error reading response body: %v\n", err)
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Print response details
	fmt.Printf("[REQUEST-DEBUG] Response status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("[REQUEST-DEBUG] Response headers:\n")
	for key, values := range resp.Header {
		fmt.Printf("[REQUEST-DEBUG]   %s: %s\n", key, values)
	}

	// Print the response body (limit it for very large responses)
	if len(responseBody) > 0 {
		previewLength := 500
		if len(responseBody) < previewLength {
			fmt.Printf("[REQUEST-DEBUG] Full response body: %s\n", string(responseBody))
		} else {
			fmt.Printf("[REQUEST-DEBUG] Response body preview (first %d bytes): %s...\n",
				previewLength, string(responseBody[:previewLength]))
		}
	} else {
		fmt.Printf("[REQUEST-DEBUG] Empty response body\n")
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		fmt.Printf("[REQUEST-DEBUG] Error status code detected: %d\n", resp.StatusCode)

		var errorResp ErrorResponse
		if err := json.Unmarshal(responseBody, &errorResp); err != nil {
			fmt.Printf("[REQUEST-DEBUG] Failed to parse error response: %v\n", err)
			fmt.Printf("[REQUEST-DEBUG] ========== END HTTP REQUEST DEBUG ==========\n")
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(responseBody))
		}

		fmt.Printf("[REQUEST-DEBUG] Error message: %s\n", errorResp.Error.Message)
		fmt.Printf("[REQUEST-DEBUG] Error type: %s\n", errorResp.Error.Type)
		fmt.Printf("[REQUEST-DEBUG] Error code: %s\n", errorResp.Error.Code)
		fmt.Printf("[REQUEST-DEBUG] ========== END HTTP REQUEST DEBUG ==========\n")
		return nil, fmt.Errorf("API error: %s", errorResp.Error.Message)
	}

	fmt.Printf("[REQUEST-DEBUG] Request successful\n")
	fmt.Printf("[REQUEST-DEBUG] ========== END HTTP REQUEST DEBUG ==========\n")
	return responseBody, nil
}

// For backward compatibility, keep the unexported version that delegates to the exported one
func (c *OpenAIClient) doRequestWithCustomAPIKey(method, path string, body interface{}, customAPIKey string) ([]byte, error) {
	return c.DoRequestWithCustomAPIKey(method, path, body, customAPIKey)
}

// ListProjectUsers retrieves all users in a project.
// This function provides a way to check if a user is already in a project.
//
// Parameters:
//   - projectID: The ID of the project to list users from
//   - customAPIKey: Optional API key to use for this request instead of the client's default API key
//
// Returns:
//   - A ProjectUserList object with all users in the project
//   - An error if the operation failed
func (c *OpenAIClient) ListProjectUsers(projectID string, customAPIKey string) (*ProjectUserList, error) {
	// Construct the URL for the request
	url := fmt.Sprintf("/v1/organization/projects/%s/users", projectID)

	// Log the request for debugging
	fmt.Printf("[DEBUG] Listing users for project %s\n", projectID)

	// Make the request with custom API key if provided
	var respBody []byte
	var err error
	if customAPIKey != "" {
		// Use custom API key for this request
		respBody, err = c.doRequestWithCustomAPIKey(http.MethodGet, url, nil, customAPIKey)
	} else {
		// Use default API key
		respBody, err = c.doRequest(http.MethodGet, url, nil)
	}

	if err != nil {
		return nil, err
	}

	// Parse the response
	var userList ProjectUserList
	if err := json.Unmarshal(respBody, &userList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project user list response: %v", err)
	}

	return &userList, nil
}

// FindProjectUser checks if a user exists in a project by ID.
// This is a helper method to avoid having to iterate through users in the provider.
//
// Parameters:
//   - projectID: The ID of the project to check
//   - userID: The ID of the user to find
//   - customAPIKey: Optional API key to use for this request
//
// Returns:
//   - The found ProjectUser if it exists
//   - A boolean indicating if the user was found
//   - An error if the operation failed
func (c *OpenAIClient) FindProjectUser(projectID, userID string, customAPIKey string) (*ProjectUser, bool, error) {
	// Get all users in the project
	userList, err := c.ListProjectUsers(projectID, customAPIKey)
	if err != nil {
		return nil, false, err
	}

	// Look for the user in the list
	for _, user := range userList.Data {
		if user.ID == userID {
			return &user, true, nil
		}
	}

	// User not found
	return nil, false, nil
}

// FindProjectUserByEmail checks if a user exists in a project by email address.
// This is a helper method to allow looking up users by email instead of ID.
//
// Parameters:
//   - projectID: The ID of the project to check
//   - email: The email address of the user to find
//   - customAPIKey: Optional API key to use for this request
//
// Returns:
//   - The found ProjectUser if it exists
//   - A boolean indicating if the user was found
//   - An error if the operation failed
func (c *OpenAIClient) FindProjectUserByEmail(projectID, email string, customAPIKey string) (*ProjectUser, bool, error) {
	// Get all users in the project
	userList, err := c.ListProjectUsers(projectID, customAPIKey)
	if err != nil {
		return nil, false, err
	}

	// Look for the user with matching email in the list (case insensitive)
	for _, user := range userList.Data {
		if strings.EqualFold(user.Email, email) {
			return &user, true, nil
		}
	}

	// User not found
	return nil, false, nil
}

// RemoveProjectUser removes a user from a project.
// Users who are organization owners cannot be removed from projects.
//
// Parameters:
//   - projectID: The ID of the project to remove the user from
//   - userID: The ID of the user to remove
//   - customAPIKey: Optional API key to use for this request instead of the client's default API key
//
// Returns:
//   - An error if the operation failed
func (c *OpenAIClient) RemoveProjectUser(projectID, userID string, customAPIKey string) error {
	// Construct the URL for the request
	url := fmt.Sprintf("/v1/organization/projects/%s/users/%s", projectID, userID)

	// Log the request for debugging
	fmt.Printf("[DEBUG] Removing user %s from project %s\n", userID, projectID)

	// Make the request with custom API key if provided
	var err error
	if customAPIKey != "" {
		// Use custom API key for this request
		_, err = c.doRequestWithCustomAPIKey(http.MethodDelete, url, nil, customAPIKey)
	} else {
		// Use default API key
		_, err = c.doRequest(http.MethodDelete, url, nil)
	}

	if err != nil {
		return err
	}

	return nil
}

// UpdateProjectUserRequest represents the request to update a user's role in a project
type UpdateProjectUserRequest struct {
	Role string `json:"role"`
}

// UpdateProjectUser updates a user's role in a project.
// Organization owners' roles cannot be modified.
//
// Parameters:
//   - projectID: The ID of the project
//   - userID: The ID of the user to update
//   - role: The new role to assign ("owner" or "member")
//   - customAPIKey: Optional API key to use for this request
//
// Returns:
//   - A ProjectUser object with updated details
//   - An error if the operation failed
func (c *OpenAIClient) UpdateProjectUser(projectID, userID, role string, customAPIKey string) (*ProjectUser, error) {
	// Validate role
	if role != "owner" && role != "member" {
		return nil, fmt.Errorf("invalid role: %s (must be 'owner' or 'member')", role)
	}

	// Create the request body
	req := UpdateProjectUserRequest{
		Role: role,
	}

	// Construct the URL for the request
	url := fmt.Sprintf("/v1/organization/projects/%s/users/%s", projectID, userID)

	// Log the request for debugging
	fmt.Printf("[DEBUG] Updating user %s in project %s to role %s\n", userID, projectID, role)

	// Make the request with custom API key if provided
	var respBody []byte
	var err error
	if customAPIKey != "" {
		// Use custom API key for this request
		respBody, err = c.doRequestWithCustomAPIKey(http.MethodPost, url, req, customAPIKey)
	} else {
		// Use default API key
		respBody, err = c.doRequest(http.MethodPost, url, req)
	}

	if err != nil {
		return nil, err
	}

	// Parse the response
	var projectUser ProjectUser
	if err := json.Unmarshal(respBody, &projectUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project user response: %v", err)
	}

	return &projectUser, nil
}

// CreateProjectServiceAccount creates a new service account in a project
// Service accounts are bot users that are not associated with a real user
// and do not have the limitation of being removed when a user leaves an organization
func (c *OpenAIClient) CreateProjectServiceAccount(projectID, name string, customAPIKey string) (*ProjectServiceAccount, error) {
	// Correct URL format based on the API endpoint structure
	url := fmt.Sprintf("/v1/organization/projects/%s/service_accounts", projectID)

	// Create request body
	req := CreateProjectServiceAccountRequest{
		Name: name,
	}

	// Log the request for debugging
	fmt.Printf("[DEBUG] Creating service account '%s' in project %s\n", name, projectID)

	// Make the request with custom API key if provided
	var respBody []byte
	var err error
	if customAPIKey != "" {
		// Use custom API key for this request
		respBody, err = c.doRequestWithCustomAPIKey(http.MethodPost, url, req, customAPIKey)
	} else {
		// Use default API key
		respBody, err = c.doRequest(http.MethodPost, url, req)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating project service account: %w", err)
	}

	// Parse the response
	var serviceAccount ProjectServiceAccount
	if err := json.Unmarshal(respBody, &serviceAccount); err != nil {
		return nil, fmt.Errorf("error parsing service account response: %w", err)
	}

	return &serviceAccount, nil
}

// GetProjectServiceAccount retrieves information about a specific service account in a project
func (c *OpenAIClient) GetProjectServiceAccount(projectID, serviceAccountID string, customAPIKey string) (*ProjectServiceAccount, error) {
	// Correct URL format based on the API endpoint structure
	url := fmt.Sprintf("/v1/organization/projects/%s/service_accounts/%s", projectID, serviceAccountID)

	// Log the request for debugging
	fmt.Printf("[DEBUG] Getting service account %s from project %s\n", serviceAccountID, projectID)

	// Make the request with custom API key if provided
	var respBody []byte
	var err error
	if customAPIKey != "" {
		// Use custom API key for this request
		respBody, err = c.doRequestWithCustomAPIKey(http.MethodGet, url, nil, customAPIKey)
	} else {
		// Use default API key
		respBody, err = c.doRequest(http.MethodGet, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error getting project service account: %w", err)
	}

	// Parse the response
	var serviceAccount ProjectServiceAccount
	if err := json.Unmarshal(respBody, &serviceAccount); err != nil {
		return nil, fmt.Errorf("error parsing service account response: %w", err)
	}

	return &serviceAccount, nil
}

// ListProjectServiceAccounts retrieves all service accounts in a project
func (c *OpenAIClient) ListProjectServiceAccounts(projectID string, customAPIKey string) (*ProjectServiceAccountList, error) {
	// Correct URL format based on the API endpoint structure
	url := fmt.Sprintf("/v1/organization/projects/%s/service_accounts", projectID)

	// Log the request for debugging
	fmt.Printf("[DEBUG] Listing service accounts for project %s\n", projectID)

	// Make the request with custom API key if provided
	var respBody []byte
	var err error
	if customAPIKey != "" {
		// Use custom API key for this request
		respBody, err = c.doRequestWithCustomAPIKey(http.MethodGet, url, nil, customAPIKey)
	} else {
		// Use default API key
		respBody, err = c.doRequest(http.MethodGet, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error listing project service accounts: %w", err)
	}

	// Parse the response
	var serviceAccountList ProjectServiceAccountList
	if err := json.Unmarshal(respBody, &serviceAccountList); err != nil {
		return nil, fmt.Errorf("error parsing service account list response: %w", err)
	}

	return &serviceAccountList, nil
}

// DeleteProjectServiceAccount removes a service account from a project
func (c *OpenAIClient) DeleteProjectServiceAccount(projectID, serviceAccountID string, customAPIKey string) error {
	// Correct URL format based on the API endpoint structure
	url := fmt.Sprintf("/v1/organization/projects/%s/service_accounts/%s", projectID, serviceAccountID)

	// Log the request for debugging
	fmt.Printf("[DEBUG] Deleting service account %s from project %s\n", serviceAccountID, projectID)

	// Make the request with custom API key if provided
	var err error
	if customAPIKey != "" {
		// Use custom API key for this request
		_, err = c.doRequestWithCustomAPIKey(http.MethodDelete, url, nil, customAPIKey)
	} else {
		// Use default API key
		_, err = c.doRequest(http.MethodDelete, url, nil)
	}

	if err != nil {
		return fmt.Errorf("error deleting project service account: %w", err)
	}

	return nil
}

// ChatCompletion makes a request to the OpenAI Chat Completions API
func (c *OpenAIClient) ChatCompletion(request *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	url := "/v1/chat/completions"

	body, err := c.DoRequest("POST", url, request)
	if err != nil {
		return nil, err
	}

	var response ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing chat completion response: %v", err)
	}

	return &response, nil
}

// CreateVectorStore creates a new vector store
func (c *OpenAIClient) CreateVectorStore(ctx context.Context, params *VectorStoreCreateParams) (*VectorStore, error) {
	req, err := c.newRequest("POST", "v1/vector_stores", params)
	if err != nil {
		return nil, err
	}

	var result VectorStore
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetVectorStore retrieves a vector store by ID
func (c *OpenAIClient) GetVectorStore(ctx context.Context, id string) (*VectorStore, error) {
	req, err := c.newRequest("GET", fmt.Sprintf("v1/vector_stores/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var result VectorStore
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateVectorStore updates an existing vector store
func (c *OpenAIClient) UpdateVectorStore(ctx context.Context, params *VectorStoreUpdateParams) (*VectorStore, error) {
	req, err := c.newRequest("POST", fmt.Sprintf("v1/vector_stores/%s", params.ID), params)
	if err != nil {
		return nil, err
	}

	var result VectorStore
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteVectorStore deletes a vector store by ID
func (c *OpenAIClient) DeleteVectorStore(ctx context.Context, id string) error {
	req, err := c.newRequest("DELETE", fmt.Sprintf("v1/vector_stores/%s", id), nil)
	if err != nil {
		return err
	}

	return c.do(ctx, req, nil)
}

// AddFileToVectorStore adds a file to a vector store
func (c *OpenAIClient) AddFileToVectorStore(ctx context.Context, params *VectorStoreFileCreateParams) (*VectorStoreFile, error) {
	req, err := c.newRequest("POST", fmt.Sprintf("v1/vector_stores/%s/files", params.VectorStoreID), params)
	if err != nil {
		return nil, err
	}

	var result VectorStoreFile
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetVectorStoreFile retrieves a file from a vector store
func (c *OpenAIClient) GetVectorStoreFile(ctx context.Context, vectorStoreID, fileID string) (*VectorStoreFile, error) {
	req, err := c.newRequest("GET", fmt.Sprintf("v1/vector_stores/%s/files/%s", vectorStoreID, fileID), nil)
	if err != nil {
		return nil, err
	}

	var result VectorStoreFile
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateVectorStoreFile updates a file in a vector store
func (c *OpenAIClient) UpdateVectorStoreFile(ctx context.Context, params *VectorStoreFileUpdateParams) (*VectorStoreFile, error) {
	req, err := c.newRequest("POST", fmt.Sprintf("v1/vector_stores/%s/files/%s", params.VectorStoreID, params.FileID), params)
	if err != nil {
		return nil, err
	}

	var result VectorStoreFile
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RemoveFileFromVectorStore removes a file from a vector store
func (c *OpenAIClient) RemoveFileFromVectorStore(ctx context.Context, vectorStoreID, fileID string) error {
	req, err := c.newRequest("DELETE", fmt.Sprintf("v1/vector_stores/%s/files/%s", vectorStoreID, fileID), nil)
	if err != nil {
		return err
	}

	return c.do(ctx, req, nil)
}

// AddFileBatchToVectorStore adds a batch of files to a vector store
func (c *OpenAIClient) AddFileBatchToVectorStore(ctx context.Context, params *VectorStoreFileBatchCreateParams) (*VectorStoreFileBatch, error) {
	req, err := c.newRequest("POST", fmt.Sprintf("v1/vector_stores/%s/file_batches", params.VectorStoreID), params)
	if err != nil {
		return nil, err
	}

	var result VectorStoreFileBatch
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetVectorStoreFileBatch retrieves a file batch from a vector store
func (c *OpenAIClient) GetVectorStoreFileBatch(ctx context.Context, vectorStoreID, batchID string) (*VectorStoreFileBatch, error) {
	req, err := c.newRequest("GET", fmt.Sprintf("v1/vector_stores/%s/file_batches/%s", vectorStoreID, batchID), nil)
	if err != nil {
		return nil, err
	}

	var result VectorStoreFileBatch
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateVectorStoreFileBatch updates a file batch in a vector store
func (c *OpenAIClient) UpdateVectorStoreFileBatch(ctx context.Context, params *VectorStoreFileBatchUpdateParams) (*VectorStoreFileBatch, error) {
	req, err := c.newRequest("POST", fmt.Sprintf("v1/vector_stores/%s/file_batches/%s", params.VectorStoreID, params.BatchID), params)
	if err != nil {
		return nil, err
	}

	var result VectorStoreFileBatch
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RemoveFileBatchFromVectorStore removes a file batch from a vector store
func (c *OpenAIClient) RemoveFileBatchFromVectorStore(ctx context.Context, vectorStoreID, batchID string) error {
	req, err := c.newRequest("DELETE", fmt.Sprintf("v1/vector_stores/%s/file_batches/%s", vectorStoreID, batchID), nil)
	if err != nil {
		return err
	}

	return c.do(ctx, req, nil)
}

// newRequest creates a new HTTP request
func (c *OpenAIClient) newRequest(method, path string, body interface{}) (*http.Request, error) {
	// Make sure path has proper formatting
	if !strings.HasPrefix(path, "/") {
		if strings.HasPrefix(path, "v1/") {
			// Path already includes v1 prefix without leading slash
			path = "/" + path
		} else if !strings.HasPrefix(path, "v1") {
			// Path doesn't have v1 prefix at all, add it with leading slash
			path = "/v1/" + path
		} else {
			// Just add leading slash
			path = "/" + path
		}
	}

	// Ensure APIURL doesn't end with /v1 to avoid duplication
	baseURL := c.APIURL
	if strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/v1")
	}

	// Also remove trailing slash from base URL
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Construct the full URL using SafeJoinURL
	u := SafeJoinURL(baseURL, path)

	var req *http.Request
	var err error

	if body != nil {
		// Marshal the body to JSON
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %v", err)
		}

		// Create the request with the body
		req, err = http.NewRequest(method, u, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}
	} else {
		// Create the request without a body
		req, err = http.NewRequest(method, u, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// Add the organization ID header if provided
	if c.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", c.OrganizationID)
	}

	return req, nil
}

// do performs an HTTP request and decodes the response
func (c *OpenAIClient) do(ctx context.Context, req *http.Request, v interface{}) error {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	// Use the context if provided
	req = req.WithContext(ctx)

	// Perform the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	// Check for API errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return fmt.Errorf("API error: %s (%s)", errResp.Error.Message, errResp.Error.Code)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// If no target is provided, we're done
	if v == nil {
		return nil
	}

	// Unmarshal the response into the target
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("error parsing response: %v", err)
	}

	return nil
}

// SafeJoinURL safely joins a base URL with a path
func SafeJoinURL(baseURL, path string) string {
	// Debug logging to track URL construction
	fmt.Printf("[URL-DEBUG] SafeJoinURL called with baseURL=%s, path=%s\n", baseURL, path)

	// Check if the path already includes query parameters
	pathPart := path
	queryPart := ""
	if strings.Contains(path, "?") {
		parts := strings.SplitN(path, "?", 2)
		pathPart = parts[0]
		queryPart = "?" + parts[1]
	}

	// If the path itself is already a complete URL (containing "://"), parse it and use just the path component
	if strings.Contains(pathPart, "://") {
		parsedPath, err := url.Parse(pathPart)
		if err == nil {
			fmt.Printf("[URL-DEBUG] Path contains full URL, extracting just the path component\n")
			// Extract just the path component from the full URL in the path
			pathPart = parsedPath.Path
		}
	}

	// Special handling for rate limit endpoints
	if strings.Contains(pathPart, "/organization/projects/") && strings.Contains(pathPart, "/rate_limits") {
		// Extract project ID
		projectPattern := `/organization/projects/([^/]+)/rate_limits`
		projectRe := regexp.MustCompile(projectPattern)
		projectMatches := projectRe.FindStringSubmatch(pathPart)

		// Check if there's a rate limit ID in the path
		rateLimitPattern := `/organization/projects/([^/]+)/rate_limits/(.+)`
		rateLimitRe := regexp.MustCompile(rateLimitPattern)
		rateLimitMatches := rateLimitRe.FindStringSubmatch(pathPart)

		// Ensure the base URL is properly formatted
		baseWithoutTrailingSlash := strings.TrimSuffix(baseURL, "/")
		// Remove any "/v1" from the base URL to avoid duplication
		baseWithoutV1 := strings.TrimSuffix(baseWithoutTrailingSlash, "/v1")

		var correctURL string

		// If we have a rate limit ID in the path, preserve it for PUT/POST/DELETE operations
		if len(rateLimitMatches) == 3 {
			projectID := rateLimitMatches[1]
			rateLimitID := rateLimitMatches[2]

			// Keep rate limit ID in the path for update/delete operations
			correctURL = fmt.Sprintf("%s/v1/organization/projects/%s/rate_limits/%s",
				baseWithoutV1, projectID, rateLimitID)

			fmt.Printf("[URL-DEBUG] Constructed rate limit URL with ID: %s\n", correctURL)
		} else if len(projectMatches) == 2 {
			// No rate limit ID in the path, this is for list operations
			projectID := projectMatches[1]

			// For listing, we don't include a rate limit ID
			correctURL = fmt.Sprintf("%s/v1/organization/projects/%s/rate_limits",
				baseWithoutV1, projectID)

			fmt.Printf("[URL-DEBUG] Constructed rate limit list URL: %s\n", correctURL)
		} else {
			// Fallback to standard URL joining
			return baseURL + pathPart + queryPart
		}

		// Include query parameters if they exist
		if strings.HasPrefix(queryPart, "?") {
			correctURL += queryPart
		}

		return correctURL
	}

	// Try to parse the base URL
	_, err := url.Parse(baseURL)
	if err != nil {
		// If we can't parse the URL, fall back to simple concatenation with v1 check
		baseWithoutV1 := strings.TrimSuffix(strings.TrimSuffix(baseURL, "/"), "/v1")
		pathWithV1 := pathPart
		if !strings.HasPrefix(pathWithV1, "/v1") {
			pathWithV1 = "/v1" + pathWithV1
		}
		result := baseWithoutV1 + pathWithV1 + queryPart
		fmt.Printf("[URL-DEBUG] Failed to parse base URL, using simple concatenation: %s\n", result)
		return result
	}

	// Handle OpenAI organization/projects paths specifically
	if strings.Contains(pathPart, "/organization/projects/") ||
		strings.Contains(pathPart, "/v1/organization/projects/") {
		// Ensure the base URL doesn't have trailing /v1 and the path begins with /v1
		trimmedBase := strings.TrimSuffix(baseURL, "/v1")
		trimmedBase = strings.TrimSuffix(trimmedBase, "/")

		// Make sure the path starts with /v1 but remove any duplicate /v1 occurrences
		if strings.HasPrefix(pathPart, "/v1") {
			pathPart = strings.TrimPrefix(pathPart, "/v1")
		}
		if !strings.HasPrefix(pathPart, "/") {
			pathPart = "/" + pathPart
		}

		// Join them properly
		result := trimmedBase + "/v1" + pathPart + queryPart
		fmt.Printf("[URL-DEBUG] Constructed organization/projects URL: %s\n", result)
		return result
	}

	// For paths with rate_limits, ensure proper construction
	if strings.Contains(pathPart, "rate_limits") {
		// Ensure we don't duplicate any parts
		trimmedBase := strings.TrimSuffix(baseURL, "/v1")
		trimmedBase = strings.TrimSuffix(trimmedBase, "/")

		// Extract clean path without any duplicate v1 prefixes
		cleanPath := pathPart
		if strings.HasPrefix(cleanPath, "/v1") {
			cleanPath = strings.TrimPrefix(cleanPath, "/v1")
		}
		if !strings.HasPrefix(cleanPath, "/") {
			cleanPath = "/" + cleanPath
		}

		// Check if this is a rate limit URL for a specific project
		projectPattern := `/organization/projects/([^/]+)/rate_limits`
		if match := regexp.MustCompile(projectPattern).FindStringSubmatch(cleanPath); match != nil {
			projectID := match[1]
			// Always use the clean endpoint with no additional path components
			cleanPath = fmt.Sprintf("/organization/projects/%s/rate_limits", projectID)

			// Only add query parameters if they start with "?"
			if strings.HasPrefix(queryPart, "?") {
				cleanPath += queryPart
				queryPart = "" // Clear query part since we've incorporated it
			}

			// Log the reconstructed path for debugging
			fmt.Printf("[URL-DEBUG] Reconstructed rate_limits path: %s\n", cleanPath)
		}

		// Join the path with the base URL, ensuring just one /v1 prefix
		result := trimmedBase + "/v1" + cleanPath + queryPart
		fmt.Printf("[URL-DEBUG] Constructed rate_limits URL: %s\n", result)
		return result
	}

	// For other cases, ensure no duplicate v1 in path
	cleanPath := pathPart
	if strings.HasPrefix(cleanPath, "/v1") {
		cleanPath = strings.TrimPrefix(cleanPath, "/v1")
	}

	// Use standard URL joining with clean paths
	trimmedBase := strings.TrimSuffix(baseURL, "/v1")
	trimmedBase = strings.TrimSuffix(trimmedBase, "/")

	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	result := trimmedBase + "/v1" + cleanPath + queryPart
	fmt.Printf("[URL-DEBUG] Standard URL joining result: %s\n", result)
	return result
}

// CreateModelResponse creates a model response using the OpenAI API
func (c *OpenAIClient) CreateModelResponse(request *ModelResponseRequest) (*ModelResponse, error) {
	fmt.Printf("\n\n[CREATEMODEL-DEBUG] ========== CREATE MODEL RESPONSE DEBUG ==========\n")
	fmt.Printf("[CREATEMODEL-DEBUG] Function called: CreateModelResponse\n")
	fmt.Printf("[CREATEMODEL-DEBUG] Initial base URL: %s\n", c.APIURL)
	fmt.Printf("[CREATEMODEL-DEBUG] Function address: %p\n", c.CreateModelResponse)
	fmt.Printf("[CREATEMODEL-DEBUG] DoRequest address: %p\n", c.DoRequest)

	// Ensure we use the correct endpoint - always use /v1/responses for model response
	path := "/v1/responses"
	fmt.Printf("[CREATEMODEL-DEBUG] Using path: %s\n", path)

	// Print stack trace to find caller
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	fmt.Printf("[CREATEMODEL-DEBUG] Stack trace:\n%s\n", buf[:n])

	// HARDCODED URL TEST
	fullURL := c.APIURL
	if !strings.HasSuffix(fullURL, "/") {
		fullURL += "/"
	}
	fullURL += "v1/responses"
	fmt.Printf("[CREATEMODEL-DEBUG] Hardcoded full URL: %s\n", fullURL)

	// Make the request using DoRequest
	fmt.Printf("[CREATEMODEL-DEBUG] About to call DoRequest with POST and path=%s\n", path)
	responseBody, err := c.DoRequest("POST", path, request)
	if err != nil {
		fmt.Printf("[CREATEMODEL-DEBUG] Error from DoRequest: %s\n", err)
		return nil, fmt.Errorf("error creating model response: %s", err)
	}

	// Parse the response
	var response ModelResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		fmt.Printf("[CREATEMODEL-DEBUG] Error parsing response: %s\n", err)
		return nil, fmt.Errorf("error parsing response: %s", err)
	}

	fmt.Printf("[CREATEMODEL-DEBUG] Successfully parsed response with ID: %s\n", response.ID)
	fmt.Printf("[CREATEMODEL-DEBUG] ========== END CREATE MODEL RESPONSE DEBUG ==========\n\n")
	return &response, nil
}

// InviteRequest represents the request to invite a user to the organization
type InviteRequest struct {
	Email    string          `json:"email"`
	Role     string          `json:"role"`
	Projects []InviteProject `json:"projects,omitempty"`
}

// InviteProject represents a project assignment for an invited user
type InviteProject struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// Invite represents an invitation to an OpenAI organization
type Invite struct {
	ID        string          `json:"id"`
	Object    string          `json:"object"`
	Email     string          `json:"email"`
	Role      string          `json:"role"`
	Status    string          `json:"status"`
	ExpiresAt int64           `json:"expires_at"`
	CreatedAt int64           `json:"created_at"`
	Projects  []InviteProject `json:"projects,omitempty"`
}

// InviteResponse represents the API response when creating an invite
type InviteResponse struct {
	ID        string          `json:"id"`
	Object    string          `json:"object"`
	Email     string          `json:"email"`
	Role      string          `json:"role"`
	Status    string          `json:"status"`
	ExpiresAt int64           `json:"expires_at"`
	CreatedAt int64           `json:"created_at"`
	Projects  []InviteProject `json:"projects,omitempty"`
}

// ListInvitesResponse represents the API response when listing invites
type ListInvitesResponse struct {
	Object  string   `json:"object"`
	Data    []Invite `json:"data"`
	HasMore bool     `json:"has_more"`
}

// CreateInvite sends an invitation to a user to join the organization
func (c *OpenAIClient) CreateInvite(email, role string, projects []InviteProject, customAPIKey string) (*InviteResponse, error) {
	inviteRequest := &InviteRequest{
		Email:    email,
		Role:     role,
		Projects: projects,
	}

	// Prepare URL for the API request
	url := "/organization/invites"

	var respBody []byte
	var err error

	// Use custom API key if provided, otherwise use the default
	if customAPIKey != "" {
		respBody, err = c.doRequestWithCustomAPIKey("POST", url, inviteRequest, customAPIKey)
	} else {
		respBody, err = c.DoRequest("POST", url, inviteRequest)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating invite: %s", err)
	}

	// Parse the response
	var inviteResponse InviteResponse
	err = json.Unmarshal(respBody, &inviteResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing invite response: %s", err)
	}

	return &inviteResponse, nil
}

// GetInvite retrieves an invitation by ID
func (c *OpenAIClient) GetInvite(inviteID string, customAPIKey string) (*Invite, error) {
	// Prepare URL for the API request
	url := fmt.Sprintf("/v1/organization/invites/%s", inviteID)

	var respBody []byte
	var err error

	// Use custom API key if provided, otherwise use the default
	if customAPIKey != "" {
		respBody, err = c.doRequestWithCustomAPIKey("GET", url, nil, customAPIKey)
	} else {
		respBody, err = c.DoRequest("GET", url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error getting invite: %s", err)
	}

	// Parse the response
	var invite Invite
	err = json.Unmarshal(respBody, &invite)
	if err != nil {
		return nil, fmt.Errorf("error parsing invite response: %s", err)
	}

	return &invite, nil
}

// ListInvites retrieves all pending invitations for the organization
func (c *OpenAIClient) ListInvites(customAPIKey string) (*ListInvitesResponse, error) {
	// Create a client with extended timeout specifically for this operation
	// which can be slow for organizations with many invites
	httpClient := &http.Client{
		Timeout: 2 * time.Minute, // Increase timeout to 2 minutes
	}

	// Save the original HTTP client to restore it later
	originalClient := c.HTTPClient
	c.HTTPClient = httpClient
	defer func() {
		// Restore the original HTTP client when done
		c.HTTPClient = originalClient
	}()

	// Prepare URL for the API request
	url := "/v1/organization/invites"

	var respBody []byte
	var err error

	// Use custom API key if provided, otherwise use the default
	if customAPIKey != "" {
		respBody, err = c.doRequestWithCustomAPIKey("GET", url, nil, customAPIKey)
	} else {
		respBody, err = c.DoRequest("GET", url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error listing invites: %s", err)
	}

	// Parse the response
	var listResponse ListInvitesResponse
	err = json.Unmarshal(respBody, &listResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing invite list response: %s", err)
	}

	return &listResponse, nil
}

// DeleteInvite cancels an invitation
func (c *OpenAIClient) DeleteInvite(inviteID string, customAPIKey string) error {
	// Prepare URL for the API request
	url := fmt.Sprintf("/v1/organization/invites/%s", inviteID)

	var err error

	// Use custom API key if provided, otherwise use the default
	if customAPIKey != "" {
		_, err = c.doRequestWithCustomAPIKey("DELETE", url, nil, customAPIKey)
	} else {
		_, err = c.DoRequest("DELETE", url, nil)
	}

	if err != nil {
		// Check if the error is due to the invitation already being accepted
		if strings.Contains(err.Error(), "already accepted") {
			// If the invite is already accepted, we consider it deleted for Terraform purposes
			// This is because an accepted invitation has served its purpose in the workflow
			return nil
		}
		return fmt.Errorf("error deleting invite: %s", err)
	}

	return nil
}

// ListRateLimits retrieves all rate limits for a specific project.
func (c *OpenAIClient) ListRateLimits(projectID string) (*RateLimitListResponse, error) {
	return c.ListRateLimitsWithKey(projectID, "")
}

// ListRateLimitsWithKey retrieves all rate limits for a specific project using a custom API key.
func (c *OpenAIClient) ListRateLimitsWithKey(projectID, customAPIKey string) (*RateLimitListResponse, error) {
	// Use consistent URL format with other methods, avoiding the absolute URL approach
	url := fmt.Sprintf("/organization/projects/%s/rate_limits", projectID)

	// Debug info about the request
	fmt.Printf("\n\n[LIST-RL-DEBUG] ========== LIST RATE LIMITS DEBUG ==========\n")
	fmt.Printf("[LIST-RL-DEBUG] ProjectID: %s\n", projectID)
	maskedKey := "*****"
	if customAPIKey != "" && len(customAPIKey) > 5 {
		maskedKey = customAPIKey[:5] + "*****"
	}
	fmt.Printf("[LIST-RL-DEBUG] Using API key (masked): %s\n", maskedKey)
	fmt.Printf("[LIST-RL-DEBUG] Client API URL: %s\n", c.APIURL)
	fmt.Printf("[LIST-RL-DEBUG] Organization ID: %s\n", c.OrganizationID)
	fmt.Printf("[LIST-RL-DEBUG] API URL: %s\n", url)

	var respBody []byte
	var err error
	if customAPIKey != "" {
		respBody, err = c.doRequestWithCustomAPIKey("GET", url, nil, customAPIKey)
	} else {
		respBody, err = c.doRequest("GET", url, nil)
	}

	if err != nil {
		fmt.Printf("[LIST-RL-DEBUG] Error listing rate limits: %v\n", err)
		fmt.Printf("[LIST-RL-DEBUG] ========== END LIST RATE LIMITS DEBUG ==========\n\n")
		return nil, fmt.Errorf("failed to list rate limits: %w", err)
	}

	var response RateLimitListResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		fmt.Printf("[LIST-RL-DEBUG] Failed to parse rate limits response: %v\n", err)
		fmt.Printf("[LIST-RL-DEBUG] ========== END LIST RATE LIMITS DEBUG ==========\n\n")
		return nil, fmt.Errorf("failed to parse rate limits response: %w", err)
	}

	// Print all rate limits for debugging
	fmt.Printf("[LIST-RL-DEBUG] Successfully listed %d rate limits\n", len(response.Data))
	for i, rl := range response.Data {
		fmt.Printf("[LIST-RL-DEBUG] Rate Limit #%d - ID: %s, Model: %s, MaxRequests: %d, MaxTokens: %d\n",
			i+1, rl.ID, rl.Model, rl.MaxRequestsPer1Minute, rl.MaxTokensPer1Minute)
	}
	fmt.Printf("[LIST-RL-DEBUG] ========== END LIST RATE LIMITS DEBUG ==========\n\n")

	return &response, nil
}

// DeleteRateLimitWithKey deletes a rate limit for a model in a project using a custom API key.
// NOTE: This is not a true deletion since OpenAI doesn't support DELETE operations on rate limits.
// Instead, this resets the rate limit to default values.
func (c *OpenAIClient) DeleteRateLimitWithKey(projectID, modelOrRateLimitID string, customAPIKey string) error {
	// Note: OpenAI doesn't support DELETE operations on rate limits.
	// Instead we "reset" them to default values.
	fmt.Printf("[DELETE-RL-DEBUG] ========== DeleteRateLimitWithKey DEBUG ==========\n")
	fmt.Printf("[DELETE-RL-DEBUG] ProjectID: %s\n", projectID)
	fmt.Printf("[DELETE-RL-DEBUG] ModelOrRateLimitID: %s\n", modelOrRateLimitID)

	// Step 1: List all rate limits and find the one we want to reset
	fmt.Printf("[DELETE-RL-DEBUG] Listing all rate limits to find target rate limit\n")
	rateLimits, err := c.ListRateLimitsWithKey(projectID, customAPIKey)
	if err != nil {
		fmt.Printf("[DELETE-RL-DEBUG] Failed to list rate limits: %v\n", err)
		return fmt.Errorf("failed to list rate limits: %w", err)
	}

	// Search logic similar to GetRateLimitWithKey
	model := ""
	searchID := modelOrRateLimitID

	if !strings.HasPrefix(modelOrRateLimitID, "rl-") {
		// If it doesn't have "rl-" prefix, it's a model name
		model = modelOrRateLimitID
		searchID = "rl-" + modelOrRateLimitID
		fmt.Printf("[DELETE-RL-DEBUG] Searching by MODEL: %s (ID would be %s)\n", model, searchID)
	} else {
		// It has the prefix, treat it as an ID but also extract the model
		// Try to extract model name from the ID
		parts := strings.Split(modelOrRateLimitID[3:], "-")
		if len(parts) >= 1 {
			model = parts[0]
			if len(parts) > 1 {
				// For multi-part model names like "gpt-4o-mini"
				if !containsProjectSuffix(parts[len(parts)-1]) {
					model = strings.Join(parts, "-")
				} else {
					model = strings.Join(parts[:len(parts)-1], "-")
				}
			}
		}
		fmt.Printf("[DELETE-RL-DEBUG] Searching by ID: %s (extracted model: %s)\n", searchID, model)
	}

	// Find the target rate limit
	var targetRateLimit *RateLimit

	// 1. Try exact ID match first
	for _, rl := range rateLimits.Data {
		if rl.ID == searchID {
			fmt.Printf("[DELETE-RL-DEBUG] Found exact ID match: %s (Model: %s)\n", rl.ID, rl.Model)
			targetRateLimit = &rl
			break
		}
	}

	// 2. Try ID prefix match (accounts for project-specific suffixes)
	if targetRateLimit == nil {
		for _, rl := range rateLimits.Data {
			if strings.HasPrefix(rl.ID, searchID+"-") || rl.ID == searchID {
				fmt.Printf("[DELETE-RL-DEBUG] Found ID prefix match: %s (Model: %s)\n", rl.ID, rl.Model)
				targetRateLimit = &rl
				break
			}
		}
	}

	// 3. Try model name match as fallback
	if targetRateLimit == nil && model != "" {
		for _, rl := range rateLimits.Data {
			if rl.Model == model {
				fmt.Printf("[DELETE-RL-DEBUG] Found by model name match: %s (ID: %s)\n", rl.Model, rl.ID)
				targetRateLimit = &rl
				break
			}
		}
	}

	// 4. Last resort: try partial model matching for compound model names
	if targetRateLimit == nil && model != "" && strings.Contains(model, "-") {
		modelBase := strings.Split(model, "-")[0]
		for _, rl := range rateLimits.Data {
			if strings.HasPrefix(rl.Model, modelBase) {
				fmt.Printf("[DELETE-RL-DEBUG] Found by partial model name match: %s matches base %s (ID: %s)\n",
					rl.Model, modelBase, rl.ID)
				targetRateLimit = &rl
				break
			}
		}
	}

	if targetRateLimit == nil {
		fmt.Printf("[DELETE-RL-DEBUG] Rate limit not found for ID/model: %s\n", modelOrRateLimitID)
		fmt.Printf("[DELETE-RL-DEBUG] ========== END DeleteRateLimitWithKey DEBUG ==========\n\n")
		return fmt.Errorf("API error: Project with ID '%s' not found or rate limit '%s' does not exist",
			projectID, modelOrRateLimitID)
	}

	// Now we have the target rate limit, construct the URL for update
	effectiveRateLimitID := targetRateLimit.ID
	effectiveModel := targetRateLimit.Model
	fmt.Printf("[DELETE-RL-DEBUG] Found rate limit to reset: %s (Model: %s)\n", effectiveRateLimitID, effectiveModel)

	// Construct the direct path with the rate limit ID included
	path := fmt.Sprintf("/organization/projects/%s/rate_limits/%s", projectID, effectiveRateLimitID)
	fmt.Printf("[DELETE-RL-DEBUG] Using path for reset: %s\n", path)

	// Get default values for this model
	defaultValues := getDefaultRateLimitValues(effectiveModel)
	fmt.Printf("[DELETE-RL-DEBUG] Default values for model %s: %+v\n", effectiveModel, defaultValues)

	// Create the request body with default values
	req := map[string]interface{}{
		"max_requests_per_1_minute": defaultValues.MaxRequestsPer1Minute,
		"max_tokens_per_1_minute":   defaultValues.MaxTokensPer1Minute,
	}

	// Add optional fields if they exist in the default values
	if defaultValues.MaxImagesPer1Minute > 0 {
		req["max_images_per_1_minute"] = defaultValues.MaxImagesPer1Minute
	}
	if defaultValues.MaxAudioMegabytesPer1Minute > 0 {
		req["max_audio_megabytes_per_1_minute"] = defaultValues.MaxAudioMegabytesPer1Minute
	}
	if defaultValues.Batch1DayMaxInputTokens > 0 {
		req["batch_1_day_max_input_tokens"] = defaultValues.Batch1DayMaxInputTokens
	}
	if defaultValues.MaxRequestsPer1Day > 0 {
		req["max_requests_per_1_day"] = defaultValues.MaxRequestsPer1Day
	}

	// Log the request for debugging
	reqJson, _ := json.Marshal(req)
	fmt.Printf("[DELETE-RL-DEBUG] Request body with default values: %s\n", string(reqJson))

	// Debug: Output the actual API key being used (first/last 4 chars)
	apiKeyToUse := c.APIKey
	if customAPIKey != "" {
		apiKeyToUse = customAPIKey
	}
	maskedKey := ""
	if len(apiKeyToUse) > 8 {
		maskedKey = apiKeyToUse[:4] + "..." + apiKeyToUse[len(apiKeyToUse)-4:]
	} else {
		maskedKey = "***"
	}
	fmt.Printf("[DELETE-RL-DEBUG] Using API key: %s\n", maskedKey)

	// Send POST request to reset the rate limit to default values
	var body []byte
	if customAPIKey != "" {
		fmt.Printf("[DELETE-RL-DEBUG] Using custom API key for request\n")
		body, err = c.doRequestWithCustomAPIKey(http.MethodPost, path, req, customAPIKey)
	} else {
		fmt.Printf("[DELETE-RL-DEBUG] Using default API key for request\n")
		body, err = c.doRequest(http.MethodPost, path, req)
	}

	if err != nil {
		fmt.Printf("[DELETE-RL-DEBUG] Reset rate limit request failed: %v\n", err)
		fmt.Printf("[DELETE-RL-DEBUG] ========== END DeleteRateLimitWithKey DEBUG ==========\n\n")
		return err
	}

	// Debug: Log the response
	fmt.Printf("[DELETE-RL-DEBUG] Response from OpenAI API: %s\n", string(body))
	fmt.Printf("[DELETE-RL-DEBUG] ========== END DeleteRateLimitWithKey DEBUG ==========\n\n")
	return nil
}

// Helper function to extract the model name from a rate limit ID
func extractModelFromRateLimitID(rateLimitID string) string {
	if !strings.HasPrefix(rateLimitID, "rl-") {
		return rateLimitID // Not a rate limit ID, return as is
	}

	// Rate limit IDs typically follow the pattern: rl-model[-model-parts][-project-suffix]
	parts := strings.Split(rateLimitID, "-")
	if len(parts) < 3 || parts[0] != "rl" {
		return strings.TrimPrefix(rateLimitID, "rl-") // Simple case
	}

	// For complex model names with hyphens, we need to determine
	// which parts belong to the model and which might be a project suffix

	// If it ends with a known project suffix pattern (typically 8 chars),
	// remove that part. Otherwise, take everything after "rl-"
	if len(parts) >= 4 && len(parts[len(parts)-1]) == 8 {
		// Return everything except the first and last parts
		modelParts := parts[1 : len(parts)-1]
		return strings.Join(modelParts, "-")
	}

	// Default case - just remove the "rl-" prefix
	return strings.TrimPrefix(rateLimitID, "rl-")
}

// defaultRateLimits contains the default rate limit values for each model
var defaultRateLimits = map[string]struct {
	MaxRequestsPer1Minute       int
	MaxTokensPer1Minute         int
	MaxImagesPer1Minute         int
	Batch1DayMaxInputTokens     int
	MaxAudioMegabytesPer1Minute int
	MaxRequestsPer1Day          int
}{
	"babbage-002": {
		MaxRequestsPer1Minute: 3000,
		MaxTokensPer1Minute:   250000,
		MaxImagesPer1Minute:   10,
	},
	"chatgpt-4o-latest": {
		MaxRequestsPer1Minute: 200,
		MaxTokensPer1Minute:   500000,
	},
	"computer-use-preview": {
		MaxRequestsPer1Minute:   3000,
		MaxTokensPer1Minute:     20000000,
		MaxImagesPer1Minute:     20000,
		Batch1DayMaxInputTokens: 450000000,
	},
	"computer-use-preview-2025-03-11": {
		MaxRequestsPer1Minute:   3000,
		MaxTokensPer1Minute:     20000000,
		MaxImagesPer1Minute:     20000,
		Batch1DayMaxInputTokens: 450000000,
	},
	"dall-e-2": {
		MaxRequestsPer1Minute: 7500,
		MaxTokensPer1Minute:   2147483647,
		MaxImagesPer1Minute:   100,
	},
	"dall-e-3": {
		MaxRequestsPer1Minute: 7500,
		MaxTokensPer1Minute:   2147483647,
		MaxImagesPer1Minute:   15,
	},
	"davinci-002": {
		MaxRequestsPer1Minute: 3000,
		MaxTokensPer1Minute:   250000,
		MaxImagesPer1Minute:   10,
	},
	"default": {
		MaxRequestsPer1Minute: 3000,
		MaxTokensPer1Minute:   250000,
		MaxImagesPer1Minute:   10,
	},
	"ft:babbage-002": {
		MaxRequestsPer1Minute: 3000,
		MaxTokensPer1Minute:   250000,
		MaxImagesPer1Minute:   10,
	},
	"ft:davinci-002": {
		MaxRequestsPer1Minute: 3000,
		MaxTokensPer1Minute:   250000,
		MaxImagesPer1Minute:   10,
	},
	"ft:gpt-3.5-turbo-0125": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"ft:gpt-3.5-turbo-0613": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"ft:gpt-3.5-turbo-1106": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"ft:	-0613": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     300000,
		Batch1DayMaxInputTokens: 30000000,
	},
	"ft:gpt-4o-2024-05-13": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"ft:gpt-4o-mini-2024-07-18": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-3.5-turbo": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-3.5-turbo-0125": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-3.5-turbo-1106": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-3.5-turbo-16k": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-3.5-turbo-instruct": {
		MaxRequestsPer1Minute:   3500,
		MaxTokensPer1Minute:     90000,
		MaxImagesPer1Minute:     2147483647,
		Batch1DayMaxInputTokens: 200000,
	},
	"gpt-3.5-turbo-instruct-0914": {
		MaxRequestsPer1Minute:   3500,
		MaxTokensPer1Minute:     90000,
		MaxImagesPer1Minute:     2147483647,
		Batch1DayMaxInputTokens: 200000,
	},
	"gpt-4": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     300000,
		Batch1DayMaxInputTokens: 30000000,
	},
	"gpt-4-0125-preview": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     300000,
		Batch1DayMaxInputTokens: 30000000,
	},
	"gpt-4-0613": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     300000,
		Batch1DayMaxInputTokens: 30000000,
	},
	"gpt-4-1106-preview": {
		MaxRequestsPer1Minute: 10000,
		MaxTokensPer1Minute:   450000,
		MaxRequestsPer1Day:    2147483647,
	},
	"gpt-4-turbo": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     800000,
		MaxImagesPer1Minute:     10000,
		Batch1DayMaxInputTokens: 80000000,
	},
	"gpt-4-turbo-2024-04-09": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     800000,
		MaxImagesPer1Minute:     10000,
		Batch1DayMaxInputTokens: 80000000,
	},
	"gpt-4-turbo-preview": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     800000,
		MaxImagesPer1Minute:     10000,
		Batch1DayMaxInputTokens: 80000000,
	},
	"gpt-4.1": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4.1-2025-04-14": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4.1-long-context": {
		MaxRequestsPer1Minute:   1000,
		MaxTokensPer1Minute:     5000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 100000000,
	},
	"gpt-4.1-mini": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-4.1-mini-2025-04-14": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-4.1-mini-long-context": {
		MaxRequestsPer1Minute:   2000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4.1-nano": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-4.1-nano-2025-04-14": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-4.1-nano-long-context": {
		MaxRequestsPer1Minute:   2000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4.5-preview": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     1000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 100000000,
	},
	"gpt-4.5-preview-2025-02-27": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     1000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 100000000,
	},
	"gpt-4o": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4o-2024-05-13": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4o-2024-08-06": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4o-2024-11-20": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4o-audio-preview": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4o-audio-preview-2024-10-01": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4o-audio-preview-2024-12-17": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"gpt-4o-mini": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-4o-mini-2024-07-18": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-4o-mini-audio-preview": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-4o-mini-audio-preview-2024-12-17": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"gpt-4o-mini-realtime-preview": {
		MaxRequestsPer1Minute: 10000,
		MaxTokensPer1Minute:   4000000,
	},
	"gpt-4o-mini-realtime-preview-2024-12-17": {
		MaxRequestsPer1Minute: 10000,
		MaxTokensPer1Minute:   4000000,
	},
	"gpt-4o-mini-search-preview": {
		MaxRequestsPer1Minute: 1000,
		MaxTokensPer1Minute:   200000,
	},
	"gpt-4o-mini-search-preview-2025-03-11": {
		MaxRequestsPer1Minute: 1000,
		MaxTokensPer1Minute:   200000,
	},
	"gpt-4o-mini-transcribe": {
		MaxRequestsPer1Minute: 10000,
		MaxTokensPer1Minute:   2000000,
	},
	"gpt-4o-mini-tts": {
		MaxRequestsPer1Minute: 10000,
		MaxTokensPer1Minute:   2000000,
	},
	"gpt-4o-realtime-preview": {
		MaxRequestsPer1Minute: 10000,
		MaxTokensPer1Minute:   4000000,
	},
	"gpt-4o-realtime-preview-2024-10-01": {
		MaxRequestsPer1Minute: 10000,
		MaxTokensPer1Minute:   4000000,
	},
	"gpt-4o-realtime-preview-2024-12-17": {
		MaxRequestsPer1Minute: 3000,
		MaxTokensPer1Minute:   250000,
		MaxImagesPer1Minute:   10,
	},
	"gpt-4o-search-preview": {
		MaxRequestsPer1Minute: 1000,
		MaxTokensPer1Minute:   200000,
	},
	"gpt-4o-search-preview-2025-03-11": {
		MaxRequestsPer1Minute: 1000,
		MaxTokensPer1Minute:   200000,
	},
	"gpt-4o-transcribe": {
		MaxRequestsPer1Minute: 10000,
		MaxTokensPer1Minute:   2000000,
	},
	"o1": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"o1-2024-12-17": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"o1-mini": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"o1-mini-2024-09-12": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"o1-preview": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"o1-preview-2024-09-12": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"o1-pro": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"o1-pro-2025-03-19": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"o3": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"o3-2025-04-16": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     2000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 200000000,
	},
	"o3-mini": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"o3-mini-2025-01-31": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"o4-mini": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"o4-mini-2025-04-16": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     10000000,
		MaxImagesPer1Minute:     50000,
		Batch1DayMaxInputTokens: 1000000000,
	},
	"omni-moderation-2024-09-26": {
		MaxRequestsPer1Minute: 2000,
		MaxTokensPer1Minute:   250000,
	},
	"omni-moderation-latest": {
		MaxRequestsPer1Minute: 2000,
		MaxTokensPer1Minute:   250000,
	},
	"text-embedding-3-large": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     5000000,
		Batch1DayMaxInputTokens: 500000000,
	},
	"text-embedding-3-small": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     5000000,
		Batch1DayMaxInputTokens: 500000000,
	},
	"text-embedding-ada-002": {
		MaxRequestsPer1Minute:   10000,
		MaxTokensPer1Minute:     5000000,
		Batch1DayMaxInputTokens: 500000000,
	},
	"text-moderation-latest": {
		MaxRequestsPer1Minute: 1000,
		MaxTokensPer1Minute:   150000,
	},
	"text-moderation-stable": {
		MaxRequestsPer1Minute: 1000,
		MaxTokensPer1Minute:   150000,
	},
	"tts-1": {
		MaxRequestsPer1Minute: 7500,
		MaxTokensPer1Minute:   2147483647,
	},
	"tts-1-1106": {
		MaxRequestsPer1Minute: 7500,
		MaxTokensPer1Minute:   2147483647,
	},
	"tts-1-hd": {
		MaxRequestsPer1Minute: 7500,
		MaxTokensPer1Minute:   2147483647,
		MaxImagesPer1Minute:   2147483647,
	},
	"tts-1-hd-1106": {
		MaxRequestsPer1Minute: 7500,
		MaxTokensPer1Minute:   2147483647,
		MaxImagesPer1Minute:   2147483647,
	},
	"whisper-1": {
		MaxRequestsPer1Minute: 7500,
		MaxTokensPer1Minute:   2147483647,
	},
}

// Helper function to get default rate limit values based on the model
func getDefaultRateLimitValues(model string) *RateLimit {
	// Look up the model in the defaults map
	defaults, ok := defaultRateLimits[model]

	// If not found, try the "default" entry
	if !ok {
		defaults, ok = defaultRateLimits["default"]

		// If still not found, use fallback values
		if !ok {
			return &RateLimit{
				Model:                       model,
				MaxRequestsPer1Minute:       1000000, // Very high value to effectively make it unlimited
				MaxTokensPer1Minute:         1000000, // Very high value to effectively make it unlimited
				MaxImagesPer1Minute:         1000000, // Very high value to effectively make it unlimited
				Batch1DayMaxInputTokens:     1000000, // Very high value to effectively make it unlimited
				MaxAudioMegabytesPer1Minute: 1000000, // Very high value to effectively make it unlimited
				MaxRequestsPer1Day:          1000000, // Very high value to effectively make it unlimited
			}
		}
	}

	// Return values from the defaults map
	return &RateLimit{
		Model:                       model,
		MaxRequestsPer1Minute:       defaults.MaxRequestsPer1Minute,
		MaxTokensPer1Minute:         defaults.MaxTokensPer1Minute,
		MaxImagesPer1Minute:         defaults.MaxImagesPer1Minute,
		Batch1DayMaxInputTokens:     defaults.Batch1DayMaxInputTokens,
		MaxAudioMegabytesPer1Minute: defaults.MaxAudioMegabytesPer1Minute,
		MaxRequestsPer1Day:          defaults.MaxRequestsPer1Day,
	}
}

// TestNetworkConnectivity tests if we can connect to the OpenAI API
func (c *OpenAIClient) TestNetworkConnectivity() error {
	fmt.Printf("[NETWORK-TEST] Testing network connectivity to OpenAI API\n")

	// Parse the API URL to get the host
	parsedURL, err := url.Parse(c.APIURL)
	if err != nil {
		return fmt.Errorf("failed to parse API URL: %v", err)
	}

	host := parsedURL.Host
	fmt.Printf("[NETWORK-TEST] Testing connectivity to host: %s\n", host)

	// Try DNS resolution first
	ips, err := net.LookupIP(host)
	if err != nil {
		fmt.Printf("[NETWORK-TEST] DNS lookup failed: %v\n", err)
		return fmt.Errorf("DNS resolution failed for %s: %v", host, err)
	}

	fmt.Printf("[NETWORK-TEST] DNS resolution successful. IPs: %v\n", ips)

	// Try establishing a TCP connection to port 443 (HTTPS)
	conn, err := net.DialTimeout("tcp", host+":443", 10*time.Second)
	if err != nil {
		fmt.Printf("[NETWORK-TEST] TCP connection failed: %v\n", err)
		return fmt.Errorf("failed to establish TCP connection to %s:443: %v", host, err)
	}
	defer conn.Close()

	fmt.Printf("[NETWORK-TEST] TCP connection successful\n")

	// Make a basic HEAD request to check HTTP connectivity
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	req, err := http.NewRequest("HEAD", c.APIURL, nil)
	if err != nil {
		fmt.Printf("[NETWORK-TEST] Failed to create HEAD request: %v\n", err)
		return fmt.Errorf("failed to create HEAD request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[NETWORK-TEST] HEAD request failed: %v\n", err)
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("[NETWORK-TEST] HTTP HEAD request successful. Status code: %d\n", resp.StatusCode)
	return nil
}
