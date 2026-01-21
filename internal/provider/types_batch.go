package provider

// BatchResponse represents the API response for batch operations.
type BatchResponse struct {
	ID               string                 `json:"id"`
	Object           string                 `json:"object"`
	Endpoint         string                 `json:"endpoint"`
	Errors           *BatchErrors           `json:"errors,omitempty"`
	InputFileID      string                 `json:"input_file_id"`
	CompletionWindow string                 `json:"completion_window"`
	Status           string                 `json:"status"`
	OutputFileID     string                 `json:"output_file_id"`
	ErrorFileID      string                 `json:"error_file_id"`
	CreatedAt        int64                  `json:"created_at"`
	InProgressAt     *int64                 `json:"in_progress_at"`
	ExpiresAt        int64                  `json:"expires_at"`
	FinalizingAt     *int64                 `json:"finalizing_at"`
	CompletedAt      *int64                 `json:"completed_at"`
	FailedAt         *int64                 `json:"failed_at"`
	ExpiredAt        *int64                 `json:"expired_at"`
	CancellingAt     *int64                 `json:"cancelling_at"`
	CancelledAt      *int64                 `json:"cancelled_at"`
	RequestCounts    *RequestCounts         `json:"request_counts"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type BatchErrors struct {
	Object string       `json:"object"`
	Data   []BatchError `json:"data"`
}

type BatchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
	Line    *int   `json:"line,omitempty"`
}

type RequestCounts struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// BatchCreateRequest represents the request payload for creating a new batch job.
type BatchCreateRequest struct {
	InputFileID      string                 `json:"input_file_id"`
	Endpoint         string                 `json:"endpoint"`
	CompletionWindow string                 `json:"completion_window"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}
