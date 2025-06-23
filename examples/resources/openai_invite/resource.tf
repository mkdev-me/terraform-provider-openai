# Create a project first (or reference an existing one)
resource "openai_project" "main" {
  name        = "Main Project"
  description = "Primary project for team collaboration"
}

# Invite a new user to the organization with owner role
resource "openai_invite" "team_lead" {
  email = "john.doe@example.com"
  role  = "owner" # Organization-level role: "owner" or "reader"

  # Assign to at least one project (required)
  projects {
    id   = openai_project.main.id
    role = "owner" # Project-level role: "owner" or "member"
  }
}

# Invite a developer with reader role at org level
resource "openai_invite" "developer" {
  email = "jane.smith@example.com"
  role  = "reader" # Limited org-level permissions

  projects {
    id   = openai_project.main.id
    role = "member" # Full project access
  }
}

# Create multiple projects for fine-grained access
resource "openai_project" "api_development" {
  name        = "API Development"
  description = "API development and testing"
}

resource "openai_project" "ml_research" {
  name        = "ML Research"
  description = "Machine learning research projects"
}

# Invite data scientist with access to multiple projects
resource "openai_invite" "data_scientist" {
  email = "alice.johnson@example.com"
  role  = "reader"

  projects {
    id   = openai_project.ml_research.id
    role = "owner"
  }

  projects {
    id   = openai_project.api_development.id
    role = "member"
  }
}

# Invite contractor with limited access
resource "openai_invite" "contractor" {
  email = "bob.wilson@contractor.example.com"
  role  = "reader" # Minimal org permissions

  projects {
    id   = openai_project.api_development.id
    role = "member"
  }
}

# Output invite ID
output "team_lead_invite_id" {
  value = openai_invite.team_lead.invite_id
}