package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const organizationSpendLimitID = "organization"

var _ resource.Resource = &ProjectSpendLimitResource{}
var _ resource.ResourceWithImportState = &ProjectSpendLimitResource{}
var _ resource.Resource = &OrganizationSpendLimitResource{}
var _ resource.ResourceWithImportState = &OrganizationSpendLimitResource{}

type ProjectSpendLimitResource struct {
	client *OpenAIClient
}

type OrganizationSpendLimitResource struct {
	client *OpenAIClient
}

type SpendLimitEnforcementModel struct {
	Status types.String `tfsdk:"status"`
}

type ProjectSpendLimitResourceModel struct {
	ID              types.String                `tfsdk:"id"`
	ProjectID       types.String                `tfsdk:"project_id"`
	ThresholdAmount types.Int64                 `tfsdk:"threshold_amount"`
	Currency        types.String                `tfsdk:"currency"`
	Interval        types.String                `tfsdk:"interval"`
	Enforcement     *SpendLimitEnforcementModel `tfsdk:"enforcement"`
	Object          types.String                `tfsdk:"object"`
}

type OrganizationSpendLimitResourceModel struct {
	ID              types.String                `tfsdk:"id"`
	ThresholdAmount types.Int64                 `tfsdk:"threshold_amount"`
	Currency        types.String                `tfsdk:"currency"`
	Interval        types.String                `tfsdk:"interval"`
	Enforcement     *SpendLimitEnforcementModel `tfsdk:"enforcement"`
	Object          types.String                `tfsdk:"object"`
}

func spendLimitAttributes(idDescription string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: idDescription,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"threshold_amount": schema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "The hard spend limit in cents.",
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
		"currency": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("USD"),
			MarkdownDescription: "The threshold currency. Currently only `USD` is supported.",
			Validators:          []validator.String{stringvalidator.OneOf("USD")},
		},
		"interval": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("month"),
			MarkdownDescription: "The evaluation interval. Currently only `month` is supported.",
			Validators:          []validator.String{stringvalidator.OneOf("month")},
		},
		"enforcement": schema.SingleNestedAttribute{
			Computed:            true,
			MarkdownDescription: "The current enforcement state of the hard spend limit.",
			Attributes: map[string]schema.Attribute{
				"status": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "Whether the hard spend limit is `inactive` or `enforcing`.",
				},
			},
		},
		"object": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The OpenAI object type.",
		},
	}
}

func configureSpendLimitResource(providerData any, target **OpenAIClient, resp *resource.ConfigureResponse) {
	if providerData == nil {
		return
	}
	client, ok := providerData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *provider.OpenAIClient, got: %T", providerData))
		return
	}
	*target = client
}

func spendLimitRequestFromValues(thresholdAmount types.Int64, currency, interval types.String) spendLimitRequest {
	return spendLimitRequest{
		ThresholdAmount: thresholdAmount.ValueInt64(),
		Currency:        currency.ValueString(),
		Interval:        interval.ValueString(),
	}
}

func spendLimitEnforcementState(value *SpendLimitAPI) *SpendLimitEnforcementModel {
	return &SpendLimitEnforcementModel{Status: types.StringValue(value.Enforcement.Status)}
}

func NewProjectSpendLimitResource() resource.Resource {
	return &ProjectSpendLimitResource{}
}

func (r *ProjectSpendLimitResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_spend_limit"
}

func (r *ProjectSpendLimitResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := spendLimitAttributes("The project ID.")
	attributes["project_id"] = schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The ID of the project.",
		PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the singleton hard spend limit for an OpenAI project. Threshold amounts are specified in cents. Only one hard spend limit exists per project; multiple resources targeting the same project overwrite one another.",
		Attributes:          attributes,
	}
}

func (r *ProjectSpendLimitResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureSpendLimitResource(req.ProviderData, &r.client, resp)
}

func setProjectSpendLimitState(data *ProjectSpendLimitResourceModel, value *SpendLimitAPI) {
	data.ID = types.StringValue(data.ProjectID.ValueString())
	data.ThresholdAmount = types.Int64Value(value.ThresholdAmount)
	data.Currency = types.StringValue(value.Currency)
	data.Interval = types.StringValue(value.Interval)
	data.Enforcement = spendLimitEnforcementState(value)
	data.Object = types.StringValue(value.Object)
}

