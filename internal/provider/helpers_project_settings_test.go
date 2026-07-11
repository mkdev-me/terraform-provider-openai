package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProjectSettingsRequestBodiesExcludeResponseFields(t *testing.T) {
	modelPermissionsBody, err := json.Marshal(projectModelPermissionsRequest{
		Mode:     "allow_list",
		ModelIDs: []string{"gpt-4.1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var modelPermissionsPayload map[string]any
	if err := json.Unmarshal(modelPermissionsBody, &modelPermissionsPayload); err != nil {
		t.Fatal(err)
	}
	wantModelPermissionsPayload := map[string]any{
		"mode":      "allow_list",
		"model_ids": []any{"gpt-4.1"},
	}
	if !reflect.DeepEqual(modelPermissionsPayload, wantModelPermissionsPayload) {
		t.Fatalf("unexpected model permissions request body: %s", modelPermissionsBody)
	}

	spendAlertBody, err := json.Marshal(projectSpendAlertRequest{
		ThresholdAmount: 100000,
		Currency:        "USD",
		Interval:        "month",
		NotificationChannel: ProjectSpendAlertNotificationAPI{
			Type:       "email",
			Recipients: []string{"finance@example.com"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var spendAlertPayload map[string]any
	if err := json.Unmarshal(spendAlertBody, &spendAlertPayload); err != nil {
		t.Fatal(err)
	}
	wantSpendAlertPayload := map[string]any{
		"threshold_amount": float64(100000),
		"currency":         "USD",
		"interval":         "month",
		"notification_channel": map[string]any{
			"type":       "email",
			"recipients": []any{"finance@example.com"},
		},
	}
	if !reflect.DeepEqual(spendAlertPayload, wantSpendAlertPayload) {
		t.Fatalf("unexpected spend alert request body: %s", spendAlertBody)
	}
}

func TestCachedProjectModelPermissions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	resetProjectSettingsCacheForTest()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/v1/organization/projects/proj_test/model_permissions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-admin-key" {
			t.Fatalf("unexpected Authorization header: %q", got)
		}
		_ = json.NewEncoder(w).Encode(ProjectModelPermissionsAPI{
			Mode: "deny_list", ModelIDs: []string{"o3"}, Object: "project.model_permissions",
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	for i := 0; i < 2; i++ {
		value, status, err := cachedProjectModelPermissions(context.Background(), c, "proj_test")
		if err != nil || status != http.StatusOK {
			t.Fatalf("cachedProjectModelPermissions() status=%d error=%v", status, err)
		}
		if value.Mode != "deny_list" || len(value.ModelIDs) != 1 || value.ModelIDs[0] != "o3" {
			t.Fatalf("unexpected response: %#v", value)
		}
		if i == 0 {
			// Prove the second read can use the persistent cache, not only memory.
			resetProjectSettingsCacheForTest()
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one API call, got %d", calls.Load())
	}

	invalidateProjectModelPermissionsCache(c, "proj_test")
	_, _, err := cachedProjectModelPermissions(context.Background(), c, "proj_test")
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected cache invalidation to trigger a second API call, got %d", calls.Load())
	}
}

func TestCachedProjectSpendAlert(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	resetProjectSettingsCacheForTest()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/v1/organization/projects/proj_test/spend_alerts" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "100" {
			t.Fatalf("unexpected list limit: %q", got)
		}
		_ = json.NewEncoder(w).Encode(ProjectSpendAlertsListAPI{
			Object: "list",
			Data: []ProjectSpendAlertAPI{
				{
					ID: "alert_test", Object: "project.spend_alert", ThresholdAmount: 100000,
					Currency: "USD", Interval: "month",
					NotificationChannel: ProjectSpendAlertNotificationAPI{Type: "email", Recipients: []string{"finance@example.com"}},
				},
				{
					ID: "alert_other", Object: "project.spend_alert", ThresholdAmount: 200000,
					Currency: "USD", Interval: "month",
					NotificationChannel: ProjectSpendAlertNotificationAPI{Type: "email", Recipients: []string{"owner@example.com"}},
				},
			},
			HasMore: false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	for _, alertID := range []string{"alert_test", "alert_other"} {
		value, status, err := cachedProjectSpendAlert(context.Background(), c, "proj_test", alertID)
		if err != nil || status != http.StatusOK {
			t.Fatalf("cachedProjectSpendAlert() status=%d error=%v", status, err)
		}
		if value.ID != alertID {
			t.Fatalf("unexpected response: %#v", value)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one list API call for multiple alerts, got %d", calls.Load())
	}

	// Prove the second refresh can use the persistent list cache, not only memory.
	resetProjectSettingsCacheForTest()
	value, status, err := cachedProjectSpendAlert(context.Background(), c, "proj_test", "alert_test")
	if err != nil || status != http.StatusOK || value.ID != "alert_test" {
		t.Fatalf("cachedProjectSpendAlert() from persistent cache status=%d error=%v value=%#v", status, err, value)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected persistent list cache to avoid another API call, got %d", calls.Load())
	}

	invalidateProjectSpendAlertsCache(c, "proj_test")
	_, _, err = cachedProjectSpendAlert(context.Background(), c, "proj_test", "alert_test")
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected cache invalidation to trigger a second list API call, got %d", calls.Load())
	}
}

func TestCachedProjectSpendAlertsCoalesceInFlight(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	resetProjectSettingsCacheForTest()
	var calls atomic.Int32
	requestStarted := make(chan struct{}, 2)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/organization/projects/proj_test/spend_alerts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			return
		}
		calls.Add(1)
		requestStarted <- struct{}{}
		<-release
		_ = json.NewEncoder(w).Encode(ProjectSpendAlertsListAPI{
			Object: "list",
			Data: []ProjectSpendAlertAPI{
				{ID: "alert_test", Object: "project.spend_alert"},
				{ID: "alert_other", Object: "project.spend_alert"},
			},
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	type result struct {
		value  *ProjectSpendAlertAPI
		status int
		err    error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for _, alertID := range []string{"alert_test", "alert_other"} {
		wg.Add(1)
		go func(alertID string) {
			defer wg.Done()
			value, status, err := cachedProjectSpendAlert(context.Background(), c, "proj_test", alertID)
			results <- result{value: value, status: status, err: err}
		}(alertID)
	}

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for spend-alert list request")
	}
	select {
	case <-requestStarted:
		t.Fatal("started a second spend-alert list request while the first was in flight")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	wg.Wait()
	close(results)

	for got := range results {
		if got.err != nil || got.status != http.StatusOK || got.value == nil {
			t.Fatalf("cachedProjectSpendAlert() status=%d error=%v value=%#v", got.status, got.err, got.value)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one in-flight-coalesced list API call, got %d", calls.Load())
	}
}

func TestProjectSettingsResourcesRegistered(t *testing.T) {
	wanted := map[string]bool{
		"openai_project_model_permissions": false,
		"openai_project_spend_alert":       false,
	}
	p := &FrameworkProvider{}
	for _, constructor := range p.Resources(context.Background()) {
		var response resource.MetadataResponse
		constructor().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openai"}, &response)
		if _, ok := wanted[response.TypeName]; ok {
			wanted[response.TypeName] = true
		}
	}
	for name, found := range wanted {
		if !found {
			t.Errorf("resource %s is not registered", name)
		}
	}
}
