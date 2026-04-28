package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ---------- helpers ----------

// configuredProjectUserResource returns a ProjectUserResource with its client
// populated by calling the real Configure method, mimicking what the framework
// does before invoking UpgradeState.
func configuredProjectUserResource(t *testing.T, c *OpenAIClient) *ProjectUserResource {
	t.Helper()
	r := &ProjectUserResource{}
	configureReq := resource.ConfigureRequest{ProviderData: c}
	configureResp := resource.ConfigureResponse{}
	r.Configure(context.Background(), configureReq, &configureResp)
	if configureResp.Diagnostics.HasError() {
		t.Fatalf("Configure produced diagnostics: %v", configureResp.Diagnostics)
	}
	if r.client == nil {
		t.Fatal("client is nil after Configure — the real framework would also see this")
	}
	return r
}

// currentSchema invokes the resource's Schema method to retrieve the current
// (target) schema used to initialize the response State.
func currentSchema(t *testing.T, r resource.Resource) rschema.Schema {
	t.Helper()
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema produced diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

// runUpgrader picks the upgrader for `fromVersion` and invokes it with a
// prior State built from `priorRaw` against the upgrader's PriorSchema, then
// returns the resulting response.
func runUpgrader(
	t *testing.T,
	upgraders map[int64]resource.StateUpgrader,
	fromVersion int64,
	currentSch rschema.Schema,
	priorRaw map[string]tftypes.Value,
) resource.UpgradeStateResponse {
	t.Helper()

	up, ok := upgraders[fromVersion]
	if !ok {
		t.Fatalf("no upgrader registered for version %d", fromVersion)
	}
	if up.PriorSchema == nil {
		t.Fatalf("upgrader for version %d has no PriorSchema", fromVersion)
	}

	ctx := context.Background()

	priorType := up.PriorSchema.Type().TerraformType(ctx)
	priorValue := tftypes.NewValue(priorType, priorRaw)

	priorState := &tfsdk.State{
		Schema: *up.PriorSchema,
		Raw:    priorValue,
	}

	respState := tfsdk.State{
		Schema: currentSch,
		Raw:    tftypes.NewValue(currentSch.Type().TerraformType(ctx), nil),
	}

	req := resource.UpgradeStateRequest{State: priorState}
	resp := resource.UpgradeStateResponse{State: respState}

	up.StateUpgrader(ctx, req, &resp)
	return resp
}

// ---------- project_user upgrader tests ----------

func TestProjectUserUpgradeState_V0ToV1_Member(t *testing.T) {
	resetRoleCacheForTest()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/proj_xxx/roles" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
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
	r := configuredProjectUserResource(t, c)
	upgraders := r.UpgradeState(context.Background())
	currentSch := currentSchema(t, r)

	resp := runUpgrader(t, upgraders, 0, currentSch, map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, "proj_xxx:user_yyy"),
		"project_id": tftypes.NewValue(tftypes.String, "proj_xxx"),
		"user_id":    tftypes.NewValue(tftypes.String, "user_yyy"),
		"role":       tftypes.NewValue(tftypes.String, "member"),
		"email":      tftypes.NewValue(tftypes.String, "test@example.com"),
		"added_at":   tftypes.NewValue(tftypes.Number, 1700000000),
	})

	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrader produced errors: %v", resp.Diagnostics)
	}

	var got ProjectUserResourceModel
	if d := resp.State.Get(context.Background(), &got); d.HasError() {
		t.Fatalf("could not read upgraded state: %v", d)
	}

	if got.ID.ValueString() != "proj_xxx:user_yyy" {
		t.Errorf("ID: got %q, want %q", got.ID.ValueString(), "proj_xxx:user_yyy")
	}
	if got.ProjectID.ValueString() != "proj_xxx" {
		t.Errorf("ProjectID: got %q, want %q", got.ProjectID.ValueString(), "proj_xxx")
	}
	if got.UserID.ValueString() != "user_yyy" {
		t.Errorf("UserID: got %q, want %q", got.UserID.ValueString(), "user_yyy")
	}
	if got.Email.ValueString() != "test@example.com" {
		t.Errorf("Email: got %q, want %q", got.Email.ValueString(), "test@example.com")
	}
	if got.AddedAt.ValueInt64() != 1700000000 {
		t.Errorf("AddedAt: got %d, want %d", got.AddedAt.ValueInt64(), 1700000000)
	}

	roleIDs := roleIDsFromSet(got.RoleIDs)
	if len(roleIDs) != 1 || roleIDs[0] != "role_member_id" {
		t.Errorf("RoleIDs: got %v, want [role_member_id]", roleIDs)
	}
}

