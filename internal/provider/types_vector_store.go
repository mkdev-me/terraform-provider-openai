package provider

// VectorStoreResponse represents a vector store object.
type VectorStoreResponse struct {
	ID               string                 `json:"id"`
	Object           string                 `json:"object"`
	CreatedAt        int64                  `json:"created_at"`
	Name             string                 `json:"name"`
	UsageBytes       int64                  `json:"usage_bytes"`
	FileCounts       *FileCounts            `json:"file_counts"`
	Status           string                 `json:"status"`
	ExpiresAfter     *ExpiresAfter          `json:"expires_after"`
	ExpiresAt        *int64                 `json:"expires_at,omitempty"`
	LastActiveAt     *int64                 `json:"last_active_at,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
	ChunkingStrategy *ChunkingStrategy      `json:"chunking_strategy,omitempty"` // For reading response defaults? Or just create param?
	// Note: file_ids is NOT returned in GET /vector_stores/{id} typically, only in create if provided?
	// Actually, API docs say "file_ids" are not in the response object directly, usually fetched via files endpoint.
	// But creation response might include it if simpler?
	// Terraform legacy code seemed to try to read it but mapped it to nil if not present.
}

type FileCounts struct {
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
	Cancelled  int `json:"cancelled"`
	Total      int `json:"total"`
}

type ExpiresAfter struct {
	Anchor string `json:"anchor"`
	Days   int    `json:"days"`
}

// VectorStoreCreateRequest represents the request to create a vector store.
type VectorStoreCreateRequest struct {
	Name             string                 `json:"name,omitempty"`
	FileIDs          []string               `json:"file_ids,omitempty"`
	ExpiresAfter     *ExpiresAfter          `json:"expires_after,omitempty"`
	ChunkingStrategy *ChunkingStrategy      `json:"chunking_strategy,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type ChunkingStrategy struct {
	Type   string          `json:"type"`
	Static *StaticChunking `json:"static,omitempty"`
	// Or it might be flattened? API says:
	// type: "static", static: { max_chunk_size_tokens, chunk_overlap_tokens, ... }
	// Legacy provider had "size" and "max_tokens" mixed up.
	// Static strategy fields: max_chunk_size_tokens (100-4096), chunk_overlap_tokens, etc.
}

type StaticChunking struct {
	MaxChunkSizeTokens int `json:"max_chunk_size_tokens"`
	ChunkOverlapTokens int `json:"chunk_overlap_tokens"`
}

// VectorStoreFileResponse represents a file inside a vector store.
type VectorStoreFileResponse struct {
	ID               string            `json:"id"`
	Object           string            `json:"object"`
	UsageBytes       int64             `json:"usage_bytes"`
	CreatedAt        int64             `json:"created_at"`
	VectorStoreID    string            `json:"vector_store_id"`
	Status           string            `json:"status"`
	LastError        *LastError        `json:"last_error,omitempty"`
	ChunkingStrategy *ChunkingStrategy `json:"chunking_strategy,omitempty"`
}

type LastError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// VectorStoreFileCreateRequest represents request to add a file.
type VectorStoreFileCreateRequest struct {
	FileID           string            `json:"file_id"`
	ChunkingStrategy *ChunkingStrategy `json:"chunking_strategy,omitempty"`
}

// VectorStoreFileBatchResponse represents a batch of files.
type VectorStoreFileBatchResponse struct {
	ID            string      `json:"id"`
	Object        string      `json:"object"`
	CreatedAt     int64       `json:"created_at"`
	VectorStoreID string      `json:"vector_store_id"`
	Status        string      `json:"status"`
	FileCounts    *FileCounts `json:"file_counts"`
}

// VectorStoreFileBatchCreateRequest represents request to add a batch of files.
type VectorStoreFileBatchCreateRequest struct {
	FileIDs          []string          `json:"file_ids"`
	ChunkingStrategy *ChunkingStrategy `json:"chunking_strategy,omitempty"`
}

// ListVectorStoresResponse represents the API response for listing vector stores.
type ListVectorStoresResponse struct {
	Object  string                `json:"object"`
	Data    []VectorStoreResponse `json:"data"`
	FirstID string                `json:"first_id"`
	LastID  string                `json:"last_id"`
	HasMore bool                  `json:"has_more"`
}
