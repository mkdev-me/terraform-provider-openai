package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	schemavalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func testPlanFromModel(t *testing.T, schema rschema.Schema, model any) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("could not build plan: %v", diags)
	}
	return plan
}

func testStateFromModel(t *testing.T, schema rschema.Schema, model any) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: schema}
	if diags := state.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("could not build state: %v", diags)
	}
	return state
}

func emptyTestState(schema rschema.Schema) tfsdk.State {
	ctx := context.Background()
	return tfsdk.State{
		Schema: schema,
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
	}
}

func decodeRequestJSON(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("could not decode request body %q: %v", body, err)
	}
	return payload
}

func assertAdminAuthorization(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("Authorization"); got != "Bearer test-admin-key" {
		t.Fatalf("Authorization = %q, want admin key", got)
	}
}

func TestProjectSpendLimitResourceLifecycle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	resetProjectSettingsCacheForTest()
	var posts, gets, deletes int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAdminAuthorization(t, r)
		if r.URL.Path != "/v1/organization/projects/proj_test/spend_limit" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		switch r.Method {
		case http.MethodPost:
			posts++
			payload := decodeRequestJSON(t, r)
			if len(payload) != 3 || payload["currency"] != "USD" || payload["interval"] != "month" {
				t.Fatalf("unexpected spend limit payload: %#v", payload)
			}
			_ = json.NewEncoder(w).Encode(SpendLimitAPI{
				Object:          "project.spend_limit",
				ThresholdAmount: int64(payload["threshold_amount"].(float64)),
				Currency:        "USD",
				Interval:        "month",
				Enforcement:     SpendLimitEnforcementAPI{Status: "enforcing"},
			})
		case http.MethodGet:
			gets++
			_ = json.NewEncoder(w).Encode(SpendLimitAPI{
				Object:          "project.spend_limit",
				ThresholdAmount: 15000,
				Currency:        "USD",
				Interval:        "month",
				Enforcement:     SpendLimitEnforcementAPI{Status: "enforcing"},
			})
		case http.MethodDelete:
			deletes++
			_ = json.NewEncoder(w).Encode(map[string]any{"object": "project.spend_limit.deleted", "deleted": true})
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	r := &ProjectSpendLimitResource{client: newTestOpenAIClient(server.URL)}
	schema := currentSchema(t, r)
	createModel := ProjectSpendLimitResourceModel{
		ID:              types.StringUnknown(),
		ProjectID:       types.StringValue("proj_test"),
		ThresholdAmount: types.Int64Value(10000),
		Currency:        types.StringValue("USD"),
		Interval:        types.StringValue("month"),
		Object:          types.StringUnknown(),
	}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: testPlanFromModel(t, schema, &createModel)}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create diagnostics: %v", createResp.Diagnostics)
	}

	var state ProjectSpendLimitResourceModel
	if diags := createResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("could not read create state: %v", diags)
	}
	if state.ID.ValueString() != "proj_test" || state.Enforcement == nil || state.Enforcement.Status.ValueString() != "enforcing" {
		t.Fatalf("unexpected create state: %#v", state)
	}

	state.ThresholdAmount = types.Int64Value(15000)
	updateResp := resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: testPlanFromModel(t, schema, &state)}, &updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update diagnostics: %v", updateResp.Diagnostics)
	}

	readResp := resource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(ctx, resource.ReadRequest{State: updateResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read diagnostics: %v", readResp.Diagnostics)
	}
	var refreshed ProjectSpendLimitResourceModel
	if diags := readResp.State.Get(ctx, &refreshed); diags.HasError() {
		t.Fatalf("could not read refreshed state: %v", diags)
	}
	if refreshed.ThresholdAmount.ValueInt64() != 15000 || refreshed.Object.ValueString() != "project.spend_limit" {
		t.Fatalf("unexpected refreshed state: %#v", refreshed)
	}

	deleteResp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: readResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete diagnostics: %v", deleteResp.Diagnostics)
	}
	if posts != 2 || gets != 1 || deletes != 1 {
		t.Fatalf("unexpected requests: posts=%d gets=%d deletes=%d", posts, gets, deletes)
	}
}