func (r *ProjectSpendLimitResource) write(ctx context.Context, data *ProjectSpendLimitResourceModel) (*SpendLimitAPI, error) {
	request := spendLimitRequestFromValues(data.ThresholdAmount, data.Currency, data.Interval)
	var value SpendLimitAPI
	_, err := projectSettingsRequest(
		ctx,
		r.client,
		http.MethodPost,
		"/v1/organization/projects/"+data.ProjectID.ValueString()+"/spend_limit",
		request,
		&value,
	)
	if err == nil {
		invalidateProjectSpendLimitCache(r.client, data.ProjectID.ValueString())
	}
	return &value, err
}

func (r *ProjectSpendLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectSpendLimitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.write(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project spend limit", err.Error())
		return
	}
	setProjectSpendLimitState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectSpendLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectSpendLimitResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, status, err := cachedProjectSpendLimit(ctx, r.client, data.ProjectID.ValueString())
	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading project spend limit", err.Error())
		return
	}
	setProjectSpendLimitState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectSpendLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectSpendLimitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.write(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project spend limit", err.Error())
		return
	}
	setProjectSpendLimitState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectSpendLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectSpendLimitResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := projectSettingsRequest(
		ctx,
		r.client,
		http.MethodDelete,
		"/v1/organization/projects/"+data.ProjectID.ValueString()+"/spend_limit",
		nil,
		nil,
	)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Error deleting project spend limit", err.Error())
		return
	}
	invalidateProjectSpendLimitCache(r.client, data.ProjectID.ValueString())
}

func (r *ProjectSpendLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be a project ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), req.ID)...)
}

func NewOrganizationSpendLimitResource() resource.Resource {
	return &OrganizationSpendLimitResource{}
}

func (r *OrganizationSpendLimitResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_spend_limit"
}

func (r *OrganizationSpendLimitResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the singleton hard spend limit for the OpenAI organization. Threshold amounts are specified in cents. Only one organization hard spend limit exists; multiple resources overwrite one another.",
		Attributes:          spendLimitAttributes("The stable singleton identifier, always `organization`."),
	}
}

func (r *OrganizationSpendLimitResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureSpendLimitResource(req.ProviderData, &r.client, resp)
}

func setOrganizationSpendLimitState(data *OrganizationSpendLimitResourceModel, value *SpendLimitAPI) {
	data.ID = types.StringValue(organizationSpendLimitID)
	data.ThresholdAmount = types.Int64Value(value.ThresholdAmount)
	data.Currency = types.StringValue(value.Currency)
	data.Interval = types.StringValue(value.Interval)
	data.Enforcement = spendLimitEnforcementState(value)
	data.Object = types.StringValue(value.Object)
}

func (r *OrganizationSpendLimitResource) write(ctx context.Context, data *OrganizationSpendLimitResourceModel) (*SpendLimitAPI, error) {
	request := spendLimitRequestFromValues(data.ThresholdAmount, data.Currency, data.Interval)
	var value SpendLimitAPI
	_, err := projectSettingsRequest(ctx, r.client, http.MethodPost, "/v1/organization/spend_limit", request, &value)
	if err == nil {
		invalidateOrganizationSpendLimitCache(r.client)
	}
	return &value, err
}

func (r *OrganizationSpendLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationSpendLimitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.write(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error creating organization spend limit", err.Error())
		return
	}
	setOrganizationSpendLimitState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationSpendLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationSpendLimitResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, status, err := cachedOrganizationSpendLimit(ctx, r.client)
	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading organization spend limit", err.Error())
		return
	}
	setOrganizationSpendLimitState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationSpendLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrganizationSpendLimitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.write(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error updating organization spend limit", err.Error())
		return
	}
	setOrganizationSpendLimitState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationSpendLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationSpendLimitResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := projectSettingsRequest(ctx, r.client, http.MethodDelete, "/v1/organization/spend_limit", nil, nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Error deleting organization spend limit", err.Error())
		return
	}
	invalidateOrganizationSpendLimitCache(r.client)
}

func (r *OrganizationSpendLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != organizationSpendLimitID {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be `organization`")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), organizationSpendLimitID)...)
}
