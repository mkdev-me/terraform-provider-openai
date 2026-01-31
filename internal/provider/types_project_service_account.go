package provider

// ProjectServiceAccountResponse represents the API response for a project service account.
type ProjectServiceAccountResponse struct {
	Object    string                               `json:"object"`
	ID        string                               `json:"id"`
	Name      string                               `json:"name"`
	Role      string                               `json:"role"`
	CreatedAt int64                                `json:"created_at"`
	APIKey    *ProjectServiceAccountAPIKeyResponse `json:"api_key,omitempty"`
}

type ProjectServiceAccountAPIKeyResponse struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	Value     string `json:"value"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

// ProjectServiceAccountCreateRequest represents the request to create a service account.
type ProjectServiceAccountCreateRequest struct {
	Name string `json:"name"`
}
