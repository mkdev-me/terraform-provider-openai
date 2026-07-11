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

type ProjectSpendAlertNotificationAPI struct {
	Type          string   `json:"type"`
	Recipients    []string `json:"recipients"`
	SubjectPrefix *string  `json:"subject_prefix,omitempty"`
}

type ProjectSpendAlertAPI struct {
	ID                  string                           `json:"id"`
	Object              string                           `json:"object"`
	ThresholdAmount     int64                            `json:"threshold_amount"`
	Currency            string                           `json:"currency"`
	Interval            string                           `json:"interval"`
	NotificationChannel ProjectSpendAlertNotificationAPI `json:"notification_channel"`
}

type ProjectSpendAlertsListAPI struct {
	Object  string                 `json:"object"`
	Data    []ProjectSpendAlertAPI `json:"data"`
	FirstID string                 `json:"first_id"`
	LastID  string                 `json:"last_id"`
	HasMore bool                   `json:"has_more"`
	Next    *string                `json:"next"`
}

type projectSpendAlertRequest struct {
	ThresholdAmount     int64                            `json:"threshold_amount"`
	Currency            string                           `json:"currency"`
	Interval            string                           `json:"interval"`
	NotificationChannel ProjectSpendAlertNotificationAPI `json:"notification_channel"`
}

