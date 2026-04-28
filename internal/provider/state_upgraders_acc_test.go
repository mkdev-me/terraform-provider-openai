package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// End-to-end acceptance test of the v0→v1 state migration.
//
// Step 1 installs the released v2.0.0 provider from the Terraform Registry and
// applies a config using the old `role` attribute. This produces a real state
// file written under the v0 schema, exactly as it exists in any user repo
// stuck on 2.0.0.
//
// Step 2 swaps in the current branch's build of the provider and a config
// using the new `role_ids` attribute. Terraform detects the prior schema
// version is older than the resource's current version and invokes our
// UpgradeState handler. If the upgrader works, the post-apply plan is empty
// (the SDKv2 test framework asserts this implicitly). If it does not, the
// step fails with either a state-decode error or a non-empty plan diagnostic.
//
// The OpenAI admin API is faked by an in-test HTTP server so the test runs
// fully offline.
//
// Set TF_ACC=1 to run; required because ExternalProviders downloads the v2.0.0
// release from the registry.

const (
	// Stable role IDs used by the mock — emulate predefined OpenAI roles.
	mockRoleMemberID = "role_predef_member"
	mockRoleOwnerID  = "role_predef_owner"

	mockTestProjectID = "proj_acc_test"
	mockTestUserID    = "user_acc_test"
	mockTestUserEmail = "acc-test@example.com"
)

func TestAccStateMigration_ProjectUser_V0ToV1(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	srv := newMockOpenAIServer(t)
	defer srv.Close()

	t.Setenv("OPENAI_API_URL", srv.URL+"/v1")
	t.Setenv("OPENAI_ADMIN_KEY", "acc-test-admin-key")
	t.Setenv("OPENAI_API_KEY", "acc-test-api-key")

	resetRoleCacheForTest()

	resource.Test(t, resource.TestCase{
		IsUnitTest: false,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"openai": {
						Source:            "mkdev-me/openai",
						VersionConstraint: "= 2.0.0",
					},
				},
				Config: testAccConfigProjectUserV0(srv.URL),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccConfigProjectUserV1(srv.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("openai_project_user.test", "project_id", mockTestProjectID),
					resource.TestCheckResourceAttr("openai_project_user.test", "user_id", mockTestUserID),
					resource.TestCheckResourceAttr("openai_project_user.test", "role_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("openai_project_user.test", "role_ids.*", mockRoleMemberID),
				),
			},
		},
	})
}

func testAccConfigProjectUserV0(apiURL string) string {
	return fmt.Sprintf(`
provider "openai" {
  api_url   = "%s/v1"
  admin_key = "acc-test-admin-key"
  api_key   = "acc-test-api-key"
}

resource "openai_project_user" "test" {
  project_id = %q
  user_id    = %q
  role       = "member"
}
`, apiURL, mockTestProjectID, mockTestUserID)
}

func testAccConfigProjectUserV1(apiURL string) string {
	return fmt.Sprintf(`
provider "openai" {
  api_url   = "%s/v1"
  admin_key = "acc-test-admin-key"
  api_key   = "acc-test-api-key"
}

data "openai_project_role" "member" {
  project_id = %q
  name       = "member"
}

resource "openai_project_user" "test" {
  project_id = %q
  user_id    = %q
  role_ids   = [data.openai_project_role.member.id]
}
`, apiURL, mockTestProjectID, mockTestProjectID, mockTestUserID)
}

// ----- Mock OpenAI admin API -----

// mockOpenAIServer is an in-process fake of the subset of the OpenAI admin
// API exercised by openai_project_user and openai_project_group across both
// the v2.0.0 and v2.1.0+ schemas.
//
// State is kept in-memory and survives across requests so that a resource
// created in step 1 (with v2.0.0) is still found by reads issued in step 2
// (after the upgrader runs).
type mockOpenAIServer struct {
	*httptest.Server
	t       *testing.T
	mu      sync.Mutex
	users   map[string]*mockUser  // key: project_id|user_id
	groups  map[string]*mockGroup // key: project_id|group_id
}

type mockUser struct {
	ID      string
	Email   string
	Roles   map[string]bool // set of role_ids assigned via the per-user roles endpoint
	AddedAt int64
}

type mockGroup struct {
	ID        string
	Name      string
	Roles     map[string]bool
	CreatedAt int64
}

func newMockOpenAIServer(t *testing.T) *mockOpenAIServer {
	t.Helper()
	srv := &mockOpenAIServer{
		t:      t,
		users:  map[string]*mockUser{},
		groups: map[string]*mockGroup{},
	}
	srv.Server = httptest.NewServer(http.HandlerFunc(srv.handle))
	return srv
}

