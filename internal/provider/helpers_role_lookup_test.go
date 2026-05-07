package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mkdev-me/terraform-provider-openai/internal/client"
)

// resetAdminSemaphoreForTest re-initialises the package-level semaphore so
// tests that mutate OPENAI_ADMIN_MAX_CONCURRENT or want a clean state can do
// so. Not safe for concurrent use; tests must serialise around it.
func resetAdminSemaphoreForTest(size int) {
	adminSemaphore = make(chan struct{}, size)
	adminSemaphoreOnce = sync.Once{}
	adminSemaphoreOnce.Do(func() {}) // mark consumed so initAdminSemaphore won't overwrite
}

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
	// Exactly retryMaxAttempts requests should be made — no extra reissue.
	if calls != retryMaxAttempts {
		t.Errorf("expected exactly %d calls, got %d", retryMaxAttempts, calls)
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
	// Retry-After header values are honoured exactly (no jitter applied).
	if got := backoffDuration(0, "5"); got.Seconds() != 5 {
		t.Errorf("retry-after=5 → %v, want 5s", got)
	}
	if got := backoffDuration(0, "120"); got.Seconds() != 60 {
		t.Errorf("retry-after=120 should cap at 60s, got %v", got)
	}
	// Fallback path uses jittered exponential backoff: actual sleep is in
	// [base/2, base] for base values 1, 2, 4, 8, 16, 30s.
	for _, tc := range []struct {
		attempt int
		base    time.Duration
	}{
		{0, 1 * time.Second},
		{2, 4 * time.Second},
		{5, 30 * time.Second},
	} {
		got := backoffDuration(tc.attempt, "")
		if got < tc.base/2 || got > tc.base {
			t.Errorf("attempt %d → %v, want in [%v, %v]", tc.attempt, got, tc.base/2, tc.base)
		}
	}
	got := backoffDuration(0, "not-a-number")
	if got < 500*time.Millisecond || got > 1*time.Second {
		t.Errorf("invalid retry-after should fall back to attempt 0 jitter range [500ms, 1s], got %v", got)
	}
}

// TestAdminSemaphore_LimitsConcurrency fires N concurrent requests against a
// test server with the semaphore set to 2 and asserts the server never sees
// more than 2 in-flight at once. Without the semaphore, all N would land
// simultaneously — exactly the burst pattern that 429s the real admin API.
func TestAdminSemaphore_LimitsConcurrency(t *testing.T) {
	resetAdminSemaphoreForTest(2)

	var (
		inFlight    int32
		maxObserved int32
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt32(&inFlight, 1)
		for {
			seen := atomic.LoadInt32(&maxObserved)
			if cur <= seen || atomic.CompareAndSwapInt32(&maxObserved, seen, cur) {
				break
			}
		}
		// Hold the slot long enough that any unbounded concurrency would
		// pile on visibly before the first request releases.
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&inFlight, -1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	httpClient := projectClientHTTP(c)

	const N = 10
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			resp, err := doRequestWithRetry(context.Background(), httpClient, c, "GET", server.URL+"/v1/anything", nil)
			if err != nil {
				t.Errorf("request failed: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&maxObserved); got > 2 {
		t.Errorf("max in-flight requests = %d, want <= 2", got)
	}
}

// TestAdminSemaphore_RateLimitedServer reproduces the production failure
// mode: a server that returns 429 when more than `serverLimit` requests are
// in-flight simultaneously. With the semaphore sized to <= serverLimit, all
// N concurrent callers must complete successfully without exhausting their
// retry budget. Without the semaphore, retries pile up and the test exhausts
// retryMaxAttempts (this was the v2.2.5 production behaviour).
func TestAdminSemaphore_RateLimitedServer(t *testing.T) {
	const (
		N           = 50 // concurrent callers — well above Terraform's default parallelism of 10
		serverLimit = 3  // server allows at most 3 in-flight; 4th and beyond get 429
	)
	resetAdminSemaphoreForTest(serverLimit)

	var (
		inFlight   int32
		totalCalls int32
		rejections int32
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&totalCalls, 1)
		cur := atomic.AddInt32(&inFlight, 1)
		defer atomic.AddInt32(&inFlight, -1)

		if cur > serverLimit {
			atomic.AddInt32(&rejections, 1)
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Simulate API work so requests overlap.
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	httpClient := projectClientHTTP(c)

	var (
		wg       sync.WaitGroup
		failures int32
	)
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			resp, err := doRequestWithRetry(context.Background(), httpClient, c, "GET", server.URL+"/v1/anything", nil)
			if err != nil {
				atomic.AddInt32(&failures, 1)
				return
			}
			if resp.StatusCode != http.StatusOK {
				atomic.AddInt32(&failures, 1)
			}
			resp.Body.Close()
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&failures); got != 0 {
		t.Errorf("%d/%d requests failed; semaphore should have kept us under server limit", got, N)
	}
	if got := atomic.LoadInt32(&rejections); got != 0 {
		t.Errorf("server saw %d 429s; semaphore should have prevented all of them", got)
	}
	t.Logf("N=%d concurrent callers, serverLimit=%d, semaphore=%d → totalCalls=%d, 429s=%d, failures=%d",
		N, serverLimit, serverLimit, atomic.LoadInt32(&totalCalls), atomic.LoadInt32(&rejections), atomic.LoadInt32(&failures))
}

// TestAdminSemaphore_RateLimitedServer_WithoutSemaphore is the negative
// control: same scenario but with the semaphore sized larger than the
// server's limit, so concurrency exceeds capacity and we see 429s. This
// asserts the test setup actually exercises the rate-limit path — without
// it, a passing TestAdminSemaphore_RateLimitedServer could be vacuous.
func TestAdminSemaphore_RateLimitedServer_WithoutSemaphore(t *testing.T) {
	const (
		N           = 50
		serverLimit = 3
	)
	resetAdminSemaphoreForTest(N) // effectively no limit vs. serverLimit

	var (
		inFlight   int32
		rejections int32
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt32(&inFlight, 1)
		defer atomic.AddInt32(&inFlight, -1)

		if cur > serverLimit {
			atomic.AddInt32(&rejections, 1)
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	httpClient := projectClientHTTP(c)

	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			resp, err := doRequestWithRetry(context.Background(), httpClient, c, "GET", server.URL+"/v1/anything", nil)
			if err == nil && resp != nil {
				resp.Body.Close()
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&rejections); got == 0 {
		t.Fatal("expected 429s when semaphore size > server limit; test setup may be wrong")
	}
	t.Logf("control: with semaphore=%d > serverLimit=%d, server returned %d 429s (proves the rate-limit path is exercised)",
		N, serverLimit, atomic.LoadInt32(&rejections))
}

// TestAdminSemaphore_ContextCancellation verifies that callers blocked
// waiting for a slot return promptly when the context is cancelled, rather
// than blocking the apply indefinitely.
func TestAdminSemaphore_ContextCancellation(t *testing.T) {
	resetAdminSemaphoreForTest(1)

	// Saturate the single slot.
	release, err := acquireAdminSlot(context.Background())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	defer release()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = acquireAdminSlot(ctx)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected error when context cancels before slot frees")
	}
	if elapsed > 200*time.Millisecond {
		t.Errorf("acquire took %v, want < 200ms", elapsed)
	}
}
