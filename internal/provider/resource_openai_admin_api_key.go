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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AdminAPIKeyResource{}
var _ resource.ResourceWithImportState = &AdminAPIKeyResource{}

type AdminAPIKeyResource struct {
	client *OpenAIClient
}

func NewAdminAPIKeyResource() resource.Resource {
	return &AdminAPIKeyResource{}
}

func (r *AdminAPIKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_api_key"
}

type AdminAPIKeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Scopes      types.List   `tfsdk:"scopes"`
	ExpiresAt   types.Int64  `tfsdk:"expires_at"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
	APIKeyValue types.String `tfsdk:"api_key_value"`
	Object      types.String `tfsdk:"object"`
}

func (r *AdminAPIKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI Admin API Key.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the API Key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scopes": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Scopes to assign to the API key.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"expires_at": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Unix timestamp when the API key should expire.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The timestamp (in Unix time) when the API key was created.",
			},
			"object": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The object type.",
			},
			"api_key_value": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The value of the API key (only available upon creation).",
			},
		},
	}
}

func (r *AdminAPIKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AdminAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AdminAPIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := AdminAPIKeyCreateRequest{
		Name: data.Name.ValueString(),
	}

	if !data.Scopes.IsNull() {
		var scopes []string
		// Converting types.List to []string
		data.Scopes.ElementsAs(ctx, &scopes, false)
		createRequest.Scopes = scopes
	}

	reqBody, err := json.Marshal(createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing request", err.Error())
		return
	}

	// Admin API Keys creation
	// POST /organization/admin_api_keys
	url := fmt.Sprintf("%s/organization/admin_api_keys", r.client.OpenAIClient.APIURL)
	apiReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
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

	apiResp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error making request", err.Error())
		return
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK {
		respBodyBytes, _ := io.ReadAll(apiResp.Body)
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}

	var keyResp AdminAPIKeyResponseFramework
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &keyResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(keyResp.ID)
	data.CreatedAt = types.Int64Value(keyResp.CreatedAt)
	data.Object = types.StringValue(keyResp.Object)
	data.APIKeyValue = types.StringValue(keyResp.Key)
	if keyResp.ExpiresAt != nil {
		data.ExpiresAt = types.Int64Value(*keyResp.ExpiresAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AdminAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AdminAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/organization/admin_api_keys/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
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

	var keyResp AdminAPIKeyResponseFramework
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &keyResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.Name = types.StringValue(keyResp.Name)
	data.CreatedAt = types.Int64Value(keyResp.CreatedAt)
	data.Object = types.StringValue(keyResp.Object)
	if keyResp.ExpiresAt != nil {
		data.ExpiresAt = types.Int64Value(*keyResp.ExpiresAt)
	}

	if len(keyResp.Scopes) > 0 {
		scopes, _ := types.ListValueFrom(ctx, types.StringType, keyResp.Scopes)
		data.Scopes = scopes
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AdminAPIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Immutable
}

func (r *AdminAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AdminAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/organization/admin_api_keys/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
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

func (r *AdminAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
