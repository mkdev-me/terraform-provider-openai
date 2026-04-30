package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mkdev-me/terraform-provider-openai/internal/client"
)

func newTestOpenAIClient(serverURL string) *OpenAIClient {
	return &OpenAIClient{
		OpenAIClient: client.NewClient("test-api-key", "", serverURL+"/v1"),
		AdminAPIKey:  "test-admin-key",
	}
}

func TestLookupProjectRoleIDByName_Found(t *testing.T) {
	resetRoleCacheForTest()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/proj_found/roles" {
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

	got, err := lookupProjectRoleIDByName(context.Background(), c, "proj_found", "member")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "role_member_id" {
		t.Fatalf("got role ID %q, want %q", got, "role_member_id")
	}
}

func TestLookupProjectRoleIDByName_CaseInsensitive(t *testing.T) {
	resetRoleCacheForTest()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/proj_caseins/roles" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":   "list",
			"data":     []map[string]interface{}{{"id": "role_owner_id", "name": "Owner"}},
			"has_more": false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)

	got, err := lookupProjectRoleIDByName(context.Background(), c, "proj_caseins", "owner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "role_owner_id" {
		t.Fatalf("got role ID %q, want %q", got, "role_owner_id")
	}
}

func TestLookupProjectRoleIDByName_NotFound(t *testing.T) {
	resetRoleCacheForTest()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/proj_notfound/roles" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":   "list",
			"data":     []map[string]interface{}{{"id": "role_other_id", "name": "viewer"}},
			"has_more": false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)

	_, err := lookupProjectRoleIDByName(context.Background(), c, "proj_notfound", "member")
	if err == nil {
		t.Fatal("expected error for missing role, got nil")
	}
	if !strings.Contains(err.Error(), "no role with name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLookupProjectRoleIDByName_Pagination(t *testing.T) {
	resetRoleCacheForTest()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/proj_paged/roles" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
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

	got, err := lookupProjectRoleIDByName(context.Background(), c, "proj_paged", "member")
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

// Defensive: a buggy API returning has_more=true with an empty cursor must not
// loop forever — the helper should break after one extra page.
func TestLookupProjectRoleIDByName_PaginationEmptyCursorGuard(t *testing.T) {
	resetRoleCacheForTest()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls > 5 {
			t.Fatalf("loop did not terminate after %d calls", calls)
		}
		emptyNext := ""
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":   "list",
			"data":     []map[string]interface{}{{"id": "role_member_id", "name": "member"}},
			"has_more": true,
			"next":     &emptyNext,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)

	got, err := lookupProjectRoleIDByName(context.Background(), c, "proj_emptycursor", "member")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "role_member_id" {
		t.Fatalf("got role ID %q, want %q", got, "role_member_id")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call (loop should break on empty cursor), got %d", calls)
	}
}

