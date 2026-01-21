package provider

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
