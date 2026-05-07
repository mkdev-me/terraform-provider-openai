package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

	// fullRoleCache stores complete role objects per project. Used by the
	// `data "openai_project_role"` data source (singular and plural) so that
	// multiple lookups against the same project — e.g. one for "member" and
	// one for "owner" — share a single admin-API list call. Without this,
	// a `terraform plan` declaring N role lookups across M projects fires
	// N concurrent paginated GETs that burst the admin rate limit even with
	// retries.
	fullRoleCacheMu sync.Mutex
	fullRoleCache   = map[string][]RoleResponseFramework{}

	// groupCache stores all SCIM-managed groups in the org. Used by the
	// `data "openai_group"` data source (singular and plural). The endpoint
	// returns the same list regardless of which group name we're filtering
	// for, so N concurrent group lookups all paginate the same data — cache
	// once and serve all subsequent lookups from memory.
	groupCacheMu sync.Mutex
	groupCache   []GroupResponseFramework
)

// resetRoleCacheForTest clears the package-level role cache. Used only by tests.
func resetRoleCacheForTest() {
	roleCacheMu.Lock()
	defer roleCacheMu.Unlock()
	roleCache = map[string]map[string]string{}

	fullRoleCacheMu.Lock()
	defer fullRoleCacheMu.Unlock()
	fullRoleCache = map[string][]RoleResponseFramework{}

	groupCacheMu.Lock()
	defer groupCacheMu.Unlock()
	groupCache = nil
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

// cachedListProjectRolesFull returns all roles for the given project, fetching
// them via the admin API on first call and serving subsequent calls from a
// per-process cache. The lock is held across the API call so concurrent
// callers for the same project resolve to a single list-roles request.
func cachedListProjectRolesFull(ctx context.Context, c *OpenAIClient, projectID string) ([]RoleResponseFramework, error) {
	if c == nil {
		return nil, fmt.Errorf("openai client is not configured")
	}
	if adminAPIKey(c) == "" {
		return nil, fmt.Errorf("admin API key is required to list roles for project %s", projectID)
	}

	fullRoleCacheMu.Lock()
	defer fullRoleCacheMu.Unlock()

	if cached, ok := fullRoleCache[projectID]; ok {
		return cached, nil
	}

	httpClient := projectClientHTTP(c)
	rolesURL := adminBaseURL(c) + "/v1/projects/" + projectID + "/roles"
	cursor := ""
	out := []RoleResponseFramework{}

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

		resp, err := doWithRetry(ctx, httpClient, c, parsedURL.String())
		if err != nil {
			return nil, err
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

		out = append(out, listResp.Data...)

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		next := *listResp.Next
		if next == "" || next == cursor {
			break
		}
		cursor = next
	}

	fullRoleCache[projectID] = out
	return out, nil
}

// cachedListAllGroups returns all SCIM-managed groups in the org, fetching on
// first call and serving subsequent calls from a per-process cache. The
// `data "openai_group"` data source filters this list by name client-side; N
// concurrent group lookups now resolve to a single paginated list call.
func cachedListAllGroups(ctx context.Context, c *OpenAIClient) ([]GroupResponseFramework, error) {
	if c == nil {
		return nil, fmt.Errorf("openai client is not configured")
	}
	if adminAPIKey(c) == "" {
		return nil, fmt.Errorf("admin API key is required to list organization groups")
	}

	groupCacheMu.Lock()
	defer groupCacheMu.Unlock()

	if groupCache != nil {
		return groupCache, nil
	}

	httpClient := projectClientHTTP(c)
	groupsURL := adminBaseURL(c) + "/v1/organization/groups"
	cursor := ""
	out := []GroupResponseFramework{}

	for {
		parsedURL, err := url.Parse(groupsURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing groups URL: %w", err)
		}
		q := parsedURL.Query()
		q.Set("limit", "100")
		if cursor != "" {
			q.Set("after", cursor)
		}
		parsedURL.RawQuery = q.Encode()

		resp, err := doWithRetry(ctx, httpClient, c, parsedURL.String())
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("API error listing organization groups: %s", resp.Status)
		}

		var listResp GroupListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("error parsing groups response: %w", err)
		}
		resp.Body.Close()

		out = append(out, listResp.Data...)

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		next := *listResp.Next
		if next == "" || next == cursor {
			break
		}
		cursor = next
	}

	groupCache = out
	return out, nil
}

