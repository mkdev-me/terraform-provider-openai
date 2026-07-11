package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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

var _ resource.Resource = &ProjectSpendAlertResource{}
var _ resource.ResourceWithImportState = &ProjectSpendAlertResource{}

type ProjectSpendAlertResource struct {
	client *OpenAIClient
}

type ProjectSpendAlertNotificationModel struct {
	Type          types.String `tfsdk:"type"`
	Recipients    types.Set    `tfsdk:"recipients"`
	SubjectPrefix types.String `tfsdk:"subject_prefix"`
}

type ProjectSpendAlertResourceModel struct {
	ID                  types.String                        `tfsdk:"id"`
	AlertID             types.String                        `tfsdk:"alert_id"`
	ProjectID           types.String                        `tfsdk:"project_id"`
	ThresholdAmount     types.Int64                         `tfsdk:"threshold_amount"`
	Currency            types.String                        `tfsdk:"currency"`
	Interval            types.String                        `tfsdk:"interval"`
	NotificationChannel *ProjectSpendAlertNotificationModel `tfsdk:"notification_channel"`
	Object              types.String                        `tfsdk:"object"`
}

func NewProjectSpendAlertResource() resource.Resource {
	return &ProjectSpendAlertResource{}
}

func (r *ProjectSpendAlertResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_spend_alert"
}

func (r *ProjectSpendAlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an email spend alert for an OpenAI project. Threshold amounts are specified in cents.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "The composite `project_id:alert_id` identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"alert_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "The OpenAI spend alert ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project_id": schema.StringAttribute{Required: true, MarkdownDescription: "The ID of the project.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"threshold_amount": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The spend threshold in cents.",
				Validators:          []validator.Int64{int64validator.AtLeast(0)},
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
			"notification_channel": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("email"), MarkdownDescription: "The notification type. Currently only `email` is supported.", Validators: []validator.String{stringvalidator.OneOf("email")}},
					"recipients": schema.SetAttribute{
						Required:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Email addresses that receive the alert.",
						Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
					},
					"subject_prefix": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional subject prefix for alert emails."},
				},
			},
			"object": schema.StringAttribute{Computed: true, MarkdownDescription: "The OpenAI object type."},
		},
	}
}

func (r *ProjectSpendAlertResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func spendAlertAPIFromModel(data *ProjectSpendAlertResourceModel) projectSpendAlertRequest {
	notification := ProjectSpendAlertNotificationAPI{
		Type:       data.NotificationChannel.Type.ValueString(),
		Recipients: modelIDsFromSet(data.NotificationChannel.Recipients),
	}
	if !data.NotificationChannel.SubjectPrefix.IsNull() {
		value := data.NotificationChannel.SubjectPrefix.ValueString()
		notification.SubjectPrefix = &value
	}
	return projectSpendAlertRequest{
		ThresholdAmount:     data.ThresholdAmount.ValueInt64(),
		Currency:            data.Currency.ValueString(),
		Interval:            data.Interval.ValueString(),
		NotificationChannel: notification,
	}
}

func setProjectSpendAlertState(data *ProjectSpendAlertResourceModel, value *ProjectSpendAlertAPI) {
	data.AlertID = types.StringValue(value.ID)
	data.ID = types.StringValue(data.ProjectID.ValueString() + ":" + value.ID)
	data.ThresholdAmount = types.Int64Value(value.ThresholdAmount)
	data.Currency = types.StringValue(value.Currency)
	data.Interval = types.StringValue(value.Interval)
	data.Object = types.StringValue(value.Object)
	data.NotificationChannel = &ProjectSpendAlertNotificationModel{
		Type:       types.StringValue(value.NotificationChannel.Type),
		Recipients: modelIDsToSet(value.NotificationChannel.Recipients),
	}
	if value.NotificationChannel.SubjectPrefix != nil {
		data.NotificationChannel.SubjectPrefix = types.StringValue(*value.NotificationChannel.SubjectPrefix)
	} else {
		data.NotificationChannel.SubjectPrefix = types.StringNull()
	}
}

func (r *ProjectSpendAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectSpendAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	request := spendAlertAPIFromModel(&data)
	var value ProjectSpendAlertAPI
	_, err := projectSettingsRequest(ctx, r.client, http.MethodPost, "/v1/organization/projects/"+data.ProjectID.ValueString()+"/spend_alerts", request, &value)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project spend alert", err.Error())
		return
	}
	invalidateProjectSpendAlertsCache(r.client, data.ProjectID.ValueString())
	setProjectSpendAlertState(&data, &value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectSpendAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectSpendAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, status, err := cachedProjectSpendAlert(ctx, r.client, data.ProjectID.ValueString(), data.AlertID.ValueString())
	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading project spend alert", err.Error())
		return
	}
	setProjectSpendAlertState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectSpendAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectSpendAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	request := spendAlertAPIFromModel(&data)
	var value ProjectSpendAlertAPI
	_, err := projectSettingsRequest(ctx, r.client, http.MethodPost, "/v1/organization/projects/"+data.ProjectID.ValueString()+"/spend_alerts/"+data.AlertID.ValueString(), request, &value)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project spend alert", err.Error())
		return
	}
	invalidateProjectSpendAlertsCache(r.client, data.ProjectID.ValueString())
	setProjectSpendAlertState(&data, &value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectSpendAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectSpendAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := projectSettingsRequest(ctx, r.client, http.MethodDelete, "/v1/organization/projects/"+data.ProjectID.ValueString()+"/spend_alerts/"+data.AlertID.ValueString(), nil, nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Error deleting project spend alert", err.Error())
		return
	}
	invalidateProjectSpendAlertsCache(r.client, data.ProjectID.ValueString())
}

func (r *ProjectSpendAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format project_id:alert_id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("alert_id"), parts[1])...)
}
