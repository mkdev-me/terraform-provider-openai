# OpenAI Groups API - Research Document

**Issue:** tpo-8lm
**Date:** 2026-02-01
**Source:** [OpenAI API Reference - Project Groups](https://platform.openai.com/docs/api-reference/project-groups)

## Overview

The OpenAI Groups API (officially "Project Groups") manages which groups have access to a project and the role they receive. Groups are collections of users that can be synced from an identity provider via SCIM.

All endpoints require an **Admin API Key** (`Authorization: Bearer $OPENAI_ADMIN_KEY`).

---

## The Project Group Object

```json
{
  "object": "project.group",
  "project_id": "proj_abc123",
  "group_id": "group_xyz789",
  "group_name": "Support Team",
  "role": "member",
  "created_at": 1711471533
}
```

| Field | Type | Description |
|-------|------|-------------|
| `object` | string | Always `"project.group"` |
| `project_id` | string | Identifier of the project |
| `group_id` | string | Identifier of the group |
| `group_name` | string | Display name of the group |
| `role` | string | Project role identifier granted to the group |
| `created_at` | integer (unix timestamp) | When the group was added to the project |

---

## Endpoints

### 1. List Project Groups

```
GET /v1/organization/projects/{project_id}/groups
```

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | string | Yes | The ID of the project |

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 20 | Number of groups to return |
| `after` | string | No | — | Cursor for pagination (ID of last group from previous page) |

**Response:**

```json
{
  "object": "list",
  "data": [
    {
      "object": "project.group",
      "project_id": "proj_abc123",
      "group_id": "group_xyz789",
      "group_name": "Support Team",
      "role": "member",
      "created_at": 1711471533
    }
  ],
  "first_id": "group_xyz789",
  "last_id": "group_xyz789",
  "has_more": false
}
```

---

### 2. Create (Add Group to Project)

```
POST /v1/organization/projects/{project_id}/groups
```

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | string | Yes | The ID of the project |

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `group_id` | string | Yes | The ID of the group to add |
| `role` | string | Yes | Project role identifier to grant |

**Response:** Returns the created `project.group` object.

---

### 3. Delete (Revoke Group from Project)

```
DELETE /v1/organization/projects/{project_id}/groups/{group_id}
```

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | string | Yes | The ID of the project |
| `group_id` | string | Yes | The ID of the group to remove |

**Response:** Confirmation of deletion.

---

## Notable: No Retrieve or Update Endpoints

The API does **not** provide:
- `GET /v1/organization/projects/{project_id}/groups/{group_id}` (retrieve single group)
- `PATCH/POST /v1/organization/projects/{project_id}/groups/{group_id}` (update role)

To read a single group's project membership, you must list all groups and filter. To change a group's role, you must delete and re-create.

---

## Related: Role Assignments API

There are additional role assignment endpoints for fine-grained RBAC:

- `GET /v1/projects/{project_id}/groups/{group_id}/roles` — List project roles for a group
- `POST /v1/projects/{project_id}/groups/{group_id}/roles` — Assign a project role to a group
- `DELETE /v1/projects/{project_id}/groups/{group_id}/roles/{role_id}` — Unassign a project role

Note the path prefix difference: `/v1/projects/` vs `/v1/organization/projects/`.

---

## Terraform Provider Implementation Notes

### Closest Existing Pattern

The `openai_project_user` resource (`resource_openai_project_user.go`) is the closest analog:
- Same parent path structure (`/organization/projects/{project_id}/...`)
- Same CRUD pattern (POST to add, DELETE to remove)
- Same admin key authentication
- Same composite ID format (`project_id:user_id` → `project_id:group_id`)

### Proposed Response Type

```go
// ProjectGroupResponseFramework represents the API response for a project group.
type ProjectGroupResponseFramework struct {
    Object    string `json:"object"`
    ProjectID string `json:"project_id"`
    GroupID   string `json:"group_id"`
    GroupName string `json:"group_name"`
    Role      string `json:"role"`
    CreatedAt int64  `json:"created_at"`
}
```

### Resource Schema Fields

| Terraform Field | Schema Type | Required/Computed | Plan Modifier | Notes |
|----------------|-------------|-------------------|---------------|-------|
| `id` | String | Computed | UseStateForUnknown | `project_id:group_id` |
| `project_id` | String | Required | RequiresReplace | Immutable |
| `group_id` | String | Required | RequiresReplace | Immutable |
| `group_name` | String | Computed | — | From API response |
| `role` | String | Required | — | Updatable (via delete+recreate since no update API) |
| `created_at` | Int64 | Computed | — | Unix timestamp |

### Key Design Decision: Role Updates

Since there is no update endpoint, changing a group's role requires destroy+recreate. Two options:
1. **RequiresReplace on `role`** — Terraform handles it automatically via ForceNew
2. **Implement Update as delete+create** — Seamless to user but riskier (gap in access)

**Recommendation:** Use `RequiresReplace` on `role` to be explicit and safe.

### Read Implementation

No single-group GET endpoint exists. Read must:
1. `GET /organization/projects/{project_id}/groups?limit=100`
2. Paginate through results
3. Filter for matching `group_id`
4. Return not-found (remove from state) if group not in list

This matches the pattern used by `data_source_openai_project_user.go` which also paginates to find users.

### Data Source

A `openai_project_groups` (list) data source would be straightforward — direct mapping of the list endpoint with pagination support.

A singular `openai_project_group` data source would need the same paginate-and-filter approach as the resource's Read.

---

## Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/provider/types_project_org.go` | Modify | Add `ProjectGroupResponseFramework` |
| `internal/provider/resource_openai_project_group.go` | Create | Resource implementation |
| `internal/provider/data_source_openai_project_group.go` | Create | Data source (singular + list) |
| `internal/provider/provider.go` | Modify | Register new resource and data sources |
| `internal/provider/resource_openai_project_group_test.go` | Create | Acceptance tests |