// listProjectRoles fetches all roles defined in a project and returns them as
// a lowercased-name → role-ID map. Pagination is followed to completion.
//
// Retries on 429 (Too Many Requests) and 5xx with exponential backoff. The
// admin API enforces a low rate limit (~60 RPM) and a state upgrade migrating
// many resources can burst past it; without retry the upgrader fails the
// entire plan.
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

		resp, err := doWithRetry(ctx, httpClient, c, parsedURL.String())
		if err != nil {
			return nil, err
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

// adminConcurrencyDefault caps the number of in-flight admin-API requests
// per provider process. Acts as a second line of defence on top of the
// rate limiter (see adminRateDefault below). Override with the
// OPENAI_ADMIN_MAX_CONCURRENT environment variable (clamped to [1, 64]).
const adminConcurrencyDefault = 3

// adminRateDefault and adminBurstDefault control the token-bucket rate
// limiter that paces admin-API calls. Empirically, the OpenAI admin API
// rate-limits per-endpoint at ~7-10 requests per minute (verified
// 2026-05-07: 7 sequential `GET /v1/projects/{id}/roles` calls succeeded,
// the next 7 in the same minute returned 429; bucket refilled after ~60s).
// No `x-ratelimit-*` or `Retry-After` headers are exposed, so the client
// cannot adapt dynamically.
//
// A pure concurrency cap (semaphore) does NOT fix this: even 1-in-flight
// sequential calls 429 once the per-minute bucket is exhausted. We need
// to pace at *rate*, not just concurrency.
//
// Default of 6 RPM with a burst of 4 leaves a comfortable margin under the
// observed ~7-10 RPM ceiling and lets short bursts (e.g. plan startup
// reading 4 cached roles) complete instantly while sustained load gets
// throttled. Override with OPENAI_ADMIN_MAX_RPM (clamped to [1, 600]) and
// OPENAI_ADMIN_BURST (clamped to [1, 100]).
const (
	adminRateDefault  = 6.0 // requests per minute (sustained)
	adminBurstDefault = 4   // initial burst capacity
)

// adminSemaphore gates entry into doRequestWithRetry. Buffered channel used
// as a counting semaphore: send to acquire, receive to release. Initialised
// once via sync.Once on first use so tests that mutate the env var before
// the first call see the right value.
var (
	adminSemaphoreOnce sync.Once
	adminSemaphore     chan struct{}
)

func initAdminSemaphore() {
	n := adminConcurrencyDefault
	if v := os.Getenv("OPENAI_ADMIN_MAX_CONCURRENT"); v != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && parsed >= 1 {
			if parsed > 64 {
				parsed = 64
			}
			n = parsed
		}
	}
	adminSemaphore = make(chan struct{}, n)
}

// acquireAdminSlot blocks until a slot is available or ctx is cancelled.
// Returns a release function the caller MUST invoke (via defer) once the
// admin call (including any retries) completes.
func acquireAdminSlot(ctx context.Context) (release func(), err error) {
	adminSemaphoreOnce.Do(initAdminSemaphore)
	select {
	case adminSemaphore <- struct{}{}:
		return func() { <-adminSemaphore }, nil
	case <-ctx.Done():
		return func() {}, ctx.Err()
	}
}

// adminTokenBucket is a simple token-bucket rate limiter. tokens accrue at
// `refillPerSec` until `capacity`; each admin request consumes one token.
// When the bucket is empty, callers wait for the next token.
//
// Hand-rolled rather than pulling in `golang.org/x/time/rate` because that
// dep would force a Go toolchain bump (>= 1.25 in current versions) on a
// project that targets 1.24, and the bucket logic we need is ~30 lines.
type adminTokenBucket struct {
	mu           sync.Mutex
	tokens       float64
	capacity     float64
	refillPerSec float64
	lastRefill   time.Time
}

var (
	adminBucketOnce sync.Once
	adminBucket     *adminTokenBucket
)

func initAdminBucket() {
	rpm := adminRateDefault
	if v := os.Getenv("OPENAI_ADMIN_MAX_RPM"); v != "" {
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && parsed >= 1 {
			if parsed > 600 {
				parsed = 600
			}
			rpm = parsed
		}
	}
	burst := adminBurstDefault
	if v := os.Getenv("OPENAI_ADMIN_BURST"); v != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && parsed >= 1 {
			if parsed > 100 {
				parsed = 100
			}
			burst = parsed
		}
	}
	adminBucket = &adminTokenBucket{
		tokens:       float64(burst),
		capacity:     float64(burst),
		refillPerSec: rpm / 60.0,
		lastRefill:   time.Now(),
	}
}

// waitForAdminToken consumes one token from the rate-limit bucket, sleeping
// (interruptible by ctx) until one becomes available. Loops on wake-up
// because multiple goroutines waiting on the same refill window will all
// get their timer signal together — only one wins the next token, the
// others recompute and sleep again.
func waitForAdminToken(ctx context.Context) error {
	adminBucketOnce.Do(initAdminBucket)
	for {
		adminBucket.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(adminBucket.lastRefill).Seconds()
		if elapsed > 0 {
			adminBucket.tokens += elapsed * adminBucket.refillPerSec
			if adminBucket.tokens > adminBucket.capacity {
				adminBucket.tokens = adminBucket.capacity
			}
			adminBucket.lastRefill = now
		}
		if adminBucket.tokens >= 1.0 {
			adminBucket.tokens -= 1.0
			adminBucket.mu.Unlock()
			return nil
		}
		needed := 1.0 - adminBucket.tokens
		wait := time.Duration(needed / adminBucket.refillPerSec * float64(time.Second))
		adminBucket.mu.Unlock()

		t := time.NewTimer(wait)
		select {
		case <-t.C:
			// loop and re-check; another goroutine may have taken the token.
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		}
	}
}

