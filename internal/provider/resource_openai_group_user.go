package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GroupUserResource{}
var _ resource.ResourceWithImportState = &GroupUserResource{}

type GroupUserResource struct {
	client *OpenAIClient
}

func NewGroupUserResource() resource.Resource {
	return &GroupUserResource{}
}

func (r *GroupUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_user"
}

type GroupUserResourceModel struct {
	ID       types.String `tfsdk:"id"`
	GroupID  types.String `tfsdk:"group_id"`
	UserID   types.String `tfsdk:"user_id"`
	UserName types.String `tfsdk:"user_name"`
	Email    types.String `tfsdk:"email"`
	Role     types.String `tfsdk:"role"`
	AddedAt  types.Int64  `tfsdk:"added_at"`
}

func (r *GroupUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a user's membership in an OpenAI organization group.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the group user (group_id:user_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the user to add to the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the user.",
			},
			"email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The email of the user.",
			},
			"role": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user's organization role.",
			},
			"added_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Unix timestamp when the user was added to the organization.",
			},
		},
	}
}

func (r *GroupUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *provider.OpenAIClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *GroupUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := data.GroupID.ValueString()
	userID := data.UserID.ValueString()

	reqBody, err := json.Marshal(GroupUserCreateRequest{
		UserID: userID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	apiURL := r.client.OpenAIClient.APIURL
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups/" + groupID + "/users"
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups/" + groupID + "/users"
	}

	apiReq, err := http.NewRequest("POST", reqURL, bytes.NewReader(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiReq.Header.Set("Content-Type", "application/json")
	apiKey := r.client.OpenAIClient.APIKey
	if r.client.AdminAPIKey != "" {
		apiKey = r.client.AdminAPIKey
	}
	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	apiResp, err := httpClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK && apiResp.StatusCode != http.StatusCreated {
		respBodyBytes, readErr := io.ReadAll(apiResp.Body)
		if readErr != nil {
			resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s (could not read body)", apiResp.Status))
			return
		}
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}

	// The create response only returns group.user object with user_id and group_id
	// We need to fetch user details from the group users list
	data.ID = types.StringValue(fmt.Sprintf("%s:%s", groupID, userID))

	// Fetch user details
	if user := r.findUserInGroup(groupID, userID); user != nil {
		data.UserName = types.StringValue(user.Name)
		data.Email = types.StringValue(user.Email)
		data.Role = types.StringValue(user.Role)
		data.AddedAt = types.Int64Value(user.AddedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupUserResource) findUserInGroup(groupID, userID string) *OrganizationUserResponseFramework {
	apiURL := r.client.OpenAIClient.APIURL
	var baseURL string
	if strings.Contains(apiURL, "/v1") {
		baseURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups/" + groupID + "/users"
	} else {
		baseURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups/" + groupID + "/users"
	}

	apiKey := r.client.OpenAIClient.APIKey
	if r.client.AdminAPIKey != "" {
		apiKey = r.client.AdminAPIKey
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	cursor := ""
	for {
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return nil
		}
		q := parsedURL.Query()
		q.Set("limit", "100")
		if cursor != "" {
			q.Set("after", cursor)
		}
		parsedURL.RawQuery = q.Encode()

		apiReq, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			return nil
		}

		apiReq.Header.Set("Authorization", "Bearer "+apiKey)
		if r.client.OpenAIClient.OrganizationID != "" {
			apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
		}

		apiResp, err := httpClient.Do(apiReq)
		if err != nil {
			return nil
		}

		if apiResp.StatusCode != http.StatusOK {
			apiResp.Body.Close()
			return nil
		}

		var listResp GroupUserListResponse
		if err := json.NewDecoder(apiResp.Body).Decode(&listResp); err != nil {
			apiResp.Body.Close()
			return nil
		}
		apiResp.Body.Close()

		for i := range listResp.Data {
			if listResp.Data[i].ID == userID {
				return &listResp.Data[i]
			}
		}

		if !listResp.HasMore {
			break
		}
		if len(listResp.Data) > 0 {
			cursor = listResp.Data[len(listResp.Data)-1].ID
		} else {
			break
		}
	}
	return nil
}

func (r *GroupUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be group_id:user_id")
		return
	}
	groupID := idParts[0]
	userID := idParts[1]

	data.GroupID = types.StringValue(groupID)
	data.UserID = types.StringValue(userID)

	apiURL := r.client.OpenAIClient.APIURL
	var baseURL string
	if strings.Contains(apiURL, "/v1") {
		baseURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups/" + groupID + "/users"
	} else {
		baseURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups/" + groupID + "/users"
	}

	apiKey := r.client.OpenAIClient.APIKey
	if r.client.AdminAPIKey != "" {
		apiKey = r.client.AdminAPIKey
	}

	var foundUser *OrganizationUserResponseFramework
	cursor := ""
	httpClient := &http.Client{Timeout: 30 * time.Second}

	for foundUser == nil {
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			resp.Diagnostics.AddError("Error parsing URL", err.Error())
			return
		}
		q := parsedURL.Query()
		q.Set("limit", "100")
		if cursor != "" {
			q.Set("after", cursor)
		}
		parsedURL.RawQuery = q.Encode()

		apiReq, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			resp.Diagnostics.AddError("Error creating request", err.Error())
			return
		}

		apiReq.Header.Set("Authorization", "Bearer "+apiKey)
		if r.client.OpenAIClient.OrganizationID != "" {
			apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
		}

		apiResp, err := httpClient.Do(apiReq)
		if err != nil {
			resp.Diagnostics.AddError("Error making request", err.Error())
			return
		}

		if apiResp.StatusCode == http.StatusNotFound {
			apiResp.Body.Close()
			resp.State.RemoveResource(ctx)
			return
		}
		if apiResp.StatusCode != http.StatusOK {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s", apiResp.Status))
			return
		}

		var listResp GroupUserListResponse
		if err := json.NewDecoder(apiResp.Body).Decode(&listResp); err != nil {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("Error parsing response", err.Error())
			return
		}
		apiResp.Body.Close()

		for i := range listResp.Data {
			user := listResp.Data[i]
			if user.ID == userID {
				foundUser = &user
				break
			}
		}

		if foundUser != nil {
			break
		}

		if !listResp.HasMore {
			break
		}
		if len(listResp.Data) > 0 {
			cursor = listResp.Data[len(listResp.Data)-1].ID
		} else {
			break
		}
	}

	if foundUser == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.UserName = types.StringValue(foundUser.Name)
	data.Email = types.StringValue(foundUser.Email)
	data.Role = types.StringValue(foundUser.Role)
	data.AddedAt = types.Int64Value(foundUser.AddedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update endpoint - group membership is either present or not
	resp.Diagnostics.AddError(
		"Update not supported",
		"Group user membership cannot be updated. Remove and re-add the user to the group.",
	)
}

func (r *GroupUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "ID must be group_id:user_id")
		return
	}
	groupID := idParts[0]
	userID := idParts[1]

	apiURL := r.client.OpenAIClient.APIURL
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups/" + groupID + "/users/" + userID
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups/" + groupID + "/users/" + userID
	}

	apiReq, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiKey := r.client.OpenAIClient.APIKey
	if r.client.AdminAPIKey != "" {
		apiKey = r.client.AdminAPIKey
	}
	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	apiResp, err := httpClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting group user", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK && apiResp.StatusCode != http.StatusNoContent && apiResp.StatusCode != http.StatusNotFound {
		respBodyBytes, readErr := io.ReadAll(apiResp.Body)
		if readErr != nil {
			resp.Diagnostics.AddError("Error deleting group user", fmt.Sprintf("API returned error: %s (could not read body)", apiResp.Status))
			return
		}
		resp.Diagnostics.AddError("Error deleting group user", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
	}
}

func (r *GroupUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
