package provider

// ProjectResponseFramework represents the API response for an OpenAI project.
type ProjectResponseFramework struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	CreatedAt  int64  `json:"created_at"`
	ArchivedAt *int64 `json:"archived_at,omitempty"`
}

// ProjectUserResponseFramework represents the API response for a project user.
type ProjectUserResponseFramework struct {
	Object  string `json:"object"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	AddedAt int64  `json:"added_at"`
}

// OrganizationUserResponseFramework represents the API response for an organization user.
type OrganizationUserResponseFramework struct {
	Object  string `json:"object"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	AddedAt int64  `json:"added_at"`
}

// AdminAPIKeyResponseFramework represents the API response for an admin API key.
// Note: This duplicates the struct in the legacy resource, but we need it here for the framework resource.
// We'll eventually delete the legacy one.
type AdminAPIKeyResponseFramework struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	CreatedAt int64        `json:"created_at"`
	ExpiresAt *int64       `json:"expires_at,omitempty"` // null if never expires
	Object    string       `json:"object"`
	Scopes    []string     `json:"scopes,omitempty"`
	Key       string       `json:"key"` // Only returned on creation
	Owner     *APIKeyOwner `json:"owner,omitempty"`
}

type APIKeyOwner struct {
	Type           string                `json:"type"`
	User           *APIKeyUser           `json:"user,omitempty"`
	ServiceAccount *APIKeyServiceAccount `json:"service_account,omitempty"`
}

type APIKeyUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type APIKeyServiceAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ProjectCreateRequest represents the request to create a project.
type ProjectCreateRequest struct {
	Name string `json:"name"`
}

// ProjectUserUpdateRequest represents the request to update a project user.
type ProjectUserUpdateRequest struct {
	Role string `json:"role"`
}

// AdminAPIKeyCreateRequest represents the request to create an admin API key.
type AdminAPIKeyCreateRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes,omitempty"`
}

// ProjectGroupResponseFramework represents the API response for a project group.
type ProjectGroupResponseFramework struct {
	Object    string `json:"object"`
	ProjectID string `json:"project_id"`
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Role      string `json:"role"`
	CreatedAt int64  `json:"created_at"`
}

// ProjectGroupListResponse represents the list response for project groups.
type ProjectGroupListResponse struct {
	Object  string                          `json:"object"`
	Data    []ProjectGroupResponseFramework `json:"data"`
	FirstID string                          `json:"first_id"`
	LastID  string                          `json:"last_id"`
	HasMore bool                            `json:"has_more"`
}

// ProjectGroupCreateRequest represents the request to add a group to a project.
type ProjectGroupCreateRequest struct {
	GroupID string `json:"group_id"`
	Role    string `json:"role"`
}
