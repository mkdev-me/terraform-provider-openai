package provider

import (
	"context"
	"fmt"
	"net/http"

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

var _ resource.Resource = &OrganizationSpendAlertResource{}
var _ resource.ResourceWithImportState = &OrganizationSpendAlertResource{}

type OrganizationSpendAlertResource struct {
	client *OpenAIClient
}

type OrganizationSpendAlertResourceModel struct {
	ID                  types.String                 `tfsdk:"id"`
	AlertID             types.String                 `tfsdk:"alert_id"`
	ThresholdAmount     types.Int64                  `tfsdk:"threshold_amount"`
	Currency            types.String                 `tfsdk:"currency"`
	Interval            types.String                 `tfsdk:"interval"`
	NotificationChannel *SpendAlertNotificationModel `tfsdk:"notification_channel"`
	Object              types.String                 `tfsdk:"object"`
}

func NewOrganizationSpendAlertResource() resource.Resource {
	return &OrganizationSpendAlertResource{}
}

func (r *OrganizationSpendAlertResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_spend_alert"
}

func (r *OrganizationSpendAlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an email spend alert for the OpenAI organization. Threshold amounts are specified in cents.",
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true, MarkdownDescription: "The OpenAI spend alert ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"alert_id": schema.StringAttribute{Computed: true, MarkdownDescription: "The OpenAI spend alert ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
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

func (r *OrganizationSpendAlertResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func organizationSpendAlertAPIFromModel(data *OrganizationSpendAlertResourceModel) spendAlertRequest {
	return spendAlertAPIFromValues(data.ThresholdAmount, data.Currency, data.Interval, data.NotificationChannel)
}

func setOrganizationSpendAlertState(data *OrganizationSpendAlertResourceModel, value *SpendAlertAPI) {
	data.ID = types.StringValue(value.ID)
	data.AlertID = types.StringValue(value.ID)
	data.ThresholdAmount = types.Int64Value(value.ThresholdAmount)
	data.Currency = types.StringValue(value.Currency)
	data.Interval = types.StringValue(value.Interval)
	data.Object = types.StringValue(value.Object)
	data.NotificationChannel = &SpendAlertNotificationModel{
		Type:       types.StringValue(value.NotificationChannel.Type),
		Recipients: modelIDsToSet(value.NotificationChannel.Recipients),
	}
	if value.NotificationChannel.SubjectPrefix != nil {
		data.NotificationChannel.SubjectPrefix = types.StringValue(*value.NotificationChannel.SubjectPrefix)
	} else {
		data.NotificationChannel.SubjectPrefix = types.StringNull()
	}
}

func (r *OrganizationSpendAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationSpendAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	request := organizationSpendAlertAPIFromModel(&data)
	var value SpendAlertAPI
	_, err := projectSettingsRequest(ctx, r.client, http.MethodPost, "/v1/organization/spend_alerts", request, &value)
	if err != nil {
		resp.Diagnostics.AddError("Error creating organization spend alert", err.Error())
		return
	}
	invalidateOrganizationSpendAlertsCache(r.client)
	setOrganizationSpendAlertState(&data, &value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationSpendAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationSpendAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, status, err := cachedOrganizationSpendAlert(ctx, r.client, data.AlertID.ValueString())
	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading organization spend alert", err.Error())
		return
	}
	setOrganizationSpendAlertState(&data, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationSpendAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrganizationSpendAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	request := organizationSpendAlertAPIFromModel(&data)
	var value SpendAlertAPI
	_, err := projectSettingsRequest(ctx, r.client, http.MethodPost, "/v1/organization/spend_alerts/"+data.AlertID.ValueString(), request, &value)
	if err != nil {
		resp.Diagnostics.AddError("Error updating organization spend alert", err.Error())
		return
	}
	invalidateOrganizationSpendAlertsCache(r.client)
	setOrganizationSpendAlertState(&data, &value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationSpendAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationSpendAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := projectSettingsRequest(ctx, r.client, http.MethodDelete, "/v1/organization/spend_alerts/"+data.AlertID.ValueString(), nil, nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Error deleting organization spend alert", err.Error())
		return
	}
	invalidateOrganizationSpendAlertsCache(r.client)
}

func (r *OrganizationSpendAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be a spend alert ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("alert_id"), req.ID)...)
}
