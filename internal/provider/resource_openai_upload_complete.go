package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UploadCompleteResource{}

func NewUploadCompleteResource() resource.Resource {
	return &UploadCompleteResource{}
}

type UploadCompleteResource struct {
	client *OpenAIClient
}

type UploadCompleteResourceModel struct {
	ID       types.String      `tfsdk:"id"`
	UploadID types.String      `tfsdk:"upload_id"`
	Parts    []UploadPartModel `tfsdk:"parts"`
}

type UploadPartModel struct {
	PartNumber types.Int64  `tfsdk:"part_number"`
	ETag       types.String `tfsdk:"etag"`
}

func (r *UploadCompleteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_upload_complete"
}

func (r *UploadCompleteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Completes a multipart upload.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"upload_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parts": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"part_number": schema.Int64Attribute{Required: true},
						"etag":        schema.StringAttribute{Required: true},
					},
				},
			},
		},
	}
}

func (r *UploadCompleteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UploadCompleteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UploadCompleteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := make([]map[string]interface{}, len(data.Parts))
	for i, p := range data.Parts {
		parts[i] = map[string]interface{}{
			"part_id": p.ETag.ValueString(), // API expects 'part_ids'? No, API usage: "part_ids": ["..."]
		}
	}
	// API spec for Complete Upload:
	// POST /v1/uploads/{upload_id}/complete
	// Body: { "part_ids": ["part_id_1", "part_id_2"] }

	partIDs := make([]string, len(data.Parts))
	for i, p := range data.Parts {
		partIDs[i] = p.ETag.ValueString()
	}

	reqBodyStruct := struct {
		PartIDs []string `json:"part_ids"`
	}{
		PartIDs: partIDs,
	}

	reqBody, _ := json.Marshal(reqBodyStruct)

	url := fmt.Sprintf("%s/uploads/%s/complete", r.client.OpenAIClient.APIURL, data.UploadID.ValueString())
	if strings.Contains(r.client.OpenAIClient.APIURL, "/v1") {
		url = strings.TrimSuffix(r.client.OpenAIClient.APIURL, "/v1") + "/v1/uploads/" + data.UploadID.ValueString() + "/complete"
	} else {
		url = fmt.Sprintf("%s/v1/uploads/%s/complete", r.client.OpenAIClient.APIURL, data.UploadID.ValueString())
	}

	apiReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
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

	var completeResp struct {
		ID string `json:"id"` // Upload ID
		// ...
	}
	json.NewDecoder(apiResp.Body).Decode(&completeResp)

	data.ID = types.StringValue(completeResp.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UploadCompleteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UploadCompleteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/uploads/%s", r.client.OpenAIClient.APIURL, data.UploadID.ValueString())
	if strings.Contains(r.client.OpenAIClient.APIURL, "/v1") {
		url = strings.TrimSuffix(r.client.OpenAIClient.APIURL, "/v1") + "/v1/uploads/" + data.UploadID.ValueString()
	} else {
		url = fmt.Sprintf("%s/v1/uploads/%s", r.client.OpenAIClient.APIURL, data.UploadID.ValueString())
	}

	apiReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}
	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
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

	// If it exists, we assume the completion resource exists.
	// We could verify status == "completed", but for now just existence of Upload ID is enough.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UploadCompleteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *UploadCompleteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Cannot undo completion
}