type projectSettingsCacheEntry[T any] struct {
	Value     T         `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

type projectSpendAlertsCacheEntry struct {
	ExpiresAt time.Time
	Alerts    []ProjectSpendAlertAPI
	ByID      map[string]ProjectSpendAlertAPI
}

type projectSpendAlertsCacheLoad struct {
	done   chan struct{}
	entry  projectSpendAlertsCacheEntry
	status int
	err    error
	stale  bool
}

type projectSpendAlertsFileCacheItem struct {
	ExpiresAt time.Time              `json:"expires_at"`
	Alerts    []ProjectSpendAlertAPI `json:"alerts"`
}

var projectSettingsCache = struct {
	sync.Mutex
	modelPermissions map[string]projectSettingsCacheEntry[ProjectModelPermissionsAPI]
}{
	modelPermissions: map[string]projectSettingsCacheEntry[ProjectModelPermissionsAPI]{},
}

var projectSpendAlertsCache = struct {
	sync.Mutex
	entries    map[string]projectSpendAlertsCacheEntry
	loads      map[string]*projectSpendAlertsCacheLoad
	generation map[string]uint64
}{
	entries:    map[string]projectSpendAlertsCacheEntry{},
	loads:      map[string]*projectSpendAlertsCacheLoad{},
	generation: map[string]uint64{},
}

func projectSettingsIdentityKey(c *OpenAIClient, values ...string) string {
	parts := []string{adminBaseURL(c), c.OrganizationID, adminAPIKey(c)}
	parts = append(parts, values...)
	hash := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return fmt.Sprintf("%x", hash)
}

func invalidateProjectModelPermissionsCache(c *OpenAIClient, projectID string) {
	key := projectSettingsIdentityKey(c, projectID, "model_permissions")
	projectSettingsCache.Lock()
	delete(projectSettingsCache.modelPermissions, key)
	projectSettingsCache.Unlock()
	_ = os.Remove(projectSettingsCachePath("model-permissions", key))
}

func invalidateProjectSpendAlertsCache(c *OpenAIClient, projectID string) {
	key := projectSettingsIdentityKey(c, projectID, "spend_alerts")
	projectSpendAlertsCache.Lock()
	delete(projectSpendAlertsCache.entries, key)
	projectSpendAlertsCache.generation[key]++
	projectSpendAlertsCache.Unlock()
	_ = os.Remove(projectSettingsCachePath("spend-alerts-list", key))
}

func resetProjectSettingsCacheForTest() {
	projectSettingsCache.Lock()
	projectSettingsCache.modelPermissions = map[string]projectSettingsCacheEntry[ProjectModelPermissionsAPI]{}
	projectSettingsCache.Unlock()

	projectSpendAlertsCache.Lock()
	projectSpendAlertsCache.entries = map[string]projectSpendAlertsCacheEntry{}
	projectSpendAlertsCache.loads = map[string]*projectSpendAlertsCacheLoad{}
	projectSpendAlertsCache.generation = map[string]uint64{}
	projectSpendAlertsCache.Unlock()
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

func cachedProjectModelPermissions(ctx context.Context, c *OpenAIClient, projectID string) (*ProjectModelPermissionsAPI, int, error) {
	key := projectSettingsIdentityKey(c, projectID, "model_permissions")
	projectSettingsCache.Lock()
	defer projectSettingsCache.Unlock()

	if cached, ok := projectSettingsCache.modelPermissions[key]; ok && time.Now().Before(cached.ExpiresAt) {
		value := cached.Value
		value.ModelIDs = append([]string(nil), value.ModelIDs...)
		return &value, http.StatusOK, nil
	}
	if cached, ok := readProjectSettingsFileCache[ProjectModelPermissionsAPI]("model-permissions", key); ok {
		projectSettingsCache.modelPermissions[key] = cached
		value := cached.Value
		value.ModelIDs = append([]string(nil), value.ModelIDs...)
		return &value, http.StatusOK, nil
	}

	var value ProjectModelPermissionsAPI
	status, err := projectSettingsRequest(ctx, c, http.MethodGet, "/v1/organization/projects/"+projectID+"/model_permissions", nil, &value)
	if err != nil {
		return nil, status, err
	}
	entry := projectSettingsCacheEntry[ProjectModelPermissionsAPI]{Value: value, ExpiresAt: time.Now().Add(projectSettingsCacheTTL)}
	projectSettingsCache.modelPermissions[key] = entry
	_ = writeProjectSettingsFileCache("model-permissions", key, entry)
	value.ModelIDs = append([]string(nil), value.ModelIDs...)
	return &value, status, nil
}

func cachedProjectSpendAlert(ctx context.Context, c *OpenAIClient, projectID, alertID string) (*ProjectSpendAlertAPI, int, error) {
	entry, status, err := cachedProjectSpendAlerts(ctx, c, projectID)
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
	value = cloneProjectSpendAlert(value)
	return &value, status, nil
}

func cachedProjectSpendAlerts(ctx context.Context, c *OpenAIClient, projectID string) (projectSpendAlertsCacheEntry, int, error) {
	if c == nil || c.OpenAIClient == nil {
		return projectSpendAlertsCacheEntry{}, 0, fmt.Errorf("admin API key is required")
	}

	key := projectSettingsIdentityKey(c, projectID, "spend_alerts")

	for {
		if cached, ok := readProjectSpendAlertsFileCache(key); ok {
			projectSpendAlertsCache.Lock()
			projectSpendAlertsCache.entries[key] = cached
			projectSpendAlertsCache.Unlock()
			return cloneProjectSpendAlertsCacheEntry(cached), http.StatusOK, nil
		}

		projectSpendAlertsCache.Lock()
		if cached, ok := projectSpendAlertsCache.entries[key]; ok {
			if time.Now().Before(cached.ExpiresAt) {
				projectSpendAlertsCache.Unlock()
				return cloneProjectSpendAlertsCacheEntry(cached), http.StatusOK, nil
			}
			delete(projectSpendAlertsCache.entries, key)
		}

		if load, ok := projectSpendAlertsCache.loads[key]; ok {
			projectSpendAlertsCache.Unlock()
			<-load.done
			if load.err != nil {
				return projectSpendAlertsCacheEntry{}, load.status, load.err
			}
			if load.stale {
				continue
			}
			return cloneProjectSpendAlertsCacheEntry(load.entry), load.status, nil
		}

		generation := projectSpendAlertsCache.generation[key]
		load := &projectSpendAlertsCacheLoad{done: make(chan struct{})}
		projectSpendAlertsCache.loads[key] = load
		projectSpendAlertsCache.Unlock()

		entry, status, err := fetchProjectSpendAlerts(ctx, c, projectID)

		projectSpendAlertsCache.Lock()
		stale := projectSpendAlertsCache.generation[key] != generation
		if err == nil && !stale {
			projectSpendAlertsCache.entries[key] = entry
			_ = writeProjectSpendAlertsFileCache(key, entry)
		}
		load.entry = entry
		load.status = status
		load.err = err
		load.stale = stale
		delete(projectSpendAlertsCache.loads, key)
		close(load.done)
		projectSpendAlertsCache.Unlock()

		if err != nil {
			return projectSpendAlertsCacheEntry{}, status, err
		}
		if stale {
			continue
		}
		return cloneProjectSpendAlertsCacheEntry(entry), status, nil
	}
}

func fetchProjectSpendAlerts(ctx context.Context, c *OpenAIClient, projectID string) (projectSpendAlertsCacheEntry, int, error) {
	var alerts []ProjectSpendAlertAPI
	after := ""
	status := http.StatusOK

	for {
		query := url.Values{"limit": []string{"100"}}
		if after != "" {
			query.Set("after", after)
		}

		var response ProjectSpendAlertsListAPI
		status, err := projectSettingsRequest(
			ctx,
			c,
			http.MethodGet,
			"/v1/organization/projects/"+projectID+"/spend_alerts?"+query.Encode(),
			nil,
			&response,
		)
		if err != nil {
			return projectSpendAlertsCacheEntry{}, status, err
		}
		alerts = append(alerts, response.Data...)

		if !response.HasMore {
			break
		}
		next := response.LastID
		if next == "" && response.Next != nil {
			next = *response.Next
		}
		if next == "" || next == after {
			break
		}
		after = next
	}

	return newProjectSpendAlertsCacheEntry(alerts, time.Now().Add(projectSettingsCacheTTL)), status, nil
}

func newProjectSpendAlertsCacheEntry(alerts []ProjectSpendAlertAPI, expiresAt time.Time) projectSpendAlertsCacheEntry {
	entry := projectSpendAlertsCacheEntry{
		ExpiresAt: expiresAt,
		Alerts:    make([]ProjectSpendAlertAPI, 0, len(alerts)),
		ByID:      make(map[string]ProjectSpendAlertAPI, len(alerts)),
	}
	for _, alert := range alerts {
		cached := cloneProjectSpendAlert(alert)
		entry.Alerts = append(entry.Alerts, cached)
		if _, ok := entry.ByID[cached.ID]; !ok {
			entry.ByID[cached.ID] = cached
		}
	}
	return entry
}

func cloneProjectSpendAlertsCacheEntry(value projectSpendAlertsCacheEntry) projectSpendAlertsCacheEntry {
	return newProjectSpendAlertsCacheEntry(value.Alerts, value.ExpiresAt)
}

func readProjectSpendAlertsFileCache(key string) (projectSpendAlertsCacheEntry, bool) {
	path := projectSettingsCachePath("spend-alerts-list", key)
	body, err := os.ReadFile(path)
	if err != nil {
		return projectSpendAlertsCacheEntry{}, false
	}

	var item projectSpendAlertsFileCacheItem
	if err := json.Unmarshal(body, &item); err != nil || !time.Now().Before(item.ExpiresAt) {
		_ = os.Remove(path)
		return projectSpendAlertsCacheEntry{}, false
	}
	return newProjectSpendAlertsCacheEntry(item.Alerts, item.ExpiresAt), true
}

func writeProjectSpendAlertsFileCache(key string, entry projectSpendAlertsCacheEntry) error {
	path := projectSettingsCachePath("spend-alerts-list", key)
	body, err := json.Marshal(projectSpendAlertsFileCacheItem{
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

func cloneProjectSpendAlert(value ProjectSpendAlertAPI) ProjectSpendAlertAPI {
	value.NotificationChannel.Recipients = append([]string(nil), value.NotificationChannel.Recipients...)
	if value.NotificationChannel.SubjectPrefix != nil {
		subjectPrefix := *value.NotificationChannel.SubjectPrefix
		value.NotificationChannel.SubjectPrefix = &subjectPrefix
	}
	return value
}