// retryStatusCodes are HTTP status codes that should trigger a retry with
// exponential backoff (rate limiting and transient server errors).
var retryStatusCodes = map[int]bool{
	http.StatusTooManyRequests:     true, // 429
	http.StatusInternalServerError: true, // 500
	http.StatusBadGateway:          true, // 502
	http.StatusServiceUnavailable:  true, // 503
	http.StatusGatewayTimeout:      true, // 504
}

// retryMaxAttempts is the maximum number of attempts (including the first)
// the retry helper makes before giving up.
const retryMaxAttempts = 6

// doWithRetry performs a GET against urlStr with exponential backoff on 429
// (rate limiting) and transient 5xx responses.
//
// Thin wrapper for backward compatibility. New code should use
// doRequestWithRetry directly.
func doWithRetry(ctx context.Context, httpClient *http.Client, c *OpenAIClient, urlStr string) (*http.Response, error) {
	return doRequestWithRetry(ctx, httpClient, c, "GET", urlStr, nil)
}

// doRequestWithRetry performs an HTTP request (any method, optional body) with
// exponential backoff and jitter on 429 (rate limiting) and transient 5xx
// responses.
//
// The OpenAI admin API enforces a low rate limit (~60 RPM org-wide); a
// terraform plan or apply touching many project users/groups can burst past
// it. Without retry the operation fails the entire run.
//
// `body` is the optional JSON request body — pass nil for GET/DELETE. The
// body bytes are buffered and a fresh reader is created per attempt, so
// retries replay the request faithfully.
//
// Honours the `Retry-After` header when present (seconds), otherwise falls
// back to a capped exponential schedule with full jitter (base values
// 1s, 2s, 4s, 8s, 16s, 30s — actual sleep is uniformly random in [base/2,
// base] to avoid thundering-herd when many concurrent retries fire).
func doRequestWithRetry(ctx context.Context, httpClient *http.Client, c *OpenAIClient, method, urlStr string, body []byte) (*http.Response, error) {
	release, err := acquireAdminSlot(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin slot acquisition cancelled: %w", err)
	}
	defer release()

	var lastErr error

	for attempt := 0; attempt < retryMaxAttempts; attempt++ {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}
		setAdminAuthHeaders(c, req)
		req.Header.Set("Content-Type", "application/json")

		// Rate-limit at *rate*, not just concurrency. The admin API throttles
		// per endpoint at ~7-10 RPM (no exposed headers) — even one-in-flight
		// sequential calls 429 once the bucket is empty. Consuming a token
		// per attempt (including retries) keeps every request the provider
		// makes paced under the ceiling.
		if err := waitForAdminToken(ctx); err != nil {
			return nil, fmt.Errorf("admin rate-limit wait cancelled: %w", err)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt == retryMaxAttempts-1 || !sleepWithBackoff(ctx, attempt, "") {
				return nil, fmt.Errorf("transport error after %d attempts: %w", attempt+1, err)
			}
			continue
		}

		if !retryStatusCodes[resp.StatusCode] {
			return resp, nil
		}

		// Final attempt: hand the (still-retryable) response back to the caller
		// untouched so the body remains readable for diagnostics.
		if attempt == retryMaxAttempts-1 {
			return resp, nil
		}

		// Going to retry: drain and close so the connection can be reused.
		retryAfter := resp.Header.Get("Retry-After")
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if !sleepWithBackoff(ctx, attempt, retryAfter) {
			return nil, fmt.Errorf("retry aborted: %w", ctx.Err())
		}
	}

	return nil, lastErr
}

// sleepWithBackoff waits before the next retry attempt. Returns false if the
// context was cancelled while waiting.
func sleepWithBackoff(ctx context.Context, attempt int, retryAfter string) bool {
	d := backoffDuration(attempt, retryAfter)
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// backoffDuration returns the wait time for a given attempt index. If
// retryAfter (the value of the `Retry-After` header) is a non-negative integer
// number of seconds, that value wins (capped at 60s; 0 means "retry now").
// Otherwise we use a capped exponential schedule with "decorrelated" jitter:
// the actual sleep is uniformly random in [base/2, base] for base values 1,
// 2, 4, 8, 16, 30s. The jitter avoids thundering-herd when N concurrent
// requests all 429 at once and would otherwise retry on the same schedule.
func backoffDuration(attempt int, retryAfter string) time.Duration {
	if retryAfter != "" {
		if secs, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && secs >= 0 {
			capped := secs
			if capped > 60 {
				capped = 60
			}
			return time.Duration(capped) * time.Second
		}
	}
	var base time.Duration
	switch attempt {
	case 0:
		base = 1 * time.Second
	case 1:
		base = 2 * time.Second
	case 2:
		base = 4 * time.Second
	case 3:
		base = 8 * time.Second
	case 4:
		base = 16 * time.Second
	default:
		base = 30 * time.Second
	}
	half := base / 2
	return half + time.Duration(rand.Int63n(int64(half)+1))
}
