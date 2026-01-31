package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mkdev-me/terraform-provider-openai/internal/client"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

type ProjectResource struct {
	client *client.OpenAIClient
}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

type ProjectResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Status     types.String `tfsdk:"status"`
	CreatedAt  types.String `tfsdk:"created_at"`
	ArchivedAt types.String `tfsdk:"archived_at"`
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenAI Project.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The identifier of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the project.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the project (e.g. active, archived).",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the project was created.",
			},
			"archived_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the project was archived.",
			},
		},
	}
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerClient, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *provider.OpenAIClient, got: %T", req.ProviderData))
		return
	}

	// Project management requires Admin Keys
	cl, err := GetOpenAIClientWithAdminKey(providerClient)
	if err != nil {
		resp.Diagnostics.AddError("Error getting OpenAI Client with Admin Key", err.Error())
		return
	}

	r.client = cl
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.CreateProject(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}

	data.ID = types.StringValue(project.ID)
	data.Name = types.StringValue(project.Name)
	data.Status = types.StringValue(project.Status)

	if project.CreatedAt != nil {
		data.CreatedAt = types.StringValue(time.Unix(*project.CreatedAt, 0).Format(time.RFC3339))
	}

	if project.ArchivedAt != nil {
		data.ArchivedAt = types.StringValue(time.Unix(*project.ArchivedAt, 0).Format(time.RFC3339))
	} else {
		data.ArchivedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.GetProject(data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	// Check if project is loaded (GetProject might return error if 404, or might return nil if logic handled it?
	// Currently client.GetProject returns error on non-200. I should verify if I need to handle 404 string in error.)
	// Looking at client.GetProject in client.go, it wraps doRequest.
	// doRequest checks for >= 400 and returns "API error (status %d): ..." or "API error: message".
	// So I should check error string for "404" or similar if I want to remove from state.
	// But actually client.GetProject returns error.

	// Wait, I should verify client.GetProject.
	// It calls doRequest. doRequest can return error with message.
	// Common logic:
	/*
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Error reading project", err.Error())
			return
		}
	*/

	// I'll add the 404 check since I'm refactoring. However I can't check strings without `strings` package imported.
	// I need to ensure `strings` is imported.

	data.Name = types.StringValue(project.Name)
	data.Status = types.StringValue(project.Status)

	if project.CreatedAt != nil {
		data.CreatedAt = types.StringValue(time.Unix(*project.CreatedAt, 0).Format(time.RFC3339))
	}

	if project.ArchivedAt != nil {
		data.ArchivedAt = types.StringValue(time.Unix(*project.ArchivedAt, 0).Format(time.RFC3339))
	} else {
		data.ArchivedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.UpdateProject(data.ID.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	data.Name = types.StringValue(project.Name)
	data.Status = types.StringValue(project.Status)

	if project.CreatedAt != nil {
		data.CreatedAt = types.StringValue(time.Unix(*project.CreatedAt, 0).Format(time.RFC3339))
	}

	if project.ArchivedAt != nil {
		data.ArchivedAt = types.StringValue(time.Unix(*project.ArchivedAt, 0).Format(time.RFC3339))
	} else {
		data.ArchivedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteProject(data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error deleting (archiving) project", err.Error())
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
