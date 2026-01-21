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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ImageEditResource{}
var _ resource.ResourceWithImportState = &ImageEditResource{}

type ImageEditResource struct {
	client *OpenAIClient
}

func NewImageEditResource() resource.Resource {
	return &ImageEditResource{}
}

func (r *ImageEditResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_edit"
}

type ImageEditResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Image          types.String `tfsdk:"image"`
	Mask           types.String `tfsdk:"mask"`
	Prompt         types.String `tfsdk:"prompt"`
	Model          types.String `tfsdk:"model"`
	N              types.Int64  `tfsdk:"n"`
	ResponseFormat types.String `tfsdk:"response_format"`
	Size           types.String `tfsdk:"size"`
	User           types.String `tfsdk:"user"`

	Created types.Int64 `tfsdk:"created"`
	Data    types.List  `tfsdk:"data"` // List of Objects
}

func (r *ImageEditResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Edits an image.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mask": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"n": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					planmodifier.Int64(nil), // Should act as ForceNew
				},
			},
			"response_format": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"size": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"data": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"url":      types.StringType,
						"b64_json": types.StringType,
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ImageEditResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	rc_client, ok := req.ProviderData.(*OpenAIClient)
	if ok {
		r.client = rc_client
	}
}

func (r *ImageEditResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ImageEditResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare multipart request
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	writer.WriteField("prompt", data.Prompt.ValueString())

	if !data.Model.IsNull() {
		writer.WriteField("model", data.Model.ValueString())
	}
	if !data.N.IsNull() {
		writer.WriteField("n", fmt.Sprintf("%d", data.N.ValueInt64()))
	}
	if !data.Size.IsNull() {
		writer.WriteField("size", data.Size.ValueString())
	}
	if !data.ResponseFormat.IsNull() {
		writer.WriteField("response_format", data.ResponseFormat.ValueString())
	}
	if !data.User.IsNull() {
		writer.WriteField("user", data.User.ValueString())
	}

	// Image
	file, err := os.Open(data.Image.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error opening image", err.Error())
		return
	}
	defer file.Close()

	part, err := writer.CreateFormFile("image", filepath.Base(data.Image.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error adding image to form", err.Error())
		return
	}
	io.Copy(part, file)

	// Mask
	if !data.Mask.IsNull() {
		mfile, err := os.Open(data.Mask.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error opening mask", err.Error())
			return
		}
		defer mfile.Close()

		mpart, err := writer.CreateFormFile("mask", filepath.Base(data.Mask.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Error adding mask to form", err.Error())
			return
		}
		io.Copy(mpart, mfile)
	}

	writer.Close()

	url := fmt.Sprintf("%s/images/edits", r.client.OpenAIClient.APIURL)

	apiReq, err := http.NewRequest("POST", url, &requestBody)
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

	if apiResp.StatusCode != http.StatusOK {
		respBodyBytes, _ := io.ReadAll(apiResp.Body)
		resp.Diagnostics.AddError("API error", fmt.Sprintf("API returned error: %s - %s", apiResp.Status, string(respBodyBytes)))
		return
	}

	var imgResp ImageEditResponseFramework
	respBodyBytes, _ := io.ReadAll(apiResp.Body)
	if err := json.Unmarshal(respBodyBytes, &imgResp); err != nil {
		resp.Diagnostics.AddError("Error parsing response", err.Error())
		return
	}

	data.Created = types.Int64Value(imgResp.Created)

	if len(imgResp.Data) > 0 {
		objs := []attr.Value{}
		for _, d := range imgResp.Data {
			obj, _ := types.ObjectValue(
				map[string]attr.Type{
					"url":      types.StringType,
					"b64_json": types.StringType,
				},
				map[string]attr.Value{
					"url":      types.StringValue(d.URL),
					"b64_json": types.StringValue(d.B64JSON),
				},
			)
			objs = append(objs, obj)
		}
		listVal, _ := types.ListValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"url":      types.StringType,
				"b64_json": types.StringType,
			},
		}, objs)
		data.Data = listVal
	} else {
		data.Data = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
			"url":      types.StringType,
			"b64_json": types.StringType,
		}})
	}

	data.ID = types.StringValue(fmt.Sprintf("img-edit-%d", imgResp.Created))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ImageEditResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ImageEditResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ImageEditResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *ImageEditResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *ImageEditResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
