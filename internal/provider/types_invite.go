package provider

// InviteResponse represents the API response for an invite.
type InviteResponse struct {
	ID         string          `json:"id"`
	Email      string          `json:"email"`
	Role       string          `json:"role"`
	Status     string          `json:"status"`
	CreatedAt  int64           `json:"created_at"`
	ExpiresAt  int64           `json:"expires_at"`
	AcceptedAt int64           `json:"accepted_at"`
	Projects   []InviteProject `json:"projects"`
}

type InviteProject struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// InviteCreateRequest represents the request to create an invite.
type InviteCreateRequest struct {
	Email    string          `json:"email"`
	Role     string          `json:"role"`
	Projects []InviteProject `json:"projects"`
}
