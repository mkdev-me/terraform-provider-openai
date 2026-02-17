package provider

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mkdev-me/terraform-provider-openai/internal/client"
)

// Default API URL if not specified
const defaultAPIURL = "https://api.openai.com"

// IMPORTANT NOTE ABOUT PROJECT API KEYS:
// The openai_project_api_key resource has been removed from this provider
// because OpenAI does not support programmatic creation of project API keys.
// Users must create API keys manually in the OpenAI dashboard.
// We still support reading existing API keys via the data sources:
//   - openai_project_api_key (for a single key)
//   - openai_project_api_keys (for all keys in a project)

// OpenAIClient represents a client for interacting with the OpenAI API.
// It handles authentication and provides methods for making API requests.
type OpenAIClient struct {
	*client.OpenAIClient        // Embed the client package's OpenAIClient
	ProjectAPIKey        string // Store the project API key separately
	AdminAPIKey          string // Store the admin API key separately
}

// GetOpenAIClient extracts the client from the meta interface passed to resource functions
func GetOpenAIClient(m interface{}) (*client.OpenAIClient, error) {
	// Check if the meta is a provider client
	if c, ok := m.(*OpenAIClient); ok {
		log.Printf("[DEBUG] Client is *provider.OpenAIClient")
		return c.OpenAIClient, nil
	}

	// Check if the meta is a client client
	if c, ok := m.(*client.OpenAIClient); ok {
		log.Printf("[DEBUG] Client is *client.OpenAIClient")
		return c, nil
	}

	return client.NewClient("", "", ""), nil
}

// GetOpenAIClientWithProjectKey returns a client configured with the project API key
// This is useful for resources that require project-level API keys (like models)
func GetOpenAIClientWithProjectKey(m interface{}) (*client.OpenAIClient, error) {
	// Check if the meta is a provider client
	if c, ok := m.(*OpenAIClient); ok {
		log.Printf("[DEBUG] Getting client with project API key")
		// If project API key is available, create a new client with it
		if c.ProjectAPIKey != "" {
			log.Printf("[DEBUG] Using project API key for request")
			config := client.ClientConfig{
				APIKey:         c.ProjectAPIKey,
				OrganizationID: c.OpenAIClient.OrganizationID,
				APIURL:         c.OpenAIClient.APIURL,
				Timeout:        c.OpenAIClient.Timeout,
			}
			return client.NewClientWithConfig(config), nil
		}
		// Fall back to the default client if no project key
		log.Printf("[DEBUG] No project API key available, using default client")
		return c.OpenAIClient, nil
	}

	// Check if the meta is a client client
	if c, ok := m.(*client.OpenAIClient); ok {
		log.Printf("[DEBUG] Client is *client.OpenAIClient")
		return c, nil
	}

	return client.NewClient("", "", ""), nil
}

// GetOpenAIClientWithAdminKey returns a client configured with the admin API key
// This is useful for resources that require admin-level API keys (like organization management)
func GetOpenAIClientWithAdminKey(m interface{}) (*client.OpenAIClient, error) {
	// Check if the meta is a provider client
	if c, ok := m.(*OpenAIClient); ok {
		log.Printf("[DEBUG] Getting client with admin API key")
		// If admin API key is available, create a new client with it
		if c.AdminAPIKey != "" {
			log.Printf("[DEBUG] Using admin API key for request")
			config := client.ClientConfig{
				APIKey:         c.AdminAPIKey,
				OrganizationID: c.OpenAIClient.OrganizationID,
				APIURL:         c.OpenAIClient.APIURL,
				Timeout:        c.OpenAIClient.Timeout,
			}
			return client.NewClientWithConfig(config), nil
		}
		// Fall back to the project API key if no admin key
		log.Printf("[DEBUG] No admin API key available, using project API key")
		return c.OpenAIClient, nil
	}

	// Check if the meta is a client client
	if c, ok := m.(*client.OpenAIClient); ok {
		log.Printf("[DEBUG] Client is *client.OpenAIClient")
		return c, nil
	}

	return client.NewClient("", "", ""), nil
}

// Ensure the implementation satisfies the expected interfaces
var _ provider.Provider = &FrameworkProvider{}

// FrameworkProvider defines the provider implementation.
type FrameworkProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// NewFrameworkProvider returns a new provider instance
func NewFrameworkProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FrameworkProvider{
			version: version,
		}
	}
}

func (p *FrameworkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openai"
	resp.Version = p.version
}

func (p *FrameworkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The OpenAI Terraform Provider allows you to manage OpenAI resources using Terraform.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Project API key (sk-proj...) for authentication. Note: Use project keys, not admin keys.",
				Optional:    true,
				Sensitive:   true,
			},
			"admin_key": schema.StringAttribute{
				Description: "The Admin API key for OpenAI administrative operations.",
				Optional:    true,
				Sensitive:   true,
			},
			"organization": schema.StringAttribute{
				Description: "The Organization ID for OpenAI API operations.",
				Optional:    true,
			},
			"api_url": schema.StringAttribute{
				Description: "The URL for OpenAI API. Defaults to https://api.openai.com/v1",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Timeout in seconds for API operations. Defaults to 300.",
				Optional:    true,
			},
		},
	}
}

