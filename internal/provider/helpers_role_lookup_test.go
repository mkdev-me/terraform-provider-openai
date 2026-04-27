package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mkdev-me/terraform-provider-openai/internal/client"
)

func newTestOpenAIClient(serverURL string) *OpenAIClient {
	return &OpenAIClient{
		OpenAIClient: client.NewClient("test-api-key", "", serverURL+"/v1"),
		AdminAPIKey:  "test-admin-key",
	}
}

func TestLookupProjectRoleIDByName_Found(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/projects/proj_123/roles") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-admin-key" {
			t.Fatalf("unexpected Authorization header: %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object": "list",
			"data": []map[string]interface{}{
				{"id": "role_owner_id", "name": "owner"},
				{"id": "role_member_id", "name": "member"},
			},
			"has_more": false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)

	got, err := lookupProjectRoleIDByName(context.Background(), c, "proj_123", "member")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "role_member_id" {
		t.Fatalf("got role ID %q, want %q", got, "role_member_id")
	}
}

func TestLookupProjectRoleIDByName_CaseInsensitive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":   "list",
			"data":     []map[string]interface{}{{"id": "role_owner_id", "name": "Owner"}},
			"has_more": false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)

	got, err := lookupProjectRoleIDByName(context.Background(), c, "proj_123", "owner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "role_owner_id" {
		t.Fatalf("got role ID %q, want %q", got, "role_owner_id")
	}
}

func TestLookupProjectRoleIDByName_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":   "list",
			"data":     []map[string]interface{}{{"id": "role_other_id", "name": "viewer"}},
			"has_more": false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)

	_, err := lookupProjectRoleIDByName(context.Background(), c, "proj_123", "member")
	if err == nil {
		t.Fatal("expected error for missing role, got nil")
	}
	if !strings.Contains(err.Error(), "no role with name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLookupProjectRoleIDByName_Pagination(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		next := "cursor-2"
		if r.URL.Query().Get("after") == "cursor-2" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"object":   "list",
				"data":     []map[string]interface{}{{"id": "role_member_id", "name": "member"}},
				"has_more": false,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":   "list",
			"data":     []map[string]interface{}{{"id": "role_owner_id", "name": "owner"}},
			"has_more": true,
			"next":     &next,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)

	got, err := lookupProjectRoleIDByName(context.Background(), c, "proj_123", "member")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "role_member_id" {
		t.Fatalf("got role ID %q, want %q", got, "role_member_id")
	}
	if calls != 2 {
		t.Fatalf("expected 2 paginated calls, got %d", calls)
	}
}

func TestLookupProjectRoleIDByName_MissingAdminKey(t *testing.T) {
	c := &OpenAIClient{
		OpenAIClient: client.NewClient("", "", "https://api.openai.com/v1"),
	}

	_, err := lookupProjectRoleIDByName(context.Background(), c, "proj_123", "member")
	if err == nil {
		t.Fatal("expected error for missing admin key, got nil")
	}
	if !strings.Contains(err.Error(), "admin API key is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