var (
	reUsers       = regexp.MustCompile(`^/v1/organization/projects/([^/]+)/users/?$`)
	reUserByID    = regexp.MustCompile(`^/v1/organization/projects/([^/]+)/users/([^/]+)/?$`)
	reUserRoles   = regexp.MustCompile(`^/v1/projects/([^/]+)/users/([^/]+)/roles/?$`)
	reUserRoleID  = regexp.MustCompile(`^/v1/projects/([^/]+)/users/([^/]+)/roles/([^/]+)/?$`)
	reGroups      = regexp.MustCompile(`^/v1/organization/projects/([^/]+)/groups/?$`)
	reGroupByID   = regexp.MustCompile(`^/v1/organization/projects/([^/]+)/groups/([^/]+)/?$`)
	reGroupRoles  = regexp.MustCompile(`^/v1/projects/([^/]+)/groups/([^/]+)/roles/?$`)
	reGroupRoleID = regexp.MustCompile(`^/v1/projects/([^/]+)/groups/([^/]+)/roles/([^/]+)/?$`)
	reProjRoles   = regexp.MustCompile(`^/v1/projects/([^/]+)/roles/?$`)
)

func (s *mockOpenAIServer) handle(w http.ResponseWriter, r *http.Request) {
	if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Bearer ") {
		http.Error(w, "missing auth", http.StatusUnauthorized)
		return
	}

	switch {
	case reProjRoles.MatchString(r.URL.Path) && r.Method == "GET":
		s.handleListProjectRoles(w)
		return

	case reUserByID.MatchString(r.URL.Path):
		m := reUserByID.FindStringSubmatch(r.URL.Path)
		s.handleUserByID(w, r, m[1], m[2])
		return

	case reUsers.MatchString(r.URL.Path):
		m := reUsers.FindStringSubmatch(r.URL.Path)
		s.handleUsers(w, r, m[1])
		return

	case reUserRoleID.MatchString(r.URL.Path):
		m := reUserRoleID.FindStringSubmatch(r.URL.Path)
		s.handleUserRoleID(w, r, m[1], m[2], m[3])
		return

	case reUserRoles.MatchString(r.URL.Path):
		m := reUserRoles.FindStringSubmatch(r.URL.Path)
		s.handleUserRoles(w, r, m[1], m[2])
		return

	case reGroupByID.MatchString(r.URL.Path):
		m := reGroupByID.FindStringSubmatch(r.URL.Path)
		s.handleGroupByID(w, r, m[1], m[2])
		return

	case reGroups.MatchString(r.URL.Path):
		m := reGroups.FindStringSubmatch(r.URL.Path)
		s.handleGroups(w, r, m[1])
		return

	case reGroupRoleID.MatchString(r.URL.Path):
		m := reGroupRoleID.FindStringSubmatch(r.URL.Path)
		s.handleGroupRoleID(w, r, m[1], m[2], m[3])
		return

	case reGroupRoles.MatchString(r.URL.Path):
		m := reGroupRoles.FindStringSubmatch(r.URL.Path)
		s.handleGroupRoles(w, r, m[1], m[2])
		return
	}

	http.Error(w, "not found: "+r.Method+" "+r.URL.Path, http.StatusNotFound)
}

func (s *mockOpenAIServer) handleListProjectRoles(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"object": "list",
		"data": []map[string]interface{}{
			{"id": mockRoleMemberID, "name": "member", "predefined_role": true, "permissions": []string{}, "resource_type": "api.project"},
			{"id": mockRoleOwnerID, "name": "owner", "predefined_role": true, "permissions": []string{}, "resource_type": "api.project"},
		},
		"has_more": false,
	})
}

func (s *mockOpenAIServer) userKey(projectID, userID string) string { return projectID + "|" + userID }
func (s *mockOpenAIServer) groupKey(projectID, groupID string) string {
	return projectID + "|" + groupID
}