func TestOrganizationSpendLimitResourceLifecycle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	resetProjectSettingsCacheForTest()
	var posts, gets, deletes int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAdminAuthorization(t, r)
		if r.URL.Path != "/v1/organization/spend_limit" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		switch r.Method {
		case http.MethodPost:
			posts++
			payload := decodeRequestJSON(t, r)
			_ = json.NewEncoder(w).Encode(SpendLimitAPI{
				Object:          "organization.spend_limit",
				ThresholdAmount: int64(payload["threshold_amount"].(float64)),
				Currency:        "USD",
				Interval:        "month",
				Enforcement:     SpendLimitEnforcementAPI{Status: "inactive"},
			})
		case http.MethodGet:
			gets++
			_ = json.NewEncoder(w).Encode(SpendLimitAPI{
				Object:          "organization.spend_limit",
				ThresholdAmount: 30000,
				Currency:        "USD",
				Interval:        "month",
				Enforcement:     SpendLimitEnforcementAPI{Status: "enforcing"},
			})
		case http.MethodDelete:
			deletes++
			_ = json.NewEncoder(w).Encode(map[string]any{"object": "organization.spend_limit.deleted", "deleted": true})
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	r := &OrganizationSpendLimitResource{client: newTestOpenAIClient(server.URL)}
	schema := currentSchema(t, r)
	model := OrganizationSpendLimitResourceModel{
		ID:              types.StringUnknown(),
		ThresholdAmount: types.Int64Value(20000),
		Currency:        types.StringValue("USD"),
		Interval:        types.StringValue("month"),
		Object:          types.StringUnknown(),
	}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: testPlanFromModel(t, schema, &model)}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create diagnostics: %v", createResp.Diagnostics)
	}
	var state OrganizationSpendLimitResourceModel
	if diags := createResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("could not read create state: %v", diags)
	}
	if state.ID.ValueString() != organizationSpendLimitID {
		t.Fatalf("ID = %q, want %q", state.ID.ValueString(), organizationSpendLimitID)
	}

	state.ThresholdAmount = types.Int64Value(30000)
	updateResp := resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: testPlanFromModel(t, schema, &state)}, &updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update diagnostics: %v", updateResp.Diagnostics)
	}
	readResp := resource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(ctx, resource.ReadRequest{State: updateResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read diagnostics: %v", readResp.Diagnostics)
	}
	var refreshed OrganizationSpendLimitResourceModel
	if diags := readResp.State.Get(ctx, &refreshed); diags.HasError() {
		t.Fatalf("could not read refreshed state: %v", diags)
	}
	if refreshed.ThresholdAmount.ValueInt64() != 30000 || refreshed.Enforcement.Status.ValueString() != "enforcing" {
		t.Fatalf("unexpected refreshed state: %#v", refreshed)
	}

	deleteResp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: readResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete diagnostics: %v", deleteResp.Diagnostics)
	}
	if posts != 2 || gets != 1 || deletes != 1 {
		t.Fatalf("unexpected requests: posts=%d gets=%d deletes=%d", posts, gets, deletes)
	}
}

