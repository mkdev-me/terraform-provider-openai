package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GroupDataSource - singular data source

var _ datasource.DataSource = &GroupDataSource{}

func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

type GroupDataSource struct {
	client *OpenAIClient
}

type GroupDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	IsSCIMManaged types.Bool   `tfsdk:"is_scim_managed"`
}

func (d *GroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *GroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific group in your OpenAI organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the group to retrieve. Either id or name must be provided.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the group to search for. Either id or name must be provided.",
				Optional:    true,
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "Unix timestamp (in seconds) when the group was created.",
				Computed:    true,
			},
			"is_scim_managed": schema.BoolAttribute{
				Description: "Whether the group is managed through SCIM and controlled by your identity provider.",
				Computed:    true,
			},
		},
	}
}

func (d *GroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *OpenAIClient, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := data.ID.ValueString()
	groupName := data.Name.ValueString()

	if groupID == "" && groupName == "" {
		resp.Diagnostics.AddError(
			"Missing Group Identifier",
			"Either id or name must be provided.",
		)
		return
	}

	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key (admin_key) to read groups.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups"
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups"
	}

	var foundGroup *GroupResponseFramework
	cursor := ""

	for foundGroup == nil {
		parsedURL, _ := url.Parse(reqURL)
		q := parsedURL.Query()
		q.Set("limit", "100")
		if cursor != "" {
			q.Set("after", cursor)
		}
		parsedURL.RawQuery = q.Encode()

		httpRequest, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			resp.Diagnostics.AddError("Error creating request", err.Error())
			return
		}

		httpRequest.Header.Set("Authorization", "Bearer "+adminKey)
		httpRequest.Header.Set("Content-Type", "application/json")
		if d.client.OpenAIClient.OrganizationID != "" {
			httpRequest.Header.Set("OpenAI-Organization", d.client.OpenAIClient.OrganizationID)
		}

		httpClient := &http.Client{}
		httpResp, err := httpClient.Do(httpRequest)
		if err != nil {
			resp.Diagnostics.AddError("Error executing request", err.Error())
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != 200 {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
			return
		}

		var listResp GroupListResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}

		for i := range listResp.Data {
			group := listResp.Data[i]
			if groupID != "" && group.ID == groupID {
				foundGroup = &group
				break
			}
			if groupName != "" && strings.EqualFold(group.Name, groupName) {
				foundGroup = &group
				break
			}
		}

		if foundGroup != nil {
			break
		}

		if !listResp.HasMore {
			break
		}
		if len(listResp.Data) > 0 {
			cursor = listResp.Data[len(listResp.Data)-1].ID
		} else {
			break
		}
	}

	if foundGroup == nil {
		identifier := groupID
		if identifier == "" {
			identifier = groupName
		}
		resp.Diagnostics.AddError(
			"Group Not Found",
			fmt.Sprintf("Group %s not found in organization", identifier),
		)
		return
	}

	data.ID = types.StringValue(foundGroup.ID)
	data.Name = types.StringValue(foundGroup.Name)
	data.CreatedAt = types.Int64Value(foundGroup.CreatedAt)
	data.IsSCIMManaged = types.BoolValue(foundGroup.IsSCIMManaged)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// GroupsDataSource - list data source

func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

type GroupsDataSource struct {
	client *OpenAIClient
}

type GroupsDataSourceModel struct {
	ID         types.String       `tfsdk:"id"`
	Groups     []GroupResultModel `tfsdk:"groups"`
	GroupIDs   []types.String     `tfsdk:"group_ids"`
	GroupCount types.Int64        `tfsdk:"group_count"`
}

type GroupResultModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	IsSCIMManaged types.Bool   `tfsdk:"is_scim_managed"`
}

func (d *GroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *GroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all groups in your OpenAI organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"groups": schema.ListNestedAttribute{
				Description: "List of groups in the organization.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the group.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the group.",
							Computed:    true,
						},
						"created_at": schema.Int64Attribute{
							Description: "Unix timestamp (in seconds) when the group was created.",
							Computed:    true,
						},
						"is_scim_managed": schema.BoolAttribute{
							Description: "Whether the group is managed through SCIM.",
							Computed:    true,
						},
					},
				},
			},
			"group_ids": schema.ListAttribute{
				Description: "List of group IDs in the organization.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"group_count": schema.Int64Attribute{
				Description: "Number of groups in the organization.",
				Computed:    true,
			},
		},
	}
}

func (d *GroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*OpenAIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *OpenAIClient, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *GroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API Key is required.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	var reqURL string
	if strings.Contains(apiURL, "/v1") {
		reqURL = strings.TrimSuffix(apiURL, "/v1") + "/v1/organization/groups"
	} else {
		reqURL = strings.TrimSuffix(apiURL, "/") + "/v1/organization/groups"
	}

	var allGroups []GroupResultModel
	var groupIDs []string

	cursor := ""
	for {
		parsedURL, _ := url.Parse(reqURL)
		q := parsedURL.Query()
		q.Set("limit", "100")
		if cursor != "" {
			q.Set("after", cursor)
		}
		parsedURL.RawQuery = q.Encode()

		httpRequest, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			resp.Diagnostics.AddError("Error creating request", err.Error())
			return
		}
		httpRequest.Header.Set("Authorization", "Bearer "+adminKey)
		httpRequest.Header.Set("Content-Type", "application/json")
		if d.client.OpenAIClient.OrganizationID != "" {
			httpRequest.Header.Set("OpenAI-Organization", d.client.OpenAIClient.OrganizationID)
		}

		httpClient := &http.Client{}
		httpResp, err := httpClient.Do(httpRequest)
		if err != nil {
			resp.Diagnostics.AddError("Error executing request", err.Error())
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != 200 {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
			return
		}

		var listResp GroupListResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}

		for _, g := range listResp.Data {
			groupModel := GroupResultModel{
				ID:            types.StringValue(g.ID),
				Name:          types.StringValue(g.Name),
				CreatedAt:     types.Int64Value(g.CreatedAt),
				IsSCIMManaged: types.BoolValue(g.IsSCIMManaged),
			}
			allGroups = append(allGroups, groupModel)
			groupIDs = append(groupIDs, g.ID)
		}

		if !listResp.HasMore {
			break
		}
		if len(listResp.Data) > 0 {
			cursor = listResp.Data[len(listResp.Data)-1].ID
		} else {
			break
		}
	}

	data.ID = types.StringValue("organization-groups")
	data.Groups = allGroups
	data.GroupCount = types.Int64Value(int64(len(allGroups)))

	data.GroupIDs = make([]types.String, len(groupIDs))
	for i, v := range groupIDs {
		data.GroupIDs[i] = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
