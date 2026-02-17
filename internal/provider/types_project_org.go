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
	CreatedAt int64  `json:"created_at"`
}

// ProjectGroupListResponse represents the list response for project groups.
type ProjectGroupListResponse struct {
	Object  string                          `json:"object"`
	Data    []ProjectGroupResponseFramework `json:"data"`
	HasMore bool                            `json:"has_more"`
	Next    *string                         `json:"next"`
}

// GroupResponseFramework represents the API response for an organization group.
type GroupResponseFramework struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	CreatedAt     int64  `json:"created_at"`
	IsSCIMManaged bool   `json:"is_scim_managed"`
}

// GroupListResponse represents the list response for organization groups.
type GroupListResponse struct {
	Object  string                   `json:"object"`
	Data    []GroupResponseFramework `json:"data"`
	HasMore bool                     `json:"has_more"`
	Next    *string                  `json:"next"`
}

// GroupCreateRequest represents the request to create a group.
type GroupCreateRequest struct {
	Name string `json:"name"`
}

// GroupUpdateRequest represents the request to update a group.
type GroupUpdateRequest struct {
	Name string `json:"name"`
}

// GroupUserResponseFramework represents the API response for a user in a group.
type GroupUserResponseFramework struct {
	ID               string  `json:"id"`
	Email            string  `json:"email"`
	Name             string  `json:"name"`
	IsServiceAccount bool    `json:"is_service_account"`
	Picture          *string `json:"picture"`
}

// GroupUserListResponse represents the list response for group users.
type GroupUserListResponse struct {
	Object  string                       `json:"object"`
	Data    []GroupUserResponseFramework `json:"data"`
	HasMore bool                         `json:"has_more"`
	Next    *string                      `json:"next"`
}

// GroupUserCreateRequest represents the request to add a user to a group.
type GroupUserCreateRequest struct {
	UserID string `json:"user_id"`
}

// RoleResponseFramework represents the API response for a role.
type RoleResponseFramework struct {
	Object         string   `json:"object"`
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    *string  `json:"description"`
	Permissions    []string `json:"permissions"`
	ResourceType   string   `json:"resource_type"`
	PredefinedRole bool     `json:"predefined_role"`
}

// RoleListResponse represents the list response for roles.
type RoleListResponse struct {
	Object  string                  `json:"object"`
	Data    []RoleResponseFramework `json:"data"`
	HasMore bool                    `json:"has_more"`
	Next    *string                 `json:"next"`
}

// GroupRoleAssignment represents a role assigned to a group.
type GroupRoleAssignment struct {
	Object   string                `json:"object"`
	ID       string                `json:"id"`
	Role     RoleResponseFramework `json:"role"`
	Group    GroupInfo             `json:"group"`
	Metadata map[string]string     `json:"metadata"`
}

// GroupInfo represents basic group information in role assignments.
type GroupInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RoleCreateRequest represents the request to create a custom role.
type RoleCreateRequest struct {
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
	Description string   `json:"description,omitempty"`
}

// RoleUpdateRequest represents the request to update a custom role.
type RoleUpdateRequest struct {
	RoleName    string   `json:"role_name"`
	Permissions []string `json:"permissions"`
	Description string   `json:"description,omitempty"`
}

// RoleAssignRequest represents the request to assign a role.
type RoleAssignRequest struct {
	RoleID string `json:"role_id"`
}

// UserRoleAssignment represents a role assigned to a user.
type UserRoleAssignment struct {
	Object   string                `json:"object"`
	ID       string                `json:"id"`
	Role     RoleResponseFramework `json:"role"`
	User     UserInfo              `json:"user"`
	Metadata map[string]string     `json:"metadata"`
}

// UserInfo represents basic user information in role assignments.
type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