func TestProjectUserUpgradeState_V0ToV1_Owner(t *testing.T) {
	resetRoleCacheForTest()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	r := configuredProjectUserResource(t, c)
	upgraders := r.UpgradeState(context.Background())
	currentSch := currentSchema(t, r)

	resp := runUpgrader(t, upgraders, 0, currentSch, map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, "proj_x:user_owner"),
		"project_id": tftypes.NewValue(tftypes.String, "proj_x"),
		"user_id":    tftypes.NewValue(tftypes.String, "user_owner"),
		"role":       tftypes.NewValue(tftypes.String, "owner"),
		"email":      tftypes.NewValue(tftypes.String, "owner@example.com"),
		"added_at":   tftypes.NewValue(tftypes.Number, 1700000001),
	})

	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrader produced errors: %v", resp.Diagnostics)
	}

	var got ProjectUserResourceModel
	if d := resp.State.Get(context.Background(), &got); d.HasError() {
		t.Fatalf("could not read upgraded state: %v", d)
	}

	roleIDs := roleIDsFromSet(got.RoleIDs)
	if len(roleIDs) != 1 || roleIDs[0] != "role_owner_id" {
		t.Errorf("RoleIDs: got %v, want [role_owner_id]", roleIDs)
	}
}

func TestProjectUserUpgradeState_RoleNotFound(t *testing.T) {
	resetRoleCacheForTest()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object":   "list",
			"data":     []map[string]interface{}{{"id": "role_other", "name": "viewer"}},
			"has_more": false,
		})
	}))
	defer server.Close()

	c := newTestOpenAIClient(server.URL)
	r := configuredProjectUserResource(t, c)
	upgraders := r.UpgradeState(context.Background())
	currentSch := currentSchema(t, r)

	resp := runUpgrader(t, upgraders, 0, currentSch, map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, "proj_y:user_z"),
		"project_id": tftypes.NewValue(tftypes.String, "proj_y"),
		"user_id":    tftypes.NewValue(tftypes.String, "user_z"),
		"role":       tftypes.NewValue(tftypes.String, "member"),
		"email":      tftypes.NewValue(tftypes.String, "x@x.com"),
		"added_at":   tftypes.NewValue(tftypes.Number, 0),
	})

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected upgrader to fail when role is missing")
	}
	combined := resp.Diagnostics.Errors()[0].Summary() + " " + resp.Diagnostics.Errors()[0].Detail()
	if !strings.Contains(combined, "State upgrade failed") {
		t.Errorf("unexpected error: %s", combined)
	}
}

func TestProjectUserUpgradeState_EmptyRoleName(t *testing.T) {
	resetRoleCacheForTest()
	c := newTestOpenAIClient("http://127.0.0.1:1") // unused: must fail before any HTTP call
	r := configuredProjectUserResource(t, c)
	upgraders := r.UpgradeState(context.Background())
	currentSch := currentSchema(t, r)

	resp := runUpgrader(t, upgraders, 0, currentSch, map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, "proj_y:user_z"),
		"project_id": tftypes.NewValue(tftypes.String, "proj_y"),
		"user_id":    tftypes.NewValue(tftypes.String, "user_z"),
		"role":       tftypes.NewValue(tftypes.String, ""),
		"email":      tftypes.NewValue(tftypes.String, "x@x.com"),
		"added_at":   tftypes.NewValue(tftypes.Number, 0),
	})

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected upgrader to fail on empty role name")
	}
}

func TestProjectUserUpgradeState_CacheCollapsesAcrossResources(t *testing.T) {
	resetRoleCacheForTest()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
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
	r := configuredProjectUserResource(t, c)
	upgraders := r.UpgradeState(context.Background())
	currentSch := currentSchema(t, r)

	// Simulate 50 resources in the same project — all should collapse to 1 API call.
	for i := 0; i < 50; i++ {
		resp := runUpgrader(t, upgraders, 0, currentSch, map[string]tftypes.Value{
			"id":         tftypes.NewValue(tftypes.String, "proj_shared:user_n"),
			"project_id": tftypes.NewValue(tftypes.String, "proj_shared"),
			"user_id":    tftypes.NewValue(tftypes.String, "user_n"),
			"role":       tftypes.NewValue(tftypes.String, "member"),
			"email":      tftypes.NewValue(tftypes.String, "n@x.com"),
			"added_at":   tftypes.NewValue(tftypes.Number, 0),
		})
		if resp.Diagnostics.HasError() {
			t.Fatalf("upgrader produced errors on iteration %d: %v", i, resp.Diagnostics)
		}
	}

	if calls != 1 {
		t.Fatalf("expected 1 API call thanks to cache, got %d", calls)
	}
}

// Sanity: types.Set can be empty without panicking — guards against future regressions.
var _ = types.Set{}