func TestOrganizationSpendAlertResourceLifecycle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	resetProjectSettingsCacheForTest()
	var creates, updates, lists, deletes int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAdminAuthorization(t, r)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/organization/spend_alerts":
			creates++
			payload := decodeRequestJSON(t, r)
			if payload["threshold_amount"] != float64(0) {
				t.Fatalf("threshold_amount = %#v, want zero", payload["threshold_amount"])
			}
			notification := payload["notification_channel"].(map[string]any)
			if !reflect.DeepEqual(notification["recipients"], []any{"finance@example.com"}) {
				t.Fatalf("unexpected recipients: %#v", notification["recipients"])
			}
			_ = json.NewEncoder(w).Encode(SpendAlertAPI{
				ID:              "alert_test",
				Object:          "organization.spend_alert",
				ThresholdAmount: 0,
				Currency:        "USD",
				Interval:        "month",
				NotificationChannel: SpendAlertNotificationAPI{
					Type:       "email",
					Recipients: []string{"finance@example.com"},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/organization/spend_alerts/alert_test":
			updates++
			payload := decodeRequestJSON(t, r)
			_ = json.NewEncoder(w).Encode(SpendAlertAPI{
				ID:              "alert_test",
				Object:          "organization.spend_alert",
				ThresholdAmount: int64(payload["threshold_amount"].(float64)),
				Currency:        "USD",
				Interval:        "month",
				NotificationChannel: SpendAlertNotificationAPI{
					Type:       "email",
					Recipients: []string{"finance@example.com"},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/organization/spend_alerts":
			lists++
			if r.URL.Query().Get("limit") != "100" {
				t.Fatalf("unexpected list query: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(SpendAlertsListAPI{
				Object: "list",
				Data: []SpendAlertAPI{{
					ID:              "alert_test",
					Object:          "organization.spend_alert",
					ThresholdAmount: 500,
					Currency:        "USD",
					Interval:        "month",
					NotificationChannel: SpendAlertNotificationAPI{
						Type:       "email",
						Recipients: []string{"finance@example.com"},
					},
				}},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/organization/spend_alerts/alert_test":
			deletes++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "alert_test", "object": "organization.spend_alert.deleted", "deleted": true,
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	ctx := context.Background()
	r := &OrganizationSpendAlertResource{client: newTestOpenAIClient(server.URL)}
	schema := currentSchema(t, r)
	model := OrganizationSpendAlertResourceModel{
		ID:              types.StringUnknown(),
		AlertID:         types.StringUnknown(),
		ThresholdAmount: types.Int64Value(0),
		Currency:        types.StringValue("USD"),
		Interval:        types.StringValue("month"),
		NotificationChannel: &SpendAlertNotificationModel{
			Type:          types.StringValue("email"),
			Recipients:    modelIDsToSet([]string{"finance@example.com"}),
			SubjectPrefix: types.StringNull(),
		},
		Object: types.StringUnknown(),
	}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: testPlanFromModel(t, schema, &model)}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create diagnostics: %v", createResp.Diagnostics)
	}
	var state OrganizationSpendAlertResourceModel
	if diags := createResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("could not read create state: %v", diags)
	}
	if state.ID.ValueString() != "alert_test" || state.AlertID.ValueString() != "alert_test" {
		t.Fatalf("unexpected IDs: id=%q alert_id=%q", state.ID.ValueString(), state.AlertID.ValueString())
	}

	state.ThresholdAmount = types.Int64Value(500)
	updateResp := resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: testPlanFromModel(t, schema, &state)}, &updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update diagnostics: %v", updateResp.Diagnostics)
	}
	readResp := resource.ReadResponse{State: tfsdk.State{Schema: schema}}
	r.Read(ctx, resource.ReadRequest{State: updateResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read diagnostics: %v", readResp.Diagnostics)
	}
	var refreshed OrganizationSpendAlertResourceModel
	if diags := readResp.State.Get(ctx, &refreshed); diags.HasError() {
		t.Fatalf("could not read refreshed state: %v", diags)
	}
	if refreshed.ThresholdAmount.ValueInt64() != 500 {
		t.Fatalf("unexpected refreshed state: %#v", refreshed)
	}

	deleteResp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: readResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete diagnostics: %v", deleteResp.Diagnostics)
	}
	if creates != 1 || updates != 1 || lists != 1 || deletes != 1 {
		t.Fatalf("unexpected requests: creates=%d updates=%d lists=%d deletes=%d", creates, updates, lists, deletes)
	}
}

func validateInt64(validator schemavalidator.Int64, value int64) bool {
	var resp schemavalidator.Int64Response
	validator.ValidateInt64(context.Background(), schemavalidator.Int64Request{
		Path:        path.Root("value"),
		ConfigValue: types.Int64Value(value),
	}, &resp)
	return resp.Diagnostics.HasError()
}

func validateString(validator schemavalidator.String, value string) bool {
	var resp schemavalidator.StringResponse
	validator.ValidateString(context.Background(), schemavalidator.StringRequest{
		Path:        path.Root("value"),
		ConfigValue: types.StringValue(value),
	}, &resp)
	return resp.Diagnostics.HasError()
}

func validateSet(validator schemavalidator.Set, value types.Set) bool {
	var resp schemavalidator.SetResponse
	validator.ValidateSet(context.Background(), schemavalidator.SetRequest{
		Path:        path.Root("value"),
		ConfigValue: value,
	}, &resp)
	return resp.Diagnostics.HasError()
}

func TestSpendGovernanceSchemas(t *testing.T) {
	projectLimitSchema := currentSchema(t, NewProjectSpendLimitResource())
	projectThreshold := projectLimitSchema.Attributes["threshold_amount"].(rschema.Int64Attribute)
	if !validateInt64(projectThreshold.Validators[0], 0) || validateInt64(projectThreshold.Validators[0], 1) {
		t.Fatal("project spend limit threshold validator does not enforce minimum 1")
	}
	if projectLimitSchema.Attributes["project_id"].(rschema.StringAttribute).PlanModifiers == nil {
		t.Fatal("project_id must require replacement")
	}
	enforcement := projectLimitSchema.Attributes["enforcement"].(rschema.SingleNestedAttribute)
	if !enforcement.Computed || !enforcement.Attributes["status"].(rschema.StringAttribute).Computed {
		t.Fatal("enforcement.status must be computed")
	}

	organizationLimitSchema := currentSchema(t, NewOrganizationSpendLimitResource())
	organizationThreshold := organizationLimitSchema.Attributes["threshold_amount"].(rschema.Int64Attribute)
	if !validateInt64(organizationThreshold.Validators[0], 0) || validateInt64(organizationThreshold.Validators[0], 1) {
		t.Fatal("organization spend limit threshold validator does not enforce minimum 1")
	}
	for _, schema := range []rschema.Schema{projectLimitSchema, organizationLimitSchema} {
		currency := schema.Attributes["currency"].(rschema.StringAttribute)
		interval := schema.Attributes["interval"].(rschema.StringAttribute)
		if currency.Default == nil || interval.Default == nil {
			t.Fatal("spend limit currency and interval defaults must be configured")
		}
		if validateString(currency.Validators[0], "USD") || !validateString(currency.Validators[0], "EUR") {
			t.Fatal("currency validator must only accept USD")
		}
		if validateString(interval.Validators[0], "month") || !validateString(interval.Validators[0], "day") {
			t.Fatal("interval validator must only accept month")
		}
	}

	alertSchema := currentSchema(t, NewOrganizationSpendAlertResource())
	alertThreshold := alertSchema.Attributes["threshold_amount"].(rschema.Int64Attribute)
	if !validateInt64(alertThreshold.Validators[0], -1) || validateInt64(alertThreshold.Validators[0], 0) {
		t.Fatal("organization spend alert threshold validator does not enforce minimum 0")
	}
	notification := alertSchema.Attributes["notification_channel"].(rschema.SingleNestedAttribute)
	channelType := notification.Attributes["type"].(rschema.StringAttribute)
	if validateString(channelType.Validators[0], "email") || !validateString(channelType.Validators[0], "sms") {
		t.Fatal("notification type validator must only accept email")
	}
	recipients := notification.Attributes["recipients"].(rschema.SetAttribute)
	if !validateSet(recipients.Validators[0], types.SetValueMust(types.StringType, []attr.Value{})) {
		t.Fatal("recipients validator must reject an empty set")
	}
}

func TestSpendGovernanceImports(t *testing.T) {
	ctx := context.Background()

	projectResource := &ProjectSpendLimitResource{}
	projectSchema := currentSchema(t, projectResource)
	projectResp := resource.ImportStateResponse{State: emptyTestState(projectSchema)}
	projectResource.ImportState(ctx, resource.ImportStateRequest{ID: "proj_test"}, &projectResp)
	if projectResp.Diagnostics.HasError() {
		t.Fatalf("project import diagnostics: %v", projectResp.Diagnostics)
	}
	var projectState ProjectSpendLimitResourceModel
	if diags := projectResp.State.Get(ctx, &projectState); diags.HasError() {
		t.Fatalf("could not read project import state: %v", diags)
	}
	if projectState.ID.ValueString() != "proj_test" || projectState.ProjectID.ValueString() != "proj_test" {
		t.Fatalf("unexpected project import state: %#v", projectState)
	}

	organizationResource := &OrganizationSpendLimitResource{}
	organizationSchema := currentSchema(t, organizationResource)
	invalidOrganizationResp := resource.ImportStateResponse{State: emptyTestState(organizationSchema)}
	organizationResource.ImportState(ctx, resource.ImportStateRequest{ID: "org_123"}, &invalidOrganizationResp)
	if !invalidOrganizationResp.Diagnostics.HasError() {
		t.Fatal("organization spend limit import must reject IDs other than organization")
	}
	organizationResp := resource.ImportStateResponse{State: emptyTestState(organizationSchema)}
	organizationResource.ImportState(ctx, resource.ImportStateRequest{ID: organizationSpendLimitID}, &organizationResp)
	if organizationResp.Diagnostics.HasError() {
		t.Fatalf("organization import diagnostics: %v", organizationResp.Diagnostics)
	}

	alertResource := &OrganizationSpendAlertResource{}
	alertSchema := currentSchema(t, alertResource)
	invalidAlertResp := resource.ImportStateResponse{State: emptyTestState(alertSchema)}
	alertResource.ImportState(ctx, resource.ImportStateRequest{ID: ""}, &invalidAlertResp)
	if !invalidAlertResp.Diagnostics.HasError() {
		t.Fatal("organization spend alert import must reject an empty ID")
	}
	alertResp := resource.ImportStateResponse{State: emptyTestState(alertSchema)}
	alertResource.ImportState(ctx, resource.ImportStateRequest{ID: "alert_test"}, &alertResp)
	if alertResp.Diagnostics.HasError() {
		t.Fatalf("alert import diagnostics: %v", alertResp.Diagnostics)
	}
	var alertState OrganizationSpendAlertResourceModel
	if diags := alertResp.State.Get(ctx, &alertState); diags.HasError() {
		t.Fatalf("could not read alert import state: %v", diags)
	}
	if alertState.ID.ValueString() != "alert_test" || alertState.AlertID.ValueString() != "alert_test" {
		t.Fatalf("unexpected alert import state: %#v", alertState)
	}
}

func TestSpendGovernanceReadAndDelete404(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	resetProjectSettingsCacheForTest()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()
	client := newTestOpenAIClient(server.URL)
	ctx := context.Background()

	t.Run("project spend limit", func(t *testing.T) {
		r := &ProjectSpendLimitResource{client: client}
		schema := currentSchema(t, r)
		state := testStateFromModel(t, schema, &ProjectSpendLimitResourceModel{
			ID:              types.StringValue("proj_test"),
			ProjectID:       types.StringValue("proj_test"),
			ThresholdAmount: types.Int64Value(1),
			Currency:        types.StringValue("USD"),
			Interval:        types.StringValue("month"),
			Enforcement:     &SpendLimitEnforcementModel{Status: types.StringValue("enforcing")},
			Object:          types.StringValue("project.spend_limit"),
		})
		readResp := resource.ReadResponse{State: tfsdk.State{Schema: schema}}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		if readResp.Diagnostics.HasError() || !readResp.State.Raw.IsNull() {
			t.Fatalf("404 read should remove project spend limit: diagnostics=%v state=%v", readResp.Diagnostics, readResp.State.Raw)
		}
		deleteResp := resource.DeleteResponse{}
		r.Delete(ctx, resource.DeleteRequest{State: state}, &deleteResp)
		if deleteResp.Diagnostics.HasError() {
			t.Fatalf("404 delete diagnostics: %v", deleteResp.Diagnostics)
		}
	})

	t.Run("organization spend limit", func(t *testing.T) {
		r := &OrganizationSpendLimitResource{client: client}
		schema := currentSchema(t, r)
		state := testStateFromModel(t, schema, &OrganizationSpendLimitResourceModel{
			ID:              types.StringValue(organizationSpendLimitID),
			ThresholdAmount: types.Int64Value(1),
			Currency:        types.StringValue("USD"),
			Interval:        types.StringValue("month"),
			Enforcement:     &SpendLimitEnforcementModel{Status: types.StringValue("enforcing")},
			Object:          types.StringValue("organization.spend_limit"),
		})
		readResp := resource.ReadResponse{State: tfsdk.State{Schema: schema}}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		if readResp.Diagnostics.HasError() || !readResp.State.Raw.IsNull() {
			t.Fatalf("404 read should remove organization spend limit: diagnostics=%v state=%v", readResp.Diagnostics, readResp.State.Raw)
		}
		deleteResp := resource.DeleteResponse{}
		r.Delete(ctx, resource.DeleteRequest{State: state}, &deleteResp)
		if deleteResp.Diagnostics.HasError() {
			t.Fatalf("404 delete diagnostics: %v", deleteResp.Diagnostics)
		}
	})

	t.Run("organization spend alert", func(t *testing.T) {
		r := &OrganizationSpendAlertResource{client: client}
		schema := currentSchema(t, r)
		state := testStateFromModel(t, schema, &OrganizationSpendAlertResourceModel{
			ID:              types.StringValue("alert_test"),
			AlertID:         types.StringValue("alert_test"),
			ThresholdAmount: types.Int64Value(0),
			Currency:        types.StringValue("USD"),
			Interval:        types.StringValue("month"),
			NotificationChannel: &SpendAlertNotificationModel{
				Type:          types.StringValue("email"),
				Recipients:    modelIDsToSet([]string{"finance@example.com"}),
				SubjectPrefix: types.StringNull(),
			},
			Object: types.StringValue("organization.spend_alert"),
		})
		readResp := resource.ReadResponse{State: tfsdk.State{Schema: schema}}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		if readResp.Diagnostics.HasError() || !readResp.State.Raw.IsNull() {
			t.Fatalf("404 read should remove organization spend alert: diagnostics=%v state=%v", readResp.Diagnostics, readResp.State.Raw)
		}
		deleteResp := resource.DeleteResponse{}
		r.Delete(ctx, resource.DeleteRequest{State: state}, &deleteResp)
		if deleteResp.Diagnostics.HasError() {
			t.Fatalf("404 delete diagnostics: %v", deleteResp.Diagnostics)
		}
	})
}
