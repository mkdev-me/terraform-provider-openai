package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &OrganizationUserResource{}
var _ resource.ResourceWithImportState = &OrganizationUserResource{}

type OrganizationUserResource struct {
	client *OpenAIClient
}

func NewOrganizationUserResource() resource.Resource {
	return &OrganizationUserResource{}
}

func (r *OrganizationUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_user"
}

type OrganizationUserResourceModel struct {
	UserID types.String `tfsdk:"user_id"`
	// ID mapping to UserID for terraform state
	ID      types.String `tfsdk:"id"`
	Role    types.String `tfsdk:"role"`
	Email   types.String `tfsdk:"email"`
	Name    types.String `tfsdk:"name"`
	AddedAt types.Int64  `tfsdk:"added_at"`
}

func (r *OrganizationUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a user in an OpenAI Organization.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the user (same as user_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The role of the user in the organization (owner or reader).",
			},
			"email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The email of the user.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the user.",
			},
			"added_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the user was added to the organization.",
			},
		},
	}
}

func (r *OrganizationUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if user exists first. API doesn't support "creating" an org user directly via this endpoint usually,
	// they are invited. This resource manages existing users.
	// Equivalent of Read + Update if needed.
	// Logic from SDKv2: "Since users cannot be created through the API, this function verifies the user exists and updates their role if necessary."

	userID := data.UserID.ValueString()
	data.ID = types.StringValue(userID)

	// Read user
	url := fmt.Sprintf("%s/organization/users/%s", r.client.OpenAIClient.APIURL, userID)
	apiReq, err := http.NewRequest("GET", url, nil)
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

	apiResp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode == http.StatusNotFound {
		resp.Diagnostics.AddError("User not found", fmt.Sprintf("User %s does not exist in the organization", userID))
		return
	}

	if apiResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned status: %s", apiResp.Status))
		return
	}

	var userResp OrganizationUserResponseFramework
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	json.Unmarshal(respBodyBytes, &userResp)

	// Check role, update if needed
	if userResp.Role != data.Role.ValueString() {
		// Update role
		reqMap := map[string]string{"role": data.Role.ValueString()}
		reqBytes, _ := json.Marshal(reqMap)
		url := fmt.Sprintf("%s/organization/users/%s", r.client.OpenAIClient.APIURL, userID)
		apiUpdateReq, _ := http.NewRequest("POST", url, bytes.NewReader(reqBytes))
		apiUpdateReq.Header.Set("Content-Type", "application/json")
		apiUpdateReq.Header.Set("Authorization", "Bearer "+apiKey)
		if r.client.OpenAIClient.OrganizationID != "" {
			apiUpdateReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
		}

		upResp, err := http.DefaultClient.Do(apiUpdateReq)
		if err != nil || upResp.StatusCode != http.StatusOK {
			resp.Diagnostics.AddError("Error updating role", "Failed to update user role")
			return
		}
		// Read again? Or rely on explicit set.
		userResp.Role = data.Role.ValueString()
	}

	data.Name = types.StringValue(userResp.Name)
	data.Email = types.StringValue(userResp.Email)
	data.AddedAt = types.Int64Value(userResp.AddedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/organization/users/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	apiReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
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

	apiResp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	var userResp OrganizationUserResponseFramework
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	json.Unmarshal(respBodyBytes, &userResp)

	data.UserID = types.StringValue(userResp.ID)
	data.Name = types.StringValue(userResp.Name)
	data.Email = types.StringValue(userResp.Email)
	data.Role = types.StringValue(userResp.Role)
	data.AddedAt = types.Int64Value(userResp.AddedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrganizationUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update role
	reqMap := map[string]string{"role": data.Role.ValueString()}
	reqBytes, _ := json.Marshal(reqMap)
	url := fmt.Sprintf("%s/organization/users/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	apiUpdateReq, _ := http.NewRequest("POST", url, bytes.NewReader(reqBytes))
	apiUpdateReq.Header.Set("Content-Type", "application/json")

	apiKey := r.client.OpenAIClient.APIKey
	if r.client.AdminAPIKey != "" {
		apiKey = r.client.AdminAPIKey
	}
	apiUpdateReq.Header.Set("Authorization", "Bearer "+apiKey)

	if r.client.OpenAIClient.OrganizationID != "" {
		apiUpdateReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	upResp, err := http.DefaultClient.Do(apiUpdateReq)
	if err != nil || upResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Error updating role", "Failed to update user role")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/organization/users/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	apiReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
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

	http.DefaultClient.Do(apiReq)
}

func (r *OrganizationUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
