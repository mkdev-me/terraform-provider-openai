package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProjectGroupDataSource - singular data source

var _ datasource.DataSource = &ProjectGroupDataSource{}

func NewProjectGroupDataSource() datasource.DataSource {
	return &ProjectGroupDataSource{}
}

type ProjectGroupDataSource struct {
	client *OpenAIClient
}

type ProjectGroupDataSourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	GroupID   types.String `tfsdk:"group_id"`
	GroupName types.String `tfsdk:"group_name"`
	ID        types.String `tfsdk:"id"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func (d *ProjectGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_group"
}

func (d *ProjectGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific group in an OpenAI project. Use the openai_project_group_roles data source to get role assignments.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to retrieve the group from.",
				Required:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of the group to retrieve.",
				Optional:    true,
			},
			"group_name": schema.StringAttribute{
				Description: "The name of the group to search for. Used if group_id is not provided.",
				Optional:    true,
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the resource (composite of project_id:group_id).",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "The Unix timestamp (in seconds) when the group was added to the project.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*OpenAIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *OpenAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ProjectGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectGroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	groupID := data.GroupID.ValueString()
	groupName := data.GroupName.ValueString()

	if groupID == "" && groupName == "" {
		resp.Diagnostics.AddError(
			"Missing Group Identifier",
			"Either group_id or group_name must be provided.",
		)
		return
	}

	// We need Admin Key
	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"The provider must be configured with an Admin API Key (admin_key) to read project groups.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	suffix := fmt.Sprintf("/organization/projects/%s/groups", projectID)

	// Safely construct URL by trimming both /v1 and trailing /
	baseURL := strings.TrimSuffix(apiURL, "/v1")
	baseURL = strings.TrimSuffix(baseURL, "/")
	reqURL := baseURL + "/v1" + suffix

	var foundGroup *ProjectGroupResponseFramework
	cursor := ""

	// Loop until found or done
	for foundGroup == nil {
		parsedURL, err := url.Parse(reqURL)
		if err != nil {
			resp.Diagnostics.AddError("Error parsing URL", err.Error())
			return
		}
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

		httpClient := &http.Client{Timeout: 30 * time.Second}
		httpResp, err := httpClient.Do(httpRequest)
		if err != nil {
			resp.Diagnostics.AddError("Error executing request", err.Error())
			return
		}

		if httpResp.StatusCode != 200 {
			httpResp.Body.Close()
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
			return
		}

		var listResp ProjectGroupListResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			httpResp.Body.Close()
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}
		httpResp.Body.Close()

		for i := range listResp.Data {
			group := listResp.Data[i]
			// If groupID is provided, only match by ID (don't fall back to name)
			if groupID != "" {
				if group.GroupID == groupID {
					foundGroup = &group
					break
				}
			} else if groupName != "" {
				// Only search by name if groupID was not provided
				if strings.EqualFold(group.GroupName, groupName) {
					foundGroup = &group
					break
				}
			}
		}

		if foundGroup != nil {
			break
		}

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		cursor = *listResp.Next
	}

	if foundGroup == nil {
		identifier := groupID
		if identifier == "" {
			identifier = groupName
		}
		resp.Diagnostics.AddError(
			"Group Not Found",
			fmt.Sprintf("Group %s not found in project %s", identifier, projectID),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", projectID, foundGroup.GroupID))
	data.GroupID = types.StringValue(foundGroup.GroupID)
	data.GroupName = types.StringValue(foundGroup.GroupName)
	data.CreatedAt = types.Int64Value(foundGroup.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ProjectGroupsDataSource - list data source

func NewProjectGroupsDataSource() datasource.DataSource {
	return &ProjectGroupsDataSource{}
}

type ProjectGroupsDataSource struct {
	client *OpenAIClient
}

type ProjectGroupsDataSourceModel struct {
	ProjectID  types.String              `tfsdk:"project_id"`
	Groups     []ProjectGroupResultModel `tfsdk:"groups"`
	GroupIDs   []types.String            `tfsdk:"group_ids"`
	GroupCount types.Int64               `tfsdk:"group_count"`
	ID         types.String              `tfsdk:"id"`
}

type ProjectGroupResultModel struct {
	GroupID   types.String `tfsdk:"group_id"`
	GroupName types.String `tfsdk:"group_name"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func (d *ProjectGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_groups"
}

func (d *ProjectGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of all groups in a specific OpenAI project. Use the openai_project_group_roles data source to get role assignments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to retrieve groups from.",
				Required:    true,
			},
			"groups": schema.ListNestedAttribute{
				Description: "List of groups in the project.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group_id": schema.StringAttribute{
							Description: "The ID of the group.",
							Computed:    true,
						},
						"group_name": schema.StringAttribute{
							Description: "The display name of the group.",
							Computed:    true,
						},
						"created_at": schema.Int64Attribute{
							Description: "The Unix timestamp (in seconds) when the group was added to the project.",
							Computed:    true,
						},
					},
				},
			},
			"group_ids": schema.ListAttribute{
				Description: "List of group IDs in the project.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"group_count": schema.Int64Attribute{
				Description: "Number of groups in the project.",
				Computed:    true,
			},
		},
	}
}

func (d *ProjectGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*OpenAIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *OpenAIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ProjectGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectGroupsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := data.ProjectID.ValueString()
	adminKey := d.client.AdminAPIKey
	if adminKey == "" {
		resp.Diagnostics.AddError(
			"Missing Admin API Key",
			"Admin API Key is required.",
		)
		return
	}

	apiURL := d.client.OpenAIClient.APIURL
	suffix := fmt.Sprintf("/organization/projects/%s/groups", projectID)

	// Safely construct URL by trimming both /v1 and trailing /
	baseURL := strings.TrimSuffix(apiURL, "/v1")
	baseURL = strings.TrimSuffix(baseURL, "/")
	reqURL := baseURL + "/v1" + suffix

	// Initialize slices to empty to avoid nil (which becomes null in state)
	allGroups := make([]ProjectGroupResultModel, 0)
	groupIDs := make([]string, 0)

	cursor := ""
	for {
		parsedURL, err := url.Parse(reqURL)
		if err != nil {
			resp.Diagnostics.AddError("Error parsing URL", err.Error())
			return
		}
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

		httpClient := &http.Client{Timeout: 30 * time.Second}
		httpResp, err := httpClient.Do(httpRequest)
		if err != nil {
			resp.Diagnostics.AddError("Error executing request", err.Error())
			return
		}

		if httpResp.StatusCode != 200 {
			httpResp.Body.Close()
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Status: %s", httpResp.Status))
			return
		}

		var listResp ProjectGroupListResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&listResp); err != nil {
			httpResp.Body.Close()
			resp.Diagnostics.AddError("Error decoding response", err.Error())
			return
		}
		httpResp.Body.Close()

		for _, g := range listResp.Data {
			groupModel := ProjectGroupResultModel{
				GroupID:   types.StringValue(g.GroupID),
				GroupName: types.StringValue(g.GroupName),
				CreatedAt: types.Int64Value(g.CreatedAt),
			}
			allGroups = append(allGroups, groupModel)
			groupIDs = append(groupIDs, g.GroupID)
		}

		if !listResp.HasMore || listResp.Next == nil {
			break
		}
		cursor = *listResp.Next
	}

	data.ID = types.StringValue(projectID)
	data.Groups = allGroups
	data.GroupCount = types.Int64Value(int64(len(allGroups)))

	data.GroupIDs = make([]types.String, len(groupIDs))
	for i, v := range groupIDs {
		data.GroupIDs[i] = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
