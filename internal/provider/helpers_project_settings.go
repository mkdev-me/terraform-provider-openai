package provider

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const projectSettingsCacheTTL = 30 * time.Second

type ProjectModelPermissionsAPI struct {
	Mode     string   `json:"mode"`
	ModelIDs []string `json:"model_ids"`
	Object   string   `json:"object"`
}

type projectModelPermissionsRequest struct {
	Mode     string   `json:"mode"`
	ModelIDs []string `json:"model_ids"`
}

type SpendAlertNotificationAPI struct {
	Type          string   `json:"type"`
	Recipients    []string `json:"recipients"`
	SubjectPrefix *string  `json:"subject_prefix,omitempty"`
}

type SpendAlertAPI struct {
	ID                  string                    `json:"id"`
	Object              string                    `json:"object"`
	ThresholdAmount     int64                     `json:"threshold_amount"`
	Currency            string                    `json:"currency"`
	Interval            string                    `json:"interval"`
	NotificationChannel SpendAlertNotificationAPI `json:"notification_channel"`
}

type SpendAlertsListAPI struct {
	Object  string          `json:"object"`
	Data    []SpendAlertAPI `json:"data"`
	FirstID string          `json:"first_id"`
	LastID  string          `json:"last_id"`
	HasMore bool            `json:"has_more"`
}

type spendAlertRequest struct {
	ThresholdAmount     int64                     `json:"threshold_amount"`
	Currency            string                    `json:"currency"`
	Interval            string                    `json:"interval"`
	NotificationChannel SpendAlertNotificationAPI `json:"notification_channel"`
}

type SpendLimitEnforcementAPI struct {
	Status string `json:"status"`
}

type SpendLimitAPI struct {
	Object          string                   `json:"object"`
	ThresholdAmount int64                    `json:"threshold_amount"`
	Currency        string                   `json:"currency"`
	Interval        string                   `json:"interval"`
	Enforcement     SpendLimitEnforcementAPI `json:"enforcement"`
}

type spendLimitRequest struct {
	ThresholdAmount int64  `json:"threshold_amount"`
	Currency        string `json:"currency"`
	Interval        string `json:"interval"`
}

