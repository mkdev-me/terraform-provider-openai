package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UploadResource{}
var _ resource.ResourceWithImportState = &UploadResource{}

func NewUploadResource() resource.Resource {
	return &UploadResource{}
}

type UploadResource struct {
	client *OpenAIClient
}

type UploadResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Purpose   types.String `tfsdk:"purpose"`
	Filename  types.String `tfsdk:"filename"`
	Bytes     types.Int64  `tfsdk:"bytes"`
	MimeType  types.String `tfsdk:"mime_type"`
	File      types.String `tfsdk:"file"`
	ProjectID types.String `tfsdk:"project_id"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
	Status    types.String `tfsdk:"status"`
}

func (r *UploadResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_upload"
}

func (r *UploadResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages file uploads for various purposes including fine-tuning, assistants, etc.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the upload.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"purpose": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The intended purpose of the uploaded file. Can be 'fine-tune', 'assistants', 'vision', 'batch', 'user_data', or 'evals'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"filename": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The name of the file to upload.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bytes": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The number of bytes in the file being uploaded.",
			},
			"mime_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The MIME type of the file.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The path to the file to upload. Required for creation, not needed for import.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The project ID to associate this upload with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the upload was created.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the upload (e.g., 'pending', 'completed').",
			},
		},
	}
}

func (r *UploadResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UploadResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UploadResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	purpose := data.Purpose.ValueString()
	filePath := data.File.ValueString()

	if filePath == "" {
		resp.Diagnostics.AddError("Missing File", "File path is required for creating a new upload")
		return
	}

	var filename string
	if !data.Filename.IsNull() {
		filename = data.Filename.ValueString()
	} else {
		filename = filepath.Base(filePath)
	}

	var fileBytes int64
	if !data.Bytes.IsNull() {
		fileBytes = data.Bytes.ValueInt64()
	} else {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			resp.Diagnostics.AddError("Error getting file size", err.Error())
			return
		}
		fileBytes = fileInfo.Size()
	}

	var mimeType string
	if !data.MimeType.IsNull() {
		mimeType = data.MimeType.ValueString()
	} else {
		ext := filepath.Ext(filename)
		switch strings.ToLower(ext) {
		case ".jsonl":
			mimeType = "application/jsonl"
		case ".json":
			mimeType = "application/json"
		case ".txt":
			mimeType = "text/plain"
		case ".csv":
			mimeType = "text/csv"
		case ".pdf":
			mimeType = "application/pdf"
		case ".md":
			mimeType = "text/markdown"
		default:
			mimeType = "application/octet-stream"
		}
	}

	// Prepare request
	uploadReq := struct {
		Purpose  string `json:"purpose"`
		Filename string `json:"filename"`
		Bytes    int64  `json:"bytes"`
		MimeType string `json:"mime_type"`
	}{
		Purpose:  purpose,
		Filename: filename,
		Bytes:    fileBytes,
		MimeType: mimeType,
	}

	reqBody, err := json.Marshal(uploadReq)
	if err != nil {
		resp.Diagnostics.AddError("Error marshalling request", err.Error())
		return
	}

	url := fmt.Sprintf("%s/uploads", r.client.OpenAIClient.APIURL)
	// Handle /v1 suffix provided in APIURL
	if strings.HasSuffix(url, "/v1/uploads") {
		// correct
	} else if strings.HasSuffix(url, "/v1") {
		url += "/uploads"
	} else {
		url = fmt.Sprintf("%s/v1/uploads", r.client.OpenAIClient.APIURL)
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

	var uploadResp struct {
		ID        string `json:"id"`
		Filename  string `json:"filename"`
		Bytes     int64  `json:"bytes"`
		CreatedAt int64  `json:"created_at"`
		Status    string `json:"status"`
		Purpose   string `json:"purpose"`
	}

	if err := json.NewDecoder(apiResp.Body).Decode(&uploadResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(uploadResp.ID)
	data.Status = types.StringValue(uploadResp.Status)
	data.CreatedAt = types.Int64Value(uploadResp.CreatedAt)
	data.Purpose = types.StringValue(uploadResp.Purpose)
	data.Filename = types.StringValue(uploadResp.Filename)
	data.Bytes = types.Int64Value(uploadResp.Bytes)
	data.MimeType = types.StringValue(mimeType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UploadResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UploadResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/uploads/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	if strings.Contains(r.client.OpenAIClient.APIURL, "/v1") {
		url = strings.TrimSuffix(r.client.OpenAIClient.APIURL, "/v1") + "/v1/uploads/" + data.ID.ValueString()
	} else {
		url = fmt.Sprintf("%s/v1/uploads/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
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

	var uploadResp struct {
		ID        string `json:"id"`
		Filename  string `json:"filename"`
		Bytes     int64  `json:"bytes"`
		CreatedAt int64  `json:"created_at"`
		Status    string `json:"status"`
		Purpose   string `json:"purpose"`
	}

	if err := json.NewDecoder(apiResp.Body).Decode(&uploadResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.Status = types.StringValue(uploadResp.Status)
	data.CreatedAt = types.Int64Value(uploadResp.CreatedAt)
	data.Purpose = types.StringValue(uploadResp.Purpose)
	data.Filename = types.StringValue(uploadResp.Filename)
	data.Bytes = types.Int64Value(uploadResp.Bytes)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UploadResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Uploads are immutable usually, but we can implement update cancellation if supported
	// For now, no update logic needed as most fields are RequiresReplace
}

func (r *UploadResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UploadResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/uploads/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	// Clean url construction
	if strings.Contains(r.client.OpenAIClient.APIURL, "/v1") {
		url = strings.TrimSuffix(r.client.OpenAIClient.APIURL, "/v1") + "/v1/uploads/" + data.ID.ValueString()
	} else {
		url = fmt.Sprintf("%s/v1/uploads/%s", r.client.OpenAIClient.APIURL, data.ID.ValueString())
	}

	// Make cancel request to "cancel" the upload
	apiReq, err := http.NewRequest("POST", url+"/cancel", nil)
	if err != nil {
		// Fallback to attempts? The API reference says cancel.
		// If we can't cancel, we just leave it?
		// Note from API docs: "Cancels an upload."
		return
	}

	apiReq.Header.Set("Authorization", "Bearer "+r.client.OpenAIClient.APIKey)
	if r.client.OpenAIClient.OrganizationID != "" {
		apiReq.Header.Set("OpenAI-Organization", r.client.OpenAIClient.OrganizationID)
	}

	http.DefaultClient.Do(apiReq)
	// We don't error if cancel fails, just attempt it.
}

func (r *UploadResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
