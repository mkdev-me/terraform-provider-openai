package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectServiceAccountResource{}
var _ resource.ResourceWithImportState = &ProjectServiceAccountResource{}

type ProjectServiceAccountResource struct {
	client *OpenAIClient
}

func NewProjectServiceAccountResource() resource.Resource {
	return &ProjectServiceAccountResource{}
}

func (r *ProjectServiceAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_service_account"
}

type ProjectServiceAccountResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ProjectID        types.String `tfsdk:"project_id"`
	Name             types.String `tfsdk:"name"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	CreatedAt        types.Int64  `tfsdk:"created_at"`
	Role             types.String `tfsdk:"role"`
	APIKeyID         types.String `tfsdk:"api_key_id"`
	APIKeyValue      types.String `tfsdk:"api_key_value"`
}

func (r *ProjectServiceAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI Project Service Account.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the project service account (project_id:service_account_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the project to which the service account belongs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the service account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_account_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the service account.",
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The timestamp (in Unix time) when the service account was created.",
			},
			"role": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The role of the service account.",
			},
			"api_key_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the API key associated with the service account.",
			},
			"api_key_value": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The value of the API key associated with the service account (only available upon creation).",
			},
		},
	}
}

func (r *ProjectServiceAccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectServiceAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()

	createRequest := ProjectServiceAccountCreateRequest{
		Name: data.Name.ValueString(),
	}

	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing request", err.Error())
		return
	}

	// API Endpoint for creating service account in a project
	// URL: POST /organization/projects/{project_id}/service_accounts
	url := fmt.Sprintf("%s/organization/projects/%s/service_accounts", r.client.OpenAIClient.APIURL, projectID)
	apiReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiReq.Header.Set("Content-Type", "application/json")
	// Needs Admin Key? SDKv2 uses GetOpenAIClientWithAdminKey
	// We should use AdminAPIKey if available.
	apiKey := r.client.OpenAIClient.APIKey
	if r.client.AdminAPIKey != "" {
		apiKey = r.client.AdminAPIKey
	}
	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
	// Does it need Organization ID? Not explicitly mentioned but safe to add if available
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

	var saResp ProjectServiceAccountResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &saResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ServiceAccountID = types.StringValue(saResp.ID)
	data.ID = types.StringValue(fmt.Sprintf("%s:%s", projectID, saResp.ID))
	data.Role = types.StringValue(saResp.Role)
	data.CreatedAt = types.Int64Value(saResp.CreatedAt)

	if saResp.APIKey != nil {
		data.APIKeyID = types.StringValue(saResp.APIKey.ID)
		data.APIKeyValue = types.StringValue(saResp.APIKey.Value)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectServiceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// ID is project_id:service_account_id
	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		// If ID format is wrong, try to fallback or error?
		// Maybe imported with wrong ID?
		// If loaded from state, ID should be correct.
		resp.Diagnostics.AddError("Invalid ID format", fmt.Sprintf("Expected project_id:service_account_id, got %s", data.ID.ValueString()))
		return
	}
	projectID := idParts[0]
	serviceAccountID := idParts[1]

	data.ProjectID = types.StringValue(projectID)

	url := fmt.Sprintf("%s/organization/projects/%s/service_accounts/%s", r.client.OpenAIClient.APIURL, projectID, serviceAccountID)
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

	var saResp ProjectServiceAccountResponse
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &saResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ServiceAccountID = types.StringValue(saResp.ID)
	data.Name = types.StringValue(saResp.Name)
	data.Role = types.StringValue(saResp.Role)
	data.CreatedAt = types.Int64Value(saResp.CreatedAt)

	if saResp.APIKey != nil {
		data.APIKeyID = types.StringValue(saResp.APIKey.ID)
		// Value not returned in read
	}

	// Preserve APIKeyValue from state if present
	// data.APIKeyValue is already populated from req.State.Get at the top

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Immutable. Name update usually implies replacement for service accounts in many systems,
	// Is update supported? SDKv2 says "ForceNew" for Name.
	// So update is not supported.
}

func (r *ProjectServiceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectServiceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) != 2 {
		return
	}
	projectID := idParts[0]
	serviceAccountID := idParts[1]

	url := fmt.Sprintf("%s/organization/projects/%s/service_accounts/%s", r.client.OpenAIClient.APIURL, projectID, serviceAccountID)
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

func (r *ProjectServiceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