func (p *FrameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OpenAIProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := data.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	adminKey := data.AdminKey.ValueString()
	if adminKey == "" {
		adminKey = os.Getenv("OPENAI_ADMIN_KEY")
	}

	organization := data.Organization.ValueString()
	if organization == "" {
		organization = os.Getenv("OPENAI_ORGANIZATION")
	}

	apiURL := data.APIURL.ValueString()
	if apiURL == "" {
		apiURL = os.Getenv("OPENAI_API_URL")
		if apiURL == "" {
			apiURL = "https://api.openai.com/v1"
		}
	}

	timeoutVal := data.Timeout.ValueInt64()
	if timeoutVal == 0 {
		if envVal := os.Getenv("OPENAI_TIMEOUT"); envVal != "" {
			if v, err := strconv.ParseInt(envVal, 10, 64); err == nil {
				timeoutVal = v
			}
		}
	}
	if timeoutVal == 0 {
		timeoutVal = 300
	}

	// Create client config
	config := client.ClientConfig{
		APIKey:         apiKey,
		OrganizationID: organization,
		APIURL:         apiURL,
		Timeout:        time.Duration(timeoutVal) * time.Second,
	}

	// Create provider client
	// OpenAIClient struct must be defined in the provider package (e.g. in provider.go)
	providerClient := &OpenAIClient{
		OpenAIClient:  client.NewClientWithConfig(config),
		ProjectAPIKey: apiKey,
		AdminAPIKey:   adminKey,
	}

	resp.DataSourceData = providerClient
	resp.ResourceData = providerClient
}

func (p *FrameworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFileResource,
		NewChatCompletionResource,
		NewVectorStoreResource,
		NewVectorStoreFileResource,
		NewVectorStoreFileBatchResource,
		NewBatchResource,
		NewFineTuningJobResource,
		NewProjectServiceAccountResource,
		NewInviteResource,
		NewProjectResource,
		NewProjectUserResource,
		NewProjectGroupResource,
		NewGroupResource,
		NewGroupUserResource,
		NewOrganizationUserResource,
		NewAdminAPIKeyResource,
		// Role management
		NewOrganizationRoleResource,
		NewProjectRoleResource,
		NewOrganizationGroupRoleResource,
		NewOrganizationUserRoleResource,
		// Batch 6
		NewAudioTranscriptionResource,
		NewAudioTranslationResource,
		NewTextToSpeechResource,
		NewSpeechToTextResource,
		NewImageGenerationResource,
		NewImageEditResource,
		NewImageVariationResource,
		// Batch 7
		NewModelResource,
		NewEmbeddingResource,
		NewModerationResource,
		NewResponseResource,
		NewRateLimitResource,
	}
}

func (p *FrameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewModelDataSource,
		NewModelsDataSource,
		NewFileDataSource,
		NewFilesDataSource,
		NewVectorStoreDataSource,
		NewVectorStoresDataSource,
		// Batch 9: Projects & Org
		NewProjectDataSource,
		NewProjectsDataSource,
		NewProjectAPIKeyDataSource,
		NewProjectAPIKeysDataSource,
		NewProjectServiceAccountDataSource,
		NewProjectServiceAccountsDataSource,
		NewProjectUserDataSource,
		NewProjectUsersDataSource,
		NewProjectGroupDataSource,
		NewProjectGroupsDataSource,
		NewRoleDataSource,
		NewRolesDataSource,
		NewGroupRolesDataSource,
		NewProjectRoleDataSource,
		NewProjectRolesDataSource,
		NewProjectGroupRolesDataSource,
		NewOrganizationUserRolesDataSource,
		NewProjectUserRolesDataSource,
		NewGroupDataSource,
		NewGroupsDataSource,
		NewGroupUserDataSource,
		NewGroupUsersDataSource,
		NewOrganizationUserDataSource,
		NewOrganizationUsersDataSource,
		NewAdminAPIKeyDataSource,
		NewAdminAPIKeysDataSource,
		NewInviteDataSource,
		NewInvitesDataSource,
		// Batch 9: Audio
		NewAudioTranscriptionDataSource,
		NewAudioTranscriptionsDataSource,
		NewAudioTranslationDataSource,
		NewAudioTranslationsDataSource,
		NewSpeechToTextDataSource,
		NewSpeechToTextsDataSource,
		NewTextToSpeechDataSource,
		NewTextToSpeechsDataSource,
		// Batch 9: Batch & Fine-Tuning
		NewBatchDataSource,
		NewBatchesDataSource,
		NewFineTuningJobDataSource,
		NewFineTuningJobsDataSource,
		// Batch 9: Chat & Model
		NewChatCompletionDataSource,
		NewChatCompletionsDataSource,
		NewChatCompletionMessagesDataSource,

		// Batch 9: Vector Store Utils
		NewVectorStoreFileDataSource,
		NewVectorStoreFileBatchDataSource,
		NewVectorStoreFileContentDataSource,
		NewVectorStoreFilesDataSource,
		NewVectorStoreFileBatchFilesDataSource,
	}
}

type OpenAIProviderModel struct {
	APIKey       types.String `tfsdk:"api_key"`
	AdminKey     types.String `tfsdk:"admin_key"`
	Organization types.String `tfsdk:"organization"`
	APIURL       types.String `tfsdk:"api_url"`
	Timeout      types.Int64  `tfsdk:"timeout"`
}
