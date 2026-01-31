package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

// Ensure implementation satisfies interfaces.
var _ resource.Resource = &FileResource{}
var _ resource.ResourceWithImportState = &FileResource{}

// FileResource defines the resource implementation.
type FileResource struct {
	client *OpenAIClient
}

// FileResourceModel describes the resource data model.
type FileResourceModel struct {
	ID        types.String `tfsdk:"id"`
	File      types.String `tfsdk:"file"`
	Purpose   types.String `tfsdk:"purpose"`
	ProjectID types.String `tfsdk:"project_id"`
	Filename  types.String `tfsdk:"filename"`
	Bytes     types.Int64  `tfsdk:"bytes"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func NewFileResource() resource.Resource {
	return &FileResource{}
}

func (r *FileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

func (r *FileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The file resource allows users to upload, read, and delete files through the OpenAI API.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the file.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file": schema.StringAttribute{
				MarkdownDescription: "Path to the file to upload. Required for creation, ignored during import.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"purpose": schema.StringAttribute{
				MarkdownDescription: "The purpose of the file. Can be 'fine-tune', 'assistants', 'vision', or 'batch'. Required for creation, computed for import.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The project ID to associate this file with (for Terraform reference only, not sent to OpenAI API)",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"filename": schema.StringAttribute{
				MarkdownDescription: "The name of the file",
				Computed:            true,
			},
			"bytes": schema.Int64Attribute{
				MarkdownDescription: "The size of the file in bytes",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "The timestamp for when the file was created",
				Computed:            true,
			},
		},
	}
}

func (r *FileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*OpenAIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider.OpenAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *FileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FileResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	filePath := data.File.ValueString()
	// Read the file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		resp.Diagnostics.AddError("Error reading file", fmt.Sprintf("Error reading file %s: %s", filePath, err))
		return
	}

	url := fmt.Sprintf("%s/v1/files", r.client.OpenAIClient.APIURL)
	if strings.Contains(r.client.OpenAIClient.APIURL, "/v1") {
		url = fmt.Sprintf("%s/files", r.client.OpenAIClient.APIURL)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		resp.Diagnostics.AddError("Error creating form file", err.Error())
		return
	}
	_, err = part.Write(fileContent)
	if err != nil {
		resp.Diagnostics.AddError("Error writing file content", err.Error())
		return
	}

	err = writer.WriteField("purpose", data.Purpose.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error writing purpose field", err.Error())
		return
	}

	err = writer.Close()
	if err != nil {
		resp.Diagnostics.AddError("Error closing writer", err.Error())
		return
	}

	apiReq, err := http.NewRequest("POST", url, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating request", err.Error())
		return
	}

	apiReq.Header.Set("Content-Type", writer.FormDataContentType())
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

	responseBody, err := io.ReadAll(apiResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Error reading response", err.Error())
		return
	}

	var fileResponse FileResponse
	err = json.Unmarshal(responseBody, &fileResponse)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.ID = types.StringValue(fileResponse.ID)
	data.Filename = types.StringValue(fileResponse.Filename)
	data.Bytes = types.Int64Value(fileResponse.Bytes)
	data.CreatedAt = types.Int64Value(fileResponse.CreatedAt)
	// Purpose is already in data
	// ProjectID is already in data (if set)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.client.OpenAIClient
	// Use project key if needed, analogous to GetOpenAIClient logic which defaults to standard client
	// but lets stick to the configured client in `r.client`.

	url := fmt.Sprintf("%s/v1/files/%s", client.APIURL, data.ID.ValueString())
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/files/%s", client.APIURL, data.ID.ValueString())
	}

	respBody, err := client.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		// Handle 404
		if strings.Contains(err.Error(), "404") { // Simple check, ideally check status code
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving file", err.Error())
		return
	}

	var fileResponse FileResponse
	err = json.Unmarshal(respBody, &fileResponse)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.Filename = types.StringValue(fileResponse.Filename)
	data.Bytes = types.Int64Value(fileResponse.Bytes)
	data.CreatedAt = types.Int64Value(fileResponse.CreatedAt)
	data.Purpose = types.StringValue(fileResponse.Purpose)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Files are immutable in OpenAI, so any change requires replacement (handled by plan modifiers)
	// Theoretically we shouldn't get here if PlanModifiers are set correctly.
}

func (r *FileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.client.OpenAIClient
	url := fmt.Sprintf("%s/v1/files/%s", client.APIURL, data.ID.ValueString())
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/files/%s", client.APIURL, data.ID.ValueString())
	}

	_, err := client.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		// If 404, consider it gone
		if strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting file", err.Error())
		return
	}
}

func (r *FileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Note: We cannot populate the 'file' attribute (path) from import because the API doesn't return it.
	// Users will have to adjust state or config manually if they want to manage the file content.
	// This matches the SDKv2 behavior where 'file' was ignored during import, or we tried to guess it.
	// In the framework, we can leave it null or try to populate it if possible, but usually for file paths
	// we can't really guess.
}
