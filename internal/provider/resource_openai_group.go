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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GroupResource{}
var _ resource.ResourceWithImportState = &GroupResource{}

type GroupResource struct {
	client *OpenAIClient
}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

type GroupResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	IsSCIMManaged types.Bool   `tfsdk:"is_scim_managed"`
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI organization group. Groups are collections of users that can be assigned roles at the organization or project level.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the group.",
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Unix timestamp (in seconds) when the group was created.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_scim_managed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the group is managed through SCIM and controlled by your identity provider.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody, err := json.Marshal(GroupCreateRequest{
		Name: data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	apiURL := r.client.OpenAIClient.APIURL
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups"
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups"
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

	var groupResp GroupResponseFramework
	respBodyBytes, err := io.ReadAll(apiResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Error reading response body", err.Error())
		return
	}
	if err := json.Unmarshal(respBodyBytes, &groupResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(groupResp.ID)
	data.Name = types.StringValue(groupResp.Name)
	data.CreatedAt = types.Int64Value(groupResp.CreatedAt)
	data.IsSCIMManaged = types.BoolValue(groupResp.IsSCIMManaged)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := data.ID.ValueString()

	// No single-group GET endpoint exists, so we must list and filter
	apiURL := r.client.OpenAIClient.APIURL
	var baseURL string
	if strings.Contains(apiURL, "/v1") {
		baseURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups"
	} else {
		baseURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups"
	}

	apiKey := r.client.OpenAIClient.APIKey
	if r.client.AdminAPIKey != "" {
		apiKey = r.client.AdminAPIKey
	}

	var foundGroup *GroupResponseFramework
	cursor := ""
	httpClient := &http.Client{Timeout: 30 * time.Second}

	for foundGroup == nil {
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

		if apiResp.StatusCode != http.StatusOK {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s", apiResp.Status))
			return
		}

		var listResp GroupListResponse
		if err := json.NewDecoder(apiResp.Body).Decode(&listResp); err != nil {
			apiResp.Body.Close()
			resp.Diagnostics.AddError("Error parsing response", err.Error())
			return
		}
		apiResp.Body.Close()

		for i := range listResp.Data {
			group := listResp.Data[i]
			if group.ID == groupID {
				foundGroup = &group
				break
			}
		}

		if foundGroup != nil {
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

	if foundGroup == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(foundGroup.ID)
	data.Name = types.StringValue(foundGroup.Name)
	data.CreatedAt = types.Int64Value(foundGroup.CreatedAt)
	data.IsSCIMManaged = types.BoolValue(foundGroup.IsSCIMManaged)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := data.ID.ValueString()

	reqBody, err := json.Marshal(GroupUpdateRequest{
		Name: data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling request", err.Error())
		return
	}

	apiURL := r.client.OpenAIClient.APIURL
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups/" + groupID
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups/" + groupID
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

	if apiResp.StatusCode != http.StatusOK {
		respBodyBytes, readErr := io.ReadAll(apiResp.Body)
		if readErr != nil {
			resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s (could not read body)", apiResp.Status))
			return
		}
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}

	var groupResp GroupResponseFramework
	respBodyBytes, err := io.ReadAll(apiResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Error reading response body", err.Error())
		return
	}
	if err := json.Unmarshal(respBodyBytes, &groupResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(groupResp.ID)
	data.Name = types.StringValue(groupResp.Name)
	data.CreatedAt = types.Int64Value(groupResp.CreatedAt)
	data.IsSCIMManaged = types.BoolValue(groupResp.IsSCIMManaged)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := data.ID.ValueString()

	apiURL := r.client.OpenAIClient.APIURL
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups/" + groupID
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups/" + groupID
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
		resp.Diagnostics.AddError("Error deleting group", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK && apiResp.StatusCode != http.StatusNoContent && apiResp.StatusCode != http.StatusNotFound {
		respBodyBytes, readErr := io.ReadAll(apiResp.Body)
		if readErr != nil {
			resp.Diagnostics.AddError("Error deleting group", fmt.Sprintf("API returned error: %s (could not read body)", apiResp.Status))
			return
		}
		resp.Diagnostics.AddError("Error deleting group", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
