package provider

import (
	"net/http"
	"strings"
)

// adminBaseURL returns the API base URL without trailing /v1 or /
func adminBaseURL(c *OpenAIClient) string {
	u := strings.TrimSuffix(c.OpenAIClient.APIURL, "/v1")
	return strings.TrimSuffix(u, "/")
}

// adminAPIKey returns the admin API key if set, otherwise the regular API key
func adminAPIKey(c *OpenAIClient) string {
	if c.AdminAPIKey != "" {
		return c.AdminAPIKey
	}
	return c.OpenAIClient.APIKey
}

// setAdminAuthHeaders sets Authorization and optional OpenAI-Organization headers
func setAdminAuthHeaders(c *OpenAIClient, req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+adminAPIKey(c))
	if c.OpenAIClient.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", c.OpenAIClient.OrganizationID)
	}
}