func (s *mockOpenAIServer) handleUsers(w http.ResponseWriter, r *http.Request, projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch r.Method {
	case "POST":
		var body struct {
			UserID string `json:"user_id"`
			Role   string `json:"role"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		key := s.userKey(projectID, body.UserID)
		u, ok := s.users[key]
		if !ok {
			u = &mockUser{
				ID:      body.UserID,
				Email:   mockTestUserEmail,
				Roles:   map[string]bool{},
				AddedAt: 1700000000,
			}
			s.users[key] = u
		}
		// Bridge: v2.0.0 sets the role via this endpoint with a name string.
		// v2.1.0+ ignores it here and uses the per-user roles endpoint instead
		// (always with a role *id*). Map known names to canonical IDs so a v0
		// create produces state the v1 read can return consistently.
		if id := canonicalRoleID(body.Role); id != "" {
			u.Roles[id] = true
		}
		writeJSON(w, http.StatusOK, userJSON(u, projectID))
		return

	case "GET":
		out := []map[string]interface{}{}
		for k, u := range s.users {
			if strings.HasPrefix(k, projectID+"|") {
				out = append(out, userJSON(u, projectID))
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"object": "list", "data": out, "has_more": false})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *mockOpenAIServer) handleUserByID(w http.ResponseWriter, r *http.Request, projectID, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.userKey(projectID, userID)
	u, ok := s.users[key]

	switch r.Method {
	case "GET":
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, userJSON(u, projectID))
		return
	case "POST":
		// v2.0.0 update endpoint (changes role)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var body struct {
			Role string `json:"role"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if id := canonicalRoleID(body.Role); id != "" {
			u.Roles = map[string]bool{id: true}
		}
		writeJSON(w, http.StatusOK, userJSON(u, projectID))
		return
	case "DELETE":
		delete(s.users, key)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *mockOpenAIServer) handleUserRoles(w http.ResponseWriter, r *http.Request, projectID, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u := s.users[s.userKey(projectID, userID)]
	if u == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		data := make([]map[string]interface{}, 0, len(u.Roles))
		for id := range u.Roles {
			data = append(data, map[string]interface{}{"id": id, "name": roleNameFromID(id)})
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"object": "list", "data": data, "has_more": false})
		return
	case "POST":
		var body struct {
			RoleID string `json:"role_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		u.Roles[body.RoleID] = true
		writeJSON(w, http.StatusOK, map[string]interface{}{"id": body.RoleID})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *mockOpenAIServer) handleUserRoleID(w http.ResponseWriter, r *http.Request, projectID, userID, roleID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if r.Method != "DELETE" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if u := s.users[s.userKey(projectID, userID)]; u != nil {
		delete(u.Roles, roleID)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *mockOpenAIServer) handleGroups(w http.ResponseWriter, r *http.Request, projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch r.Method {
	case "POST":
		var body struct {
			GroupID string `json:"group_id"`
			Role    string `json:"role"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		key := s.groupKey(projectID, body.GroupID)
		g, ok := s.groups[key]
		if !ok {
			g = &mockGroup{
				ID:        body.GroupID,
				Name:      "Test Group",
				Roles:     map[string]bool{},
				CreatedAt: 1700000000,
			}
			s.groups[key] = g
		}
		// v2.0.0 used the role field for an ID directly; v2.1+ uses the
		// per-group roles endpoint. Either way, store it.
		if body.Role != "" {
			g.Roles[body.Role] = true
		}
		writeJSON(w, http.StatusOK, groupJSON(g))
		return
	case "GET":
		out := []map[string]interface{}{}
		for k, g := range s.groups {
			if strings.HasPrefix(k, projectID+"|") {
				out = append(out, groupJSON(g))
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"object": "list", "data": out, "has_more": false})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *mockOpenAIServer) handleGroupByID(w http.ResponseWriter, r *http.Request, projectID, groupID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch r.Method {
	case "GET":
		g := s.groups[s.groupKey(projectID, groupID)]
		if g == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, groupJSON(g))
	case "DELETE":
		delete(s.groups, s.groupKey(projectID, groupID))
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *mockOpenAIServer) handleGroupRoles(w http.ResponseWriter, r *http.Request, projectID, groupID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g := s.groups[s.groupKey(projectID, groupID)]
	if g == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	switch r.Method {
	case "GET":
		data := make([]map[string]interface{}, 0, len(g.Roles))
		for id := range g.Roles {
			data = append(data, map[string]interface{}{"id": id, "name": roleNameFromID(id)})
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"object": "list", "data": data, "has_more": false})
	case "POST":
		var body struct {
			RoleID string `json:"role_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		g.Roles[body.RoleID] = true
		writeJSON(w, http.StatusOK, map[string]interface{}{"id": body.RoleID})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *mockOpenAIServer) handleGroupRoleID(w http.ResponseWriter, r *http.Request, projectID, groupID, roleID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if r.Method != "DELETE" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if g := s.groups[s.groupKey(projectID, groupID)]; g != nil {
		delete(g.Roles, roleID)
	}
	w.WriteHeader(http.StatusNoContent)
}

func userJSON(u *mockUser, projectID string) map[string]interface{} {
	role := "member"
	if u.Roles[mockRoleOwnerID] {
		role = "owner"
	}
	return map[string]interface{}{
		"object":   "organization.project.user",
		"id":       u.ID,
		"name":     "Test User",
		"email":    u.Email,
		"role":     role,
		"added_at": u.AddedAt,
	}
}

func groupJSON(g *mockGroup) map[string]interface{} {
	return map[string]interface{}{
		"object":     "organization.project.group",
		"group_id":   g.ID,
		"group_name": g.Name,
		"created_at": g.CreatedAt,
	}
}

func canonicalRoleID(roleName string) string {
	switch strings.ToLower(roleName) {
	case "member":
		return mockRoleMemberID
	case "owner":
		return mockRoleOwnerID
	default:
		return ""
	}
}

func roleNameFromID(id string) string {
	switch id {
	case mockRoleMemberID:
		return "member"
	case mockRoleOwnerID:
		return "owner"
	default:
		return "unknown"
	}
}

func writeJSON(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
