package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// lookupProjectRoleIDByName resolves a project role name (e.g. "member", "owner")
// to its role ID by listing the project's roles via the admin API.
//
// Used by state upgraders to translate v0 schemas (which stored a role *name*)
// into v1 schemas (which store role *IDs*). The lookup is case-insensitive,
// matching the behavior of the openai_project_role data source.
func lookupProjectRoleIDByName(ctx context.Context, c *OpenAIClient, projectID, roleName string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("openai client is not configured")
	}
	if adminAPIKey(c) == "" {
		return "", fmt.Errorf("admin API key is required to resolve project role %q in project %s", roleName, projectID)
	}

	rolesURL := adminBaseURL(c) + "/v1/projects/" + projectID + "/roles"
	httpClient := &http.Client{Timeout: 30 * time.Second}
	cursor := ""

	for {
		parsedURL, err := url.Parse(rolesURL)
		if err != nil {
			return "", fmt.Errorf("error parsing roles URL: %w", err)
		}
		q := parsedURL.Query()
		q.Set("limit", "100")
		if cursor != "" {
			q.Set("after", cursor)
		}
		parsedURL.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
		if err != nil {
			return "", fmt.Errorf("error creating roles request: %w", err)
		}
		setAdminAuthHeaders(c, req)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("error executing roles request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return "", fmt.Errorf("API error listing project roles for %s: %s", projectID, resp.Status)
		}

		var listResp RoleListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			resp.Body.Close()
			return "", fmt.Errorf("error parsing roles response: %w", err)
		}
		resp.Body.Close()

		for _, r := range listResp.Data {
			if strings.EqualFold(r.Name, roleName) {
				return r.ID, nil
			}
		}

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		cursor = *listResp.Next
	}

	return "", fmt.Errorf("no role with name %q found in project %s", roleName, projectID)
}
