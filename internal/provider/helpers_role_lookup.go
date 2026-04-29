package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Per-process cache of project roles, populated lazily on first lookup.
//
// Used by state upgraders so that migrating N project_user resources that share
// the same project_id results in a single role-list API call rather than N
// (the v0→v1 migration of users.infrastructure has ~245 resources spread over
// ~14 projects). The cache lives for the duration of the provider process,
// which matches a single `terraform plan` invocation.
var (
	roleCacheMu sync.Mutex
	// projectID → lowercased role name → role ID
	roleCache = map[string]map[string]string{}
)

// resetRoleCacheForTest clears the package-level role cache. Used only by tests.
func resetRoleCacheForTest() {
	roleCacheMu.Lock()
	defer roleCacheMu.Unlock()
	roleCache = map[string]map[string]string{}
}

// projectClientHTTP returns the provider's configured *http.Client when
// available so that proxy/transport settings on the OpenAIClient are honoured.
// Falls back to a sensible default if the client wasn't initialised with one.
func projectClientHTTP(c *OpenAIClient) *http.Client {
	if c != nil && c.OpenAIClient != nil && c.OpenAIClient.HTTPClient != nil {
		return c.OpenAIClient.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// lookupProjectRoleIDByName resolves a project role name (e.g. "member", "owner")
// to its role ID by listing the project's roles via the admin API.
//
// Used by state upgraders to translate v0 schemas (which stored a role *name*)
// into v1 schemas (which store role *IDs*). The lookup is case-insensitive.
// Results are cached per project for the lifetime of the process to avoid
// repeated identical list calls when many resources share a project.
func lookupProjectRoleIDByName(ctx context.Context, c *OpenAIClient, projectID, roleName string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("openai client is not configured")
	}
	if adminAPIKey(c) == "" {
		return "", fmt.Errorf("admin API key is required to resolve project role %q in project %s", roleName, projectID)
	}

	nameKey := strings.ToLower(roleName)

	// The lock is held across the API call so that concurrent migrations of
	// resources in the same project resolve to a single list-roles request:
	// later goroutines block briefly, then hit the populated cache.
	roleCacheMu.Lock()
	defer roleCacheMu.Unlock()

	if cached, ok := roleCache[projectID]; ok {
		if id, ok := cached[nameKey]; ok {
			return id, nil
		}
		return "", fmt.Errorf("no role with name %q found in project %s", roleName, projectID)
	}

	rolesByName, err := listProjectRoles(ctx, c, projectID)
	if err != nil {
		return "", err
	}
	roleCache[projectID] = rolesByName

	if id, ok := rolesByName[nameKey]; ok {
		return id, nil
	}
	return "", fmt.Errorf("no role with name %q found in project %s", roleName, projectID)
}

// listProjectRoles fetches all roles defined in a project and returns them as
// a lowercased-name → role-ID map. Pagination is followed to completion.
func listProjectRoles(ctx context.Context, c *OpenAIClient, projectID string) (map[string]string, error) {
	rolesURL := adminBaseURL(c) + "/v1/projects/" + projectID + "/roles"
	httpClient := projectClientHTTP(c)
	cursor := ""
	out := map[string]string{}

	for {
		parsedURL, err := url.Parse(rolesURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing roles URL: %w", err)
		}
		q := parsedURL.Query()
		q.Set("limit", "100")
		if cursor != "" {
			q.Set("after", cursor)
		}
		parsedURL.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("error creating roles request: %w", err)
		}
		setAdminAuthHeaders(c, req)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error executing roles request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("API error listing project roles for %s: %s", projectID, resp.Status)
		}

		var listResp RoleListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("error parsing roles response: %w", err)
		}
		resp.Body.Close()

		for _, r := range listResp.Data {
			out[strings.ToLower(r.Name)] = r.ID
		}

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		// Defensive: an empty or non-progressing cursor would loop forever.
		next := *listResp.Next
		if next == "" || next == cursor {
			break
		}
		cursor = next
	}

	return out, nil
}