type projectSettingsCacheEntry[T any] struct {
	Value     T         `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

type settingsCacheLoad[T any] struct {
	done   chan struct{}
	entry  projectSettingsCacheEntry[T]
	status int
	err    error
	stale  bool
}

type settingsCache[T any] struct {
	sync.Mutex
	entries    map[string]projectSettingsCacheEntry[T]
	loads      map[string]*settingsCacheLoad[T]
	generation map[string]uint64
}

func newSettingsCache[T any]() settingsCache[T] {
	return settingsCache[T]{
		entries:    map[string]projectSettingsCacheEntry[T]{},
		loads:      map[string]*settingsCacheLoad[T]{},
		generation: map[string]uint64{},
	}
}

type spendAlertsCacheEntry struct {
	ExpiresAt time.Time
	Alerts    []SpendAlertAPI
	ByID      map[string]SpendAlertAPI
}

type spendAlertsCacheLoad struct {
	done   chan struct{}
	entry  spendAlertsCacheEntry
	status int
	err    error
	stale  bool
}

type spendAlertsFileCacheItem struct {
	ExpiresAt time.Time       `json:"expires_at"`
	Alerts    []SpendAlertAPI `json:"alerts"`
}

var projectModelPermissionsCache = newSettingsCache[ProjectModelPermissionsAPI]()
var spendLimitsCache = newSettingsCache[SpendLimitAPI]()

var spendAlertsCache = struct {
	sync.Mutex
	entries    map[string]spendAlertsCacheEntry
	loads      map[string]*spendAlertsCacheLoad
	generation map[string]uint64
}{
	entries:    map[string]spendAlertsCacheEntry{},
	loads:      map[string]*spendAlertsCacheLoad{},
	generation: map[string]uint64{},
}

func projectSettingsIdentityKey(c *OpenAIClient, values ...string) string {
	parts := []string{adminBaseURL(c), c.OrganizationID, adminAPIKey(c)}
	parts = append(parts, values...)
	hash := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return fmt.Sprintf("%x", hash)
}

func invalidateSetting[T any](cache *settingsCache[T], key, fileKind string) {
	cache.Lock()
	delete(cache.entries, key)
	cache.generation[key]++
	cache.Unlock()
	_ = os.Remove(projectSettingsCachePath(fileKind, key))
}

func invalidateProjectModelPermissionsCache(c *OpenAIClient, projectID string) {
	key := projectSettingsIdentityKey(c, projectID, "model_permissions")
	invalidateSetting(&projectModelPermissionsCache, key, "model-permissions")
}

func spendLimitCacheKey(c *OpenAIClient, scopeKind, scopeID string) string {
	return projectSettingsIdentityKey(c, scopeKind, scopeID, "spend_limit")
}

func invalidateProjectSpendLimitCache(c *OpenAIClient, projectID string) {
	key := spendLimitCacheKey(c, "project", projectID)
	invalidateSetting(&spendLimitsCache, key, "spend-limit")
}

func invalidateOrganizationSpendLimitCache(c *OpenAIClient) {
	key := spendLimitCacheKey(c, "organization", "")
	invalidateSetting(&spendLimitsCache, key, "spend-limit")
}

func spendAlertsCacheKey(c *OpenAIClient, scopeKind, scopeID string) string {
	return projectSettingsIdentityKey(c, scopeKind, scopeID, "spend_alerts")
}

func invalidateSpendAlertsCache(c *OpenAIClient, scopeKind, scopeID string) {
	key := spendAlertsCacheKey(c, scopeKind, scopeID)
	spendAlertsCache.Lock()
	delete(spendAlertsCache.entries, key)
	spendAlertsCache.generation[key]++
	spendAlertsCache.Unlock()
	_ = os.Remove(projectSettingsCachePath("spend-alerts-list", key))
}

func invalidateProjectSpendAlertsCache(c *OpenAIClient, projectID string) {
	invalidateSpendAlertsCache(c, "project", projectID)
}

func invalidateOrganizationSpendAlertsCache(c *OpenAIClient) {
	invalidateSpendAlertsCache(c, "organization", "")
}

func resetSettingsCacheForTest[T any](cache *settingsCache[T]) {
	cache.Lock()
	cache.entries = map[string]projectSettingsCacheEntry[T]{}
	cache.loads = map[string]*settingsCacheLoad[T]{}
	cache.generation = map[string]uint64{}
	cache.Unlock()
}

func resetProjectSettingsCacheForTest() {
	resetSettingsCacheForTest(&projectModelPermissionsCache)
	resetSettingsCacheForTest(&spendLimitsCache)

	spendAlertsCache.Lock()
	spendAlertsCache.entries = map[string]spendAlertsCacheEntry{}
	spendAlertsCache.loads = map[string]*spendAlertsCacheLoad{}
	spendAlertsCache.generation = map[string]uint64{}
	spendAlertsCache.Unlock()
}

func projectSettingsRequest(ctx context.Context, c *OpenAIClient, method, path string, requestBody any, responseBody any) (int, error) {
	if c == nil || c.OpenAIClient == nil || adminAPIKey(c) == "" {
		return 0, fmt.Errorf("admin API key is required")
	}

	var body []byte
	var err error
	if requestBody != nil {
		body, err = json.Marshal(requestBody)
		if err != nil {
			return 0, fmt.Errorf("error marshaling request: %w", err)
		}
	}

	resp, err := doRequestWithRetry(ctx, projectClientHTTP(c), c, method, adminBaseURL(c)+path, body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("error reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("OpenAI API returned %s: %s", resp.Status, string(responseBytes))
	}
	if responseBody != nil && len(responseBytes) > 0 {
		if err := json.Unmarshal(responseBytes, responseBody); err != nil {
			return resp.StatusCode, fmt.Errorf("error parsing response: %w", err)
		}
	}
	return resp.StatusCode, nil
}

func cachedSetting[T any](
	ctx context.Context,
	cache *settingsCache[T],
	key string,
	fileKind string,
	fetch func() (T, int, error),
	clone func(T) T,
) (*T, int, error) {
	for {
		cache.Lock()
		if cached, ok := cache.entries[key]; ok {
			if time.Now().Before(cached.ExpiresAt) {
				cache.Unlock()
				value := clone(cached.Value)
				return &value, http.StatusOK, nil
			}
			delete(cache.entries, key)
		}

		if load, ok := cache.loads[key]; ok {
			cache.Unlock()
			select {
			case <-load.done:
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			}
			if load.err != nil {
				return nil, load.status, load.err
			}
			if load.stale {
				continue
			}
			value := clone(load.entry.Value)
			return &value, load.status, nil
		}

		generation := cache.generation[key]
		load := &settingsCacheLoad[T]{done: make(chan struct{})}
		cache.loads[key] = load
		cache.Unlock()

		entry, ok := readProjectSettingsFileCache[T](fileKind, key)
		status := http.StatusOK
		var err error
		if !ok {
			var value T
			value, status, err = fetch()
			if err == nil {
				entry = projectSettingsCacheEntry[T]{
					Value:     value,
					ExpiresAt: time.Now().Add(projectSettingsCacheTTL),
				}
			}
		}

		cache.Lock()
		stale := cache.generation[key] != generation
		if err == nil && !stale {
			cache.entries[key] = entry
			_ = writeProjectSettingsFileCache(fileKind, key, entry)
		}
		load.entry = entry
		load.status = status
		load.err = err
		load.stale = stale
		delete(cache.loads, key)
		close(load.done)
		cache.Unlock()

		if err != nil {
			return nil, status, err
		}
		if stale {
			continue
		}
		value := clone(entry.Value)
		return &value, status, nil
	}
}

func cloneProjectModelPermissions(value ProjectModelPermissionsAPI) ProjectModelPermissionsAPI {
	value.ModelIDs = append([]string(nil), value.ModelIDs...)
	return value
}

func cachedProjectModelPermissions(ctx context.Context, c *OpenAIClient, projectID string) (*ProjectModelPermissionsAPI, int, error) {
	if c == nil || c.OpenAIClient == nil {
		return nil, 0, fmt.Errorf("admin API key is required")
	}
	key := projectSettingsIdentityKey(c, projectID, "model_permissions")
	return cachedSetting(
		ctx,
		&projectModelPermissionsCache,
		key,
		"model-permissions",
		func() (ProjectModelPermissionsAPI, int, error) {
			var value ProjectModelPermissionsAPI
			status, err := projectSettingsRequest(ctx, c, http.MethodGet, "/v1/organization/projects/"+projectID+"/model_permissions", nil, &value)
			return value, status, err
		},
		cloneProjectModelPermissions,
	)
}

func cloneSpendLimit(value SpendLimitAPI) SpendLimitAPI {
	return value
}

func cachedSpendLimit(ctx context.Context, c *OpenAIClient, scopeKind, scopeID, requestPath string) (*SpendLimitAPI, int, error) {
	if c == nil || c.OpenAIClient == nil {
		return nil, 0, fmt.Errorf("admin API key is required")
	}
	key := spendLimitCacheKey(c, scopeKind, scopeID)
	return cachedSetting(
		ctx,
		&spendLimitsCache,
		key,
		"spend-limit",
		func() (SpendLimitAPI, int, error) {
			var value SpendLimitAPI
			status, err := projectSettingsRequest(ctx, c, http.MethodGet, requestPath, nil, &value)
			return value, status, err
		},
		cloneSpendLimit,
	)
}

func cachedProjectSpendLimit(ctx context.Context, c *OpenAIClient, projectID string) (*SpendLimitAPI, int, error) {
	return cachedSpendLimit(ctx, c, "project", projectID, "/v1/organization/projects/"+projectID+"/spend_limit")
}

func cachedOrganizationSpendLimit(ctx context.Context, c *OpenAIClient) (*SpendLimitAPI, int, error) {
	return cachedSpendLimit(ctx, c, "organization", "", "/v1/organization/spend_limit")
}

func cachedSpendAlert(ctx context.Context, c *OpenAIClient, scopeKind, scopeID, basePath, alertID string) (*SpendAlertAPI, int, error) {
	entry, status, err := cachedSpendAlerts(ctx, c, scopeKind, scopeID, basePath)
	if status == http.StatusNotFound {
		return nil, status, nil
	}
	if err != nil {
		return nil, status, err
	}

	value, ok := entry.ByID[alertID]
	if !ok {
		return nil, http.StatusNotFound, nil
	}
	value = cloneSpendAlert(value)
	return &value, status, nil
}

func cachedProjectSpendAlert(ctx context.Context, c *OpenAIClient, projectID, alertID string) (*SpendAlertAPI, int, error) {
	return cachedSpendAlert(
		ctx,
		c,
		"project",
		projectID,
		"/v1/organization/projects/"+projectID+"/spend_alerts",
		alertID,
	)
}

func cachedOrganizationSpendAlert(ctx context.Context, c *OpenAIClient, alertID string) (*SpendAlertAPI, int, error) {
	return cachedSpendAlert(ctx, c, "organization", "", "/v1/organization/spend_alerts", alertID)
}

func cachedSpendAlerts(ctx context.Context, c *OpenAIClient, scopeKind, scopeID, basePath string) (spendAlertsCacheEntry, int, error) {
	return cachedSpendAlertsWithFileReader(ctx, c, scopeKind, scopeID, basePath, readSpendAlertsFileCache)
}

func cachedSpendAlertsWithFileReader(
	ctx context.Context,
	c *OpenAIClient,
	scopeKind string,
	scopeID string,
	basePath string,
	readFileCache func(string) (spendAlertsCacheEntry, bool),
) (spendAlertsCacheEntry, int, error) {
	if c == nil || c.OpenAIClient == nil {
		return spendAlertsCacheEntry{}, 0, fmt.Errorf("admin API key is required")
	}

	key := spendAlertsCacheKey(c, scopeKind, scopeID)

	for {
		spendAlertsCache.Lock()
		if cached, ok := spendAlertsCache.entries[key]; ok {
			if time.Now().Before(cached.ExpiresAt) {
				spendAlertsCache.Unlock()
				return cloneSpendAlertsCacheEntry(cached), http.StatusOK, nil
			}
			delete(spendAlertsCache.entries, key)
		}

		if load, ok := spendAlertsCache.loads[key]; ok {
			spendAlertsCache.Unlock()
			select {
			case <-load.done:
			case <-ctx.Done():
				return spendAlertsCacheEntry{}, 0, ctx.Err()
			}
			if load.err != nil {
				return spendAlertsCacheEntry{}, load.status, load.err
			}
			if load.stale {
				continue
			}
			return cloneSpendAlertsCacheEntry(load.entry), load.status, nil
		}

		generation := spendAlertsCache.generation[key]
		load := &spendAlertsCacheLoad{done: make(chan struct{})}
		spendAlertsCache.loads[key] = load
		spendAlertsCache.Unlock()

		entry, ok := readFileCache(key)
		status := http.StatusOK
		var err error
		if !ok {
			entry, status, err = fetchSpendAlerts(ctx, c, basePath)
		}

		spendAlertsCache.Lock()
		stale := spendAlertsCache.generation[key] != generation
		if err == nil && !stale {
			spendAlertsCache.entries[key] = entry
			_ = writeSpendAlertsFileCache(key, entry)
		}
		load.entry = entry
		load.status = status
		load.err = err
		load.stale = stale
		delete(spendAlertsCache.loads, key)
		close(load.done)
		spendAlertsCache.Unlock()

		if err != nil {
			return spendAlertsCacheEntry{}, status, err
		}
		if stale {
			continue
		}
		return cloneSpendAlertsCacheEntry(entry), status, nil
	}
}

func fetchSpendAlerts(ctx context.Context, c *OpenAIClient, basePath string) (spendAlertsCacheEntry, int, error) {
	var alerts []SpendAlertAPI
	after := ""
	status := http.StatusOK

	for {
		query := url.Values{"limit": []string{"100"}}
		if after != "" {
			query.Set("after", after)
		}

		var response SpendAlertsListAPI
		status, err := projectSettingsRequest(
			ctx,
			c,
			http.MethodGet,
			basePath+"?"+query.Encode(),
			nil,
			&response,
		)
		if err != nil {
			return spendAlertsCacheEntry{}, status, err
		}
		alerts = append(alerts, response.Data...)

		if !response.HasMore || response.LastID == "" || response.LastID == after {
			break
		}
		after = response.LastID
	}

	return newSpendAlertsCacheEntry(alerts, time.Now().Add(projectSettingsCacheTTL)), status, nil
}

func newSpendAlertsCacheEntry(alerts []SpendAlertAPI, expiresAt time.Time) spendAlertsCacheEntry {
	entry := spendAlertsCacheEntry{
		ExpiresAt: expiresAt,
		Alerts:    make([]SpendAlertAPI, 0, len(alerts)),
		ByID:      make(map[string]SpendAlertAPI, len(alerts)),
	}
	for _, alert := range alerts {
		cached := cloneSpendAlert(alert)
		entry.Alerts = append(entry.Alerts, cached)
		if _, ok := entry.ByID[cached.ID]; !ok {
			entry.ByID[cached.ID] = cached
		}
	}
	return entry
}

func cloneSpendAlertsCacheEntry(value spendAlertsCacheEntry) spendAlertsCacheEntry {
	return newSpendAlertsCacheEntry(value.Alerts, value.ExpiresAt)
}

func readSpendAlertsFileCache(key string) (spendAlertsCacheEntry, bool) {
	path := projectSettingsCachePath("spend-alerts-list", key)
	body, err := os.ReadFile(path)
	if err != nil {
		return spendAlertsCacheEntry{}, false
	}

	var item spendAlertsFileCacheItem
	if err := json.Unmarshal(body, &item); err != nil || !time.Now().Before(item.ExpiresAt) {
		_ = os.Remove(path)
		return spendAlertsCacheEntry{}, false
	}
	return newSpendAlertsCacheEntry(item.Alerts, item.ExpiresAt), true
}

func writeSpendAlertsFileCache(key string, entry spendAlertsCacheEntry) error {
	path := projectSettingsCachePath("spend-alerts-list", key)
	body, err := json.Marshal(spendAlertsFileCacheItem{
		ExpiresAt: entry.ExpiresAt,
		Alerts:    entry.Alerts,
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0600)
}

func projectSettingsCachePath(kind, key string) string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	return filepath.Join(cacheDir, "terraform-provider-openai", "project-settings", kind, key+".json")
}

func readProjectSettingsFileCache[T any](kind, key string) (projectSettingsCacheEntry[T], bool) {
	path := projectSettingsCachePath(kind, key)
	body, err := os.ReadFile(path)
	if err != nil {
		return projectSettingsCacheEntry[T]{}, false
	}
	var entry projectSettingsCacheEntry[T]
	if err := json.Unmarshal(body, &entry); err != nil || !time.Now().Before(entry.ExpiresAt) {
		_ = os.Remove(path)
		return projectSettingsCacheEntry[T]{}, false
	}
	return entry, true
}

func writeProjectSettingsFileCache[T any](kind, key string, entry projectSettingsCacheEntry[T]) error {
	path := projectSettingsCachePath(kind, key)
	body, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0600)
}

func cloneSpendAlert(value SpendAlertAPI) SpendAlertAPI {
	value.NotificationChannel.Recipients = append([]string(nil), value.NotificationChannel.Recipients...)
	if value.NotificationChannel.SubjectPrefix != nil {
		subjectPrefix := *value.NotificationChannel.SubjectPrefix
		value.NotificationChannel.SubjectPrefix = &subjectPrefix
	}
	return value
}