func TestLookupProjectRoleIDByName_MissingAdminKey(t *testing.T) {
	resetRoleCacheForTest()
	c := &OpenAIClient{
		OpenAIClient: client.NewClient("", "", "https://api.openai.com/v1"),
	}

	_, err := lookupProjectRoleIDByName(context.Background(), c, "proj_noauth", "member")
	if err == nil {
		t.Fatal("expected error for missing admin key, got nil")
	}
	if !strings.Contains(err.Error(), "admin API key is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Cache hit: looking up several roles in the same project should produce only
// a single API call (the role list is fetched once and reused).
func TestLookupProjectRoleIDByName_CacheReusesAcrossCalls(t *testing.T) {
	resetRoleCacheForTest()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object": "list",
			"data": []map[string]interface{}{
				{"id": "role_member_id", "name": "member"},
				{"id": "role_owner_id", "name": "owner"},
			},
			"has_more": false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)

	for i := 0; i < 5; i++ {
		if _, err := lookupProjectRoleIDByName(context.Background(), c, "proj_cached", "member"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := lookupProjectRoleIDByName(context.Background(), c, "proj_cached", "owner"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if calls != 1 {
		t.Fatalf("expected exactly 1 API call thanks to the cache, got %d", calls)
	}
}

// Cache miss for a role name that genuinely doesn't exist must fail without
// re-listing on every call.
func TestLookupProjectRoleIDByName_CacheRemembersMisses(t *testing.T) {
	resetRoleCacheForTest()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":   "list",
			"data":     []map[string]interface{}{{"id": "role_owner_id", "name": "owner"}},
			"has_more": false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)

	for i := 0; i < 3; i++ {
		_, err := lookupProjectRoleIDByName(context.Background(), c, "proj_missrole", "nonexistent")
		if err == nil {
			t.Fatal("expected error for missing role")
		}
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 API call, got %d", calls)
	}
}

// Override the test backoff to be near-instant so retry tests don't take ages.
// We swap the default backoff with a fixed 5ms wait via a test-only knob.
// (Using a tiny default in code would make production retries useless, so we
// keep production fast-on-no-retry-after but tests use Retry-After header.)

func TestDoWithRetry_RetriesOn429(t *testing.T) {
	resetRoleCacheForTest()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.Header().Set("Retry-After", "0") // honour: 0 => immediate retry
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":   "list",
			"data":     []map[string]interface{}{{"id": "role_member_id", "name": "member"}},
			"has_more": false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	got, err := lookupProjectRoleIDByName(context.Background(), c, "proj_retry", "member")
	if err != nil {
		t.Fatalf("expected eventual success, got error: %v", err)
	}
	if got != "role_member_id" {
		t.Errorf("got role ID %q, want %q", got, "role_member_id")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls (2 × 429 + 1 × 200), got %d", calls)
	}
}

func TestDoWithRetry_RetriesOn5xx(t *testing.T) {
	resetRoleCacheForTest()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			http.Error(w, "transient", http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":   "list",
			"data":     []map[string]interface{}{{"id": "role_x", "name": "member"}},
			"has_more": false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	got, err := lookupProjectRoleIDByName(context.Background(), c, "proj_5xx", "member")
	if err != nil {
		t.Fatalf("expected eventual success, got: %v", err)
	}
	if got != "role_x" {
		t.Errorf("got %q, want %q", got, "role_x")
	}
	if calls != 2 {
		t.Errorf("expected 2 calls (1 × 502 + 1 × 200), got %d", calls)
	}
}

func TestDoWithRetry_GivesUpAfterMaxAttempts(t *testing.T) {
	resetRoleCacheForTest()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "0")
		http.Error(w, "still rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	_, err := lookupProjectRoleIDByName(context.Background(), c, "proj_giveup", "member")
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}
	// retryMaxAttempts = 6, plus the final re-issue inside doWithRetry, then
	// listProjectRoles surfaces the non-OK response as an error. We expect
	// at least retryMaxAttempts attempts to have been made.
	if calls < retryMaxAttempts {
		t.Errorf("expected at least %d calls, got %d", retryMaxAttempts, calls)
	}
}

func TestDoWithRetry_DoesNotRetryOn4xxOtherThan429(t *testing.T) {
	resetRoleCacheForTest()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	_, err := lookupProjectRoleIDByName(context.Background(), c, "proj_403", "member")
	if err == nil {
		t.Fatal("expected error on 403, got nil")
	}
	if calls != 1 {
		t.Errorf("403 should not retry; expected 1 call, got %d", calls)
	}
}

func TestBackoffDuration_HonoursRetryAfter(t *testing.T) {
	if got := backoffDuration(0, "5"); got.Seconds() != 5 {
		t.Errorf("retry-after=5 → %v, want 5s", got)
	}
	if got := backoffDuration(0, "120"); got.Seconds() != 60 {
		t.Errorf("retry-after=120 should cap at 60s, got %v", got)
	}
	if got := backoffDuration(0, "not-a-number"); got != 1*time.Second {
		t.Errorf("invalid retry-after should fall back, got %v", got)
	}
	if got := backoffDuration(2, ""); got != 4*time.Second {
		t.Errorf("attempt 2 with no retry-after → %v, want 4s", got)
	}
}
