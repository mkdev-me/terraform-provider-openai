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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &InviteResource{}
var _ resource.ResourceWithImportState = &InviteResource{}

type InviteResource struct {
	client *OpenAIClient
}

func NewInviteResource() resource.Resource {
	return &InviteResource{}
}

func (r *InviteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invite"
}

type InviteResourceModel struct {
	ID        types.String         `tfsdk:"id"`
	Email     types.String         `tfsdk:"email"`
	Role      types.String         `tfsdk:"role"`
	Projects  []InviteProjectModel `tfsdk:"projects"`
	InviteID  types.String         `tfsdk:"invite_id"`
	Status    types.String         `tfsdk:"status"`
	CreatedAt types.Int64          `tfsdk:"created_at"`
	ExpiresAt types.Int64          `tfsdk:"expires_at"`
}

type InviteProjectModel struct {
	ID   types.String `tfsdk:"id"`
	Role types.String `tfsdk:"role"`
}

func (r *InviteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI User Invitation.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the invitation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The email address of the user to invite.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The role to assign to the user (owner or reader).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expires_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "When the invitation expires.",
			},
			"invite_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the invitation.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the invitation.",
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "When the invitation was created.",
			},
		},

		Blocks: map[string]schema.Block{
			"projects": schema.ListNestedBlock{
				Description: "The projects to invite the user to.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the project",
						},
						"role": schema.StringAttribute{
							Required:    true,
							Description: "The role to assign to the user within the project (owner or member)",
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *InviteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InviteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InviteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := InviteCreateRequest{
		Email: data.Email.ValueString(),
		Role:  data.Role.ValueString(),
	}

	if len(data.Projects) > 0 {
		projects := []InviteProject{}
		for _, p := range data.Projects {
			projects = append(projects, InviteProject{
				ID:   p.ID.ValueString(),
				Role: p.Role.ValueString(),
			})
		}
		createRequest.Projects = projects
	}

	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing request", err.Error())
		return
	}

	url := fmt.Sprintf("%s/organization/invites", r.client.OpenAIClient.APIURL)
	apiReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiReq.Header.Set("Content-Type", "application/json")

	// Use Admin Key if available
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

	if apiResp.StatusCode != http.StatusOK && apiResp.StatusCode != http.StatusCreated {
		respBodyBytes, _ := io.ReadAll(apiResp.Body)
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}

	var inviteResp InviteResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &inviteResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(inviteResp.ID)
	data.InviteID = types.StringValue(inviteResp.ID)
	data.Status = types.StringValue(inviteResp.Status)
	data.CreatedAt = types.Int64Value(inviteResp.CreatedAt)
	data.ExpiresAt = types.Int64Value(inviteResp.ExpiresAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InviteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InviteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/organization/invites/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
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
		resp.State.RemoveResource(ctx)
		return
	}
	if apiResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s", apiResp.Status))
		return
	}

	var inviteResp InviteResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &inviteResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.Email = types.StringValue(inviteResp.Email)
	data.Role = types.StringValue(inviteResp.Role)
	data.Status = types.StringValue(inviteResp.Status)
	data.CreatedAt = types.Int64Value(inviteResp.CreatedAt)
	data.ExpiresAt = types.Int64Value(inviteResp.ExpiresAt)
	data.InviteID = types.StringValue(inviteResp.ID)

	if len(inviteResp.Projects) > 0 {
		projects := []InviteProjectModel{}
		for _, p := range inviteResp.Projects {
			projects = append(projects, InviteProjectModel{
				ID:   types.StringValue(p.ID),
				Role: types.StringValue(p.Role),
			})
		}
		data.Projects = projects
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InviteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Immutable
}

func (r *InviteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InviteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/organization/invites/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
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

func (r *InviteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
