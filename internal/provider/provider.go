package provider

import (
	"context"
	"log"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			return client.NewClient(c.ProjectAPIKey, c.OpenAIClient.OrganizationID, c.OpenAIClient.APIURL), nil
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
			return client.NewClient(c.AdminAPIKey, c.OpenAIClient.OrganizationID, c.OpenAIClient.APIURL), nil
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

// Provider returns a terraform.ResourceProvider that implements the OpenAI provider.
// This provider allows Terraform to manage OpenAI resources including models, assistants,
// files, and other OpenAI API resources.
func Provider() *schema.Provider {
	log.Printf("[DEBUG] Starting provider initialization")

	// Create a map to store resources
	log.Printf("[DEBUG] Creating resource map")

	// Create each resource with debug logging
	log.Printf("[DEBUG] Creating openai_file resource")
	fileResource := resourceOpenAIFile()
	log.Printf("[DEBUG] openai_file schema: %+v", fileResource.Schema)

	log.Printf("[DEBUG] Creating openai_model resource")
	modelResource := resourceOpenAIModel()
	log.Printf("[DEBUG] openai_model schema: %+v", modelResource.Schema)

	log.Printf("[DEBUG] Creating openai_project resource")
	projectResource := resourceOpenAIProject()
	log.Printf("[DEBUG] openai_project schema: %+v", projectResource.Schema)
	log.Printf("[DEBUG] Project CRUD handlers: Create=%p, Read=%p, Update=%p, Delete=%p",
		projectResource.CreateContext, projectResource.ReadContext,
		projectResource.UpdateContext, projectResource.DeleteContext)

	log.Printf("[DEBUG] Creating openai_text_to_speech resource")
	ttsResource := resourceOpenAITextToSpeech()
	log.Printf("[DEBUG] openai_text_to_speech schema: %+v", ttsResource.Schema)

	log.Printf("[DEBUG] Creating openai_audio_transcription resource")
	transcriptionResource := resourceOpenAIAudioTranscription()
	log.Printf("[DEBUG] openai_audio_transcription schema: %+v", transcriptionResource.Schema)

	log.Printf("[DEBUG] Creating openai_audio_translation resource")
	translationResource := resourceOpenAIAudioTranslation()
	log.Printf("[DEBUG] openai_audio_translation schema: %+v", translationResource.Schema)

	log.Printf("[DEBUG] Creating openai_speech_to_text resource")
	speechToTextResource := resourceOpenAISpeechToText()
	log.Printf("[DEBUG] openai_speech_to_text schema: %+v", speechToTextResource.Schema)

	log.Printf("[DEBUG] Creating openai_embedding resource")
	embeddingResource := resourceOpenAIEmbedding()
	log.Printf("[DEBUG] openai_embedding schema: %+v", embeddingResource.Schema)

	log.Printf("[DEBUG] Creating openai_fine_tuned_model resource")
	fineTunedModelResource := resourceOpenAIFineTunedModel()
	log.Printf("[DEBUG] openai_fine_tuned_model schema: %+v", fineTunedModelResource.Schema)

	log.Printf("[DEBUG] Creating openai_fine_tuning_job resource")
	fineTuningJobResource := resourceOpenAIFineTuningJob()
	log.Printf("[DEBUG] openai_fine_tuning_job schema: %+v", fineTuningJobResource.Schema)

	log.Printf("[DEBUG] Creating openai_fine_tuning_checkpoint_permission resource")
	fineTuningCheckpointPermissionResource := resourceOpenAIFineTuningCheckpointPermission()
	log.Printf("[DEBUG] openai_fine_tuning_checkpoint_permission schema: %+v", fineTuningCheckpointPermissionResource.Schema)

	log.Printf("[DEBUG] Creating openai_image_generation resource")
	imageGenerationResource := resourceOpenAIImageGeneration()
	log.Printf("[DEBUG] openai_image_generation schema: %+v", imageGenerationResource.Schema)

	log.Printf("[DEBUG] Creating openai_image_edit resource")
	imageEditResource := resourceOpenAIImageEdit()
	log.Printf("[DEBUG] openai_image_edit schema: %+v", imageEditResource.Schema)

	log.Printf("[DEBUG] Creating openai_image_variation resource")
	imageVariationResource := resourceOpenAIImageVariation()
	log.Printf("[DEBUG] openai_image_variation schema: %+v", imageVariationResource.Schema)

	log.Printf("[DEBUG] Creating openai_moderation resource")
	moderationResource := resourceOpenAIModeration()
	log.Printf("[DEBUG] openai_moderation schema: %+v", moderationResource.Schema)

	log.Printf("[DEBUG] Creating openai_completion resource")
	completionResource := resourceOpenAICompletion()
	log.Printf("[DEBUG] openai_completion schema: %+v", completionResource.Schema)

	log.Printf("[DEBUG] Creating openai_chat_completion resource")
	chatCompletionResource := resourceOpenAIChatCompletion()
	log.Printf("[DEBUG] openai_chat_completion schema: %+v", chatCompletionResource.Schema)

	log.Printf("[DEBUG] Creating openai_edit resource")
	editResource := resourceOpenAIEdit()
	log.Printf("[DEBUG] openai_edit schema: %+v", editResource.Schema)

	log.Printf("[DEBUG] Creating openai_assistant resource")
	assistantResource := resourceOpenAIAssistant()
	log.Printf("[DEBUG] openai_assistant schema: %+v", assistantResource.Schema)

	log.Printf("[DEBUG] Creating openai_thread resource")
	threadResource := resourceOpenAIThread()
	log.Printf("[DEBUG] openai_thread schema: %+v", threadResource.Schema)

	log.Printf("[DEBUG] Creating openai_message resource")
	messageResource := resourceOpenAIMessage()
	log.Printf("[DEBUG] openai_message schema: %+v", messageResource.Schema)

	log.Printf("[DEBUG] Creating openai_run resource")
	runResource := resourceOpenAIRun()
	log.Printf("[DEBUG] openai_run schema: %+v", runResource.Schema)

	log.Printf("[DEBUG] Creating openai_thread_run resource")
	threadRunResource := resourceOpenAIThreadRun()
	log.Printf("[DEBUG] openai_thread_run schema: %+v", threadRunResource.Schema)

	log.Printf("[DEBUG] Creating openai_batch resource")
	batchResource := resourceOpenAIBatch()
	log.Printf("[DEBUG] openai_batch schema: %+v", batchResource.Schema)

	log.Printf("[DEBUG] Creating openai_rate_limit resource")
	rateLimitResource := resourceOpenAIRateLimit()
	log.Printf("[DEBUG] openai_rate_limit schema: %+v", rateLimitResource.Schema)

	log.Printf("[DEBUG] Creating openai_project_service_account resource")
	projectServiceAccountResource := resourceOpenAIProjectServiceAccount()
	log.Printf("[DEBUG] openai_project_service_account schema: %+v", projectServiceAccountResource.Schema)

	log.Printf("[DEBUG] Creating openai_project_user resource")
	projectUserResource := resourceOpenAIProjectUser()
	log.Printf("[DEBUG] openai_project_user schema: %+v", projectUserResource.Schema)

	log.Printf("[DEBUG] Creating openai_invite resource")
	inviteResource := resourceOpenAIInvite()
	log.Printf("[DEBUG] openai_invite schema: %+v", inviteResource.Schema)

	log.Printf("[DEBUG] Creating openai_admin_api_key resource")
	adminAPIKeyResource := resourceOpenAIAdminAPIKey()
	log.Printf("[DEBUG] openai_admin_api_key schema: %+v", adminAPIKeyResource.Schema)

	log.Printf("[DEBUG] Creating openai_model_response resource")
	modelResponseResource := resourceOpenAIModelResponse()
	log.Printf("[DEBUG] openai_model_response schema: %+v", modelResponseResource.Schema)

	// Add Vector Store resources
	log.Printf("[DEBUG] Creating openai_vector_store resource")
	vectorStoreResource := resourceOpenAIVectorStore()
	log.Printf("[DEBUG] openai_vector_store schema: %+v", vectorStoreResource.Schema)

	log.Printf("[DEBUG] Creating openai_vector_store_file resource")
	vectorStoreFileResource := resourceOpenAIVectorStoreFile()
	log.Printf("[DEBUG] openai_vector_store_file schema: %+v", vectorStoreFileResource.Schema)

	log.Printf("[DEBUG] Creating openai_vector_store_file_batch resource")
	vectorStoreFileBatchResource := resourceOpenAIVectorStoreFileBatch()
	log.Printf("[DEBUG] openai_vector_store_file_batch schema: %+v", vectorStoreFileBatchResource.Schema)

	log.Printf("[DEBUG] Creating openai_organization_user resource")
	organizationUserResource := resourceOpenAIOrganizationUser()
	log.Printf("[DEBUG] openai_organization_user schema: %+v", organizationUserResource.Schema)

	modelResponseDataSource := dataSourceOpenAIModelResponse()
	log.Printf("[DEBUG] openai_model_response data source schema: %+v", modelResponseDataSource.Schema)

	modelResponsesDataSource := dataSourceOpenAIModelResponses()
	log.Printf("[DEBUG] openai_model_responses data source schema: %+v", modelResponsesDataSource.Schema)

	modelResponseInputItemsDataSource := dataSourceOpenAIModelResponseInputItems()
	log.Printf("[DEBUG] openai_model_response_input_items data source schema: %+v", modelResponseInputItemsDataSource.Schema)

	resourceMap := map[string]*schema.Resource{
		"openai_file":                              fileResource,
		"openai_model":                             modelResource,
		"openai_project":                           projectResource,
		"openai_text_to_speech":                    ttsResource,
		"openai_audio_transcription":               transcriptionResource,
		"openai_audio_translation":                 translationResource,
		"openai_speech_to_text":                    speechToTextResource,
		"openai_embedding":                         embeddingResource,
		"openai_fine_tuned_model":                  fineTunedModelResource,
		"openai_fine_tuning_job":                   fineTuningJobResource,
		"openai_fine_tuning_checkpoint_permission": fineTuningCheckpointPermissionResource,
		"openai_image_generation":                  imageGenerationResource,
		"openai_image_edit":                        imageEditResource,
		"openai_image_variation":                   imageVariationResource,
		"openai_moderation":                        moderationResource,
		"openai_completion":                        completionResource,
		"openai_chat_completion":                   chatCompletionResource,
		"openai_edit":                              editResource,
		"openai_assistant":                         assistantResource,
		"openai_thread":                            threadResource,
		"openai_message":                           messageResource,
		"openai_run":                               runResource,
		"openai_thread_run":                        threadRunResource,
		"openai_batch":                             batchResource,
		"openai_rate_limit":                        rateLimitResource,
		"openai_project_service_account":           projectServiceAccountResource,
		"openai_project_user":                      projectUserResource,
		"openai_invite":                            inviteResource,
		"openai_admin_api_key":                     adminAPIKeyResource,
		"openai_model_response":                    modelResponseResource,
		"openai_vector_store":                      vectorStoreResource,
		"openai_vector_store_file":                 vectorStoreFileResource,
		"openai_vector_store_file_batch":           vectorStoreFileBatchResource,
		"openai_organization_user":                 organizationUserResource,
	}

	// Debug log each registered resource
	log.Printf("[DEBUG] Number of resources registered: %d", len(resourceMap))
	log.Printf("[DEBUG] Note: Project API key creation resource has been intentionally removed as OpenAI doesn't support programmatic API key creation")
	for resourceName, resource := range resourceMap {
		log.Printf("[DEBUG] Resource: %s, Schema fields: %d", resourceName, len(resource.Schema))
	}

	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Project API key (sk-proj...) for authentication. Note: Use project keys, not admin keys.",
				DefaultFunc: schema.EnvDefaultFunc("OPENAI_API_KEY", nil),
			},
			"admin_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("OPENAI_ADMIN_KEY", nil),
				Description: "The Admin API key for OpenAI administrative operations. If not set, the OPENAI_ADMIN_KEY environment variable will be used.",
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OPENAI_ORGANIZATION", nil),
				Description: "The Organization ID for OpenAI API operations. If not set, the OPENAI_ORGANIZATION environment variable will be used.",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OPENAI_API_URL", "https://api.openai.com/v1"),
				Description: "The URL for OpenAI API. If not set, the OPENAI_API_URL environment variable will be used, or the default value of 'https://api.openai.com/v1'.",
			},
		},
		ResourcesMap: resourceMap,
		DataSourcesMap: map[string]*schema.Resource{
			"openai_file":                               dataSourceOpenAIFile(),
			"openai_files":                              dataSourceOpenAIFiles(),
			"openai_models":                             dataSourceOpenAIModels(),
			"openai_model":                              dataSourceOpenAIModel(),
			"openai_assistants":                         dataSourceOpenAIAssistants(),
			"openai_assistant":                          dataSourceOpenAIAssistant(),
			"openai_thread":                             dataSourceOpenAIThread(),
			"openai_message":                            dataSourceOpenAIMessage(),
			"openai_messages":                           dataSourceOpenAIMessages(),
			"openai_run":                                dataSourceOpenAIRun(),
			"openai_runs":                               dataSourceOpenAIRuns(),
			"openai_invite":                             dataSourceOpenAIInvite(),
			"openai_invites":                            dataSourceOpenAIInvites(),
			"openai_project_user":                       dataSourceOpenAIProjectUser(),
			"openai_project_users":                      dataSourceOpenAIProjectUsers(),
			"openai_project":                            dataSourceOpenAIProject(),
			"openai_projects":                           dataSourceOpenAIProjects(),
			"openai_project_api_key":                    dataSourceOpenAIProjectAPIKey(),
			"openai_project_api_keys":                   dataSourceOpenAIProjectAPIKeys(),
			"openai_rate_limit":                         dataSourceOpenAIRateLimit(),
			"openai_rate_limits":                        dataSourceOpenAIRateLimits(),
			"openai_project_service_account":            dataSourceOpenAIProjectServiceAccount(),
			"openai_project_service_accounts":           dataSourceOpenAIProjectServiceAccounts(),
			"openai_text_to_speech":                     dataSourceOpenAITextToSpeech(),
			"openai_text_to_speechs":                    dataSourceOpenAITextToSpeechs(),
			"openai_speech_to_text":                     dataSourceOpenAISpeechToText(),
			"openai_speech_to_texts":                    dataSourceOpenAISpeechToTexts(),
			"openai_audio_transcription":                dataSourceOpenAIAudioTranscription(),
			"openai_audio_transcriptions":               dataSourceOpenAIAudioTranscriptions(),
			"openai_audio_translation":                  dataSourceOpenAIAudioTranslation(),
			"openai_audio_translations":                 dataSourceOpenAIAudioTranslations(),
			"openai_admin_api_key":                      dataSourceOpenAIAdminAPIKey(),
			"openai_admin_api_keys":                     dataSourceOpenAIAdminAPIKeys(),
			"openai_batch":                              dataSourceOpenAIBatch(),
			"openai_batches":                            dataSourceOpenAIBatches(),
			"openai_fine_tuning_job":                    dataSourceOpenAIFineTuningJob(),
			"openai_fine_tuning_jobs":                   dataSourceOpenAIFineTuningJobs(),
			"openai_fine_tuning_checkpoints":            dataSourceOpenAIFineTuningCheckpoints(),
			"openai_fine_tuning_events":                 dataSourceOpenAIFineTuningEvents(),
			"openai_fine_tuning_checkpoint_permissions": dataSourceOpenAIFineTuningCheckpointPermissions(),
			"openai_chat_completion":                    dataSourceOpenAIChatCompletion(),
			"openai_chat_completion_messages":           dataSourceOpenAIChatCompletionMessages(),
			"openai_chat_completions":                   dataSourceOpenAIChatCompletions(),
			"openai_model_response":                     dataSourceOpenAIModelResponse(),
			"openai_model_responses":                    dataSourceOpenAIModelResponses(),
			"openai_model_response_input_items":         dataSourceOpenAIModelResponseInputItems(),
			"openai_vector_store":                       dataSourceOpenAIVectorStore(),
			"openai_vector_stores":                      dataSourceOpenAIVectorStores(),
			"openai_vector_store_file":                  dataSourceOpenAIVectorStoreFile(),
			"openai_vector_store_file_batch":            dataSourceOpenAIVectorStoreFileBatch(),
			"openai_vector_store_files":                 dataSourceOpenAIVectorStoreFiles(),
			"openai_vector_store_file_content":          dataSourceOpenAIVectorStoreFileContent(),
			"openai_vector_store_file_batch_files":      dataSourceOpenAIVectorStoreFileBatchFiles(),
			"openai_thread_run":                         dataSourceOpenAIThreadRun(),
			"openai_organization_users":                 dataSourceOpenAIOrganizationUsers(),
			"openai_organization_user":                  dataSourceOpenAIOrganizationUser(),
		},
		ConfigureContextFunc: providerConfigure,
	}

	log.Printf("[DEBUG] Provider initialization complete")
	return provider
}

// providerConfigure configures the provider with the necessary clients and connections.
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	adminKey := d.Get("admin_key").(string)
	organization := d.Get("organization").(string)
	apiURL := d.Get("api_url").(string)

	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	log.Printf("[DEBUG] Configuring provider with base URL: %s", apiURL)
	log.Printf("[DEBUG] Organization ID: %s", organization)

	// Validate base URL
	baseURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, diag.Errorf("invalid API URL: %v", err)
	}

	// Initialize OpenAI client with project API key by default
	// The embedded client should use the project API key for standard operations
	client := &OpenAIClient{
		OpenAIClient:  client.NewClient(apiKey, organization, baseURL.String()),
		ProjectAPIKey: apiKey,
		AdminAPIKey:   adminKey,
	}

	return client, nil
}
