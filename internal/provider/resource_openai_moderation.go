package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ModerationResponse represents the API response for content moderation.
// It contains the moderation results and metadata about the moderation process.
// This structure provides comprehensive information about content safety and compliance.
type ModerationResponse struct {
	ID      string             `json:"id"`      // Unique identifier for the moderation request
	Model   string             `json:"model"`   // Model used for moderation
	Results []ModerationResult `json:"results"` // List of moderation results
}

// ModerationResult represents the moderation analysis for a single input.
// It contains detailed information about flagged content and category scores.
type ModerationResult struct {
	Flagged        bool                     `json:"flagged"`         // Whether the content was flagged
	Categories     ModerationCategories     `json:"categories"`      // Categories of detected content
	CategoryScores ModerationCategoryScores `json:"category_scores"` // Confidence scores for each category
}

// ModerationCategories represents the categories of content detected during moderation.
// Each field indicates whether content of that category was detected in the input.
type ModerationCategories struct {
	Sexual                bool `json:"sexual"`                 // Sexual content
	Hate                  bool `json:"hate"`                   // Hate speech
	Harassment            bool `json:"harassment"`             // Harassing content
	SelfHarm              bool `json:"self-harm"`              // Self-harm content
	SexualMinors          bool `json:"sexual/minors"`          // Sexual content involving minors
	HateThreatening       bool `json:"hate/threatening"`       // Threatening hate speech
	ViolenceGraphic       bool `json:"violence/graphic"`       // Graphic violence
	SelfHarmIntent        bool `json:"self-harm/intent"`       // Intent to self-harm
	SelfHarmInstructions  bool `json:"self-harm/instructions"` // Instructions for self-harm
	HarassmentThreatening bool `json:"harassment/threatening"` // Threatening harassment
	Violence              bool `json:"violence"`               // General violence
}

// ModerationCategoryScores represents the confidence scores for each moderation category.
// Each field contains a float value between 0 and 1 indicating the confidence level.
type ModerationCategoryScores struct {
	Sexual                float64 `json:"sexual"`                 // Score for sexual content
	Hate                  float64 `json:"hate"`                   // Score for hate speech
	Harassment            float64 `json:"harassment"`             // Score for harassment
	SelfHarm              float64 `json:"self-harm"`              // Score for self-harm
	SexualMinors          float64 `json:"sexual/minors"`          // Score for sexual content involving minors
	HateThreatening       float64 `json:"hate/threatening"`       // Score for threatening hate speech
	ViolenceGraphic       float64 `json:"violence/graphic"`       // Score for graphic violence
	SelfHarmIntent        float64 `json:"self-harm/intent"`       // Score for intent to self-harm
	SelfHarmInstructions  float64 `json:"self-harm/instructions"` // Score for self-harm instructions
	HarassmentThreatening float64 `json:"harassment/threatening"` // Score for threatening harassment
	Violence              float64 `json:"violence"`               // Score for general violence
}

// ModerationRequest represents the request payload for content moderation.
// It specifies the input text to moderate and optionally the model to use.
type ModerationRequest struct {
	Input string `json:"input"`           // Text content to moderate
	Model string `json:"model,omitempty"` // Optional model to use for moderation
}

// resourceOpenAIModeration defines the schema and CRUD operations for OpenAI content moderation.
// This resource allows users to analyze text content for safety and compliance using OpenAI's models.
// It provides comprehensive moderation capabilities with detailed category analysis.
func resourceOpenAIModeration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIModerationCreate,
		ReadContext:   resourceOpenAIModerationRead,
		DeleteContext: resourceOpenAIModerationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customizeDiffModeration,
		Schema: map[string]*schema.Schema{
			"input": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Input (or inputs) to classify. Can be a single string, an array of strings, or an array of multi-modal input objects.",
			},
			"model": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The content moderation model you would like to use.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the moderation result.",
			},
			"results": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The full results of the moderation API call.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flagged": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the content violates OpenAI's usage policies.",
						},
						"categories": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Map of category names to boolean values indicating if the content violates that category.",
							Elem: &schema.Schema{
								Type: schema.TypeBool,
							},
						},
						"category_scores": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Map of category names to scores (0.0 to 1.0) indicating the confidence of the model that the content violates that category.",
							Elem: &schema.Schema{
								Type: schema.TypeFloat,
							},
						},
						"category_applied_input_types": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Map indicating which input types (text, image) each category applies to.",
							Elem: &schema.Schema{
								Type: schema.TypeList,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
			},
			"flagged": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the content violates OpenAI's usage policies (first result if multiple inputs).",
			},
			"categories": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Map of category names to boolean values indicating if the content violates that category (first result if multiple inputs).",
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
			"category_scores": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Map of category names to scores (0.0 to 1.0) indicating the confidence of the model (first result if multiple inputs).",
				Elem: &schema.Schema{
					Type: schema.TypeFloat,
				},
			},
			"_api_response": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The full API response from OpenAI.",
			},
		},
	}
}

// resourceOpenAIModerationCreate initiates content moderation analysis.
// It processes the moderation request, handles the API call, and manages the response.
// The function provides detailed analysis of content safety and compliance.
func resourceOpenAIModerationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	input := d.Get("input")
	model := d.Get("model").(string)

	var requestBody map[string]interface{}

	// Store the original input for state consistency
	originalInput := input
	inputStr, isString := input.(string)

	// Handle different input types with better checks for JSON strings
	if isString && (strings.HasPrefix(inputStr, "[") || strings.HasPrefix(inputStr, "{")) {
		// Input appears to be a JSON string (array or object)
		var parsedJSON interface{}
		if err := json.Unmarshal([]byte(inputStr), &parsedJSON); err == nil {
			// Successfully parsed JSON
			requestBody = map[string]interface{}{
				"input": parsedJSON,
			}
		} else {
			// Failed to parse as JSON, use as regular string
			requestBody = map[string]interface{}{
				"input": inputStr,
			}
		}
	} else if isString {
		// Regular string input
		requestBody = map[string]interface{}{
			"input": inputStr,
		}
	} else {
		// Non-string input (likely an array or object)
		requestBody = map[string]interface{}{
			"input": input,
		}
	}

	// Always set the input back to its original form for state consistency
	if err := d.Set("input", originalInput); err != nil {
		return diag.FromErr(fmt.Errorf("error setting input: %s", err))
	}

	if model != "" {
		requestBody["model"] = model
	}

	// Perform the request
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing moderation request: %v", err))
	}

	// Debug output
	fmt.Printf("DEBUG: Moderation request body: %s\n", string(reqBody))

	// Use the provider's API key
	respBody, reqErr := client.DoRequest("POST", "moderations", requestBody)
	if reqErr != nil {
		return diag.FromErr(reqErr)
	}

	// Debug output
	fmt.Printf("DEBUG: Moderation response: %s\n", string(respBody))

	// Parse the response
	var moderationResp map[string]interface{}
	if err := json.Unmarshal(respBody, &moderationResp); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing moderation response: %s", err))
	}

	// Store the original API response for better state preservation
	respBodyStr := string(respBody)
	if err := d.Set("_api_response", respBodyStr); err != nil {
		return diag.FromErr(err)
	}

	// Set the id to a unique value based on the response id
	id, ok := moderationResp["id"].(string)
	if !ok {
		return diag.FromErr(fmt.Errorf("error getting moderation id from response"))
	}
	d.SetId(id)

	// Convert any lists to proper format before setting in state
	resultsRaw, ok := moderationResp["results"].([]interface{})
	if !ok {
		return diag.FromErr(fmt.Errorf("error getting results from response"))
	}

	results := make([]map[string]interface{}, 0, len(resultsRaw))
	for _, r := range resultsRaw {
		result, ok := r.(map[string]interface{})
		if !ok {
			return diag.FromErr(fmt.Errorf("error parsing result in response"))
		}

		// Make a deep copy to avoid modifying the original data
		resultCopy := make(map[string]interface{})
		for k, v := range result {
			resultCopy[k] = v
		}

		// Ensure categories and category_scores are properly formatted
		if cats, ok := resultCopy["categories"].(map[string]interface{}); ok {
			// Convert any non-bool values to bools if needed
			for k, v := range cats {
				switch v.(type) {
				case bool:
					// Already a bool, nothing to do
				default:
					// Try to convert to bool
					if b, ok := v.(bool); ok {
						cats[k] = b
					} else {
						// If can't convert, set to false
						cats[k] = false
					}
				}
			}
			resultCopy["categories"] = cats
		}

		// Handle category_applied_input_types field if present
		if catAppliedTypes, ok := resultCopy["category_applied_input_types"].(map[string]interface{}); ok {
			// The field exists, but ensure it's properly formatted
			resultCopy["category_applied_input_types"] = catAppliedTypes
		} else {
			// If field is missing, add an empty map
			resultCopy["category_applied_input_types"] = map[string]interface{}{}
		}

		results = append(results, resultCopy)
	}

	// Set results in state
	if err := d.Set("results", results); err != nil {
		return diag.FromErr(err)
	}

	// For convenience, also save the first result's data separately
	if len(results) > 0 {
		result := results[0]
		if err := d.Set("flagged", result["flagged"]); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("categories", result["categories"]); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("category_scores", result["category_scores"]); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set the model used
	usedModel, ok := moderationResp["model"].(string)
	if ok {
		// Store the model returned by the API, not the one specified in the config
		// This way when applying again, it won't detect a diff between config and state
		if err := d.Set("model", usedModel); err != nil {
			return diag.FromErr(fmt.Errorf("error setting model: %s", err))
		}

		// Print debug info about the model values
		fmt.Printf("DEBUG: Model in config: %s, Model returned by API: %s\n", model, usedModel)
	}

	return nil
}

// resourceOpenAIModerationRead retrieves the current state of moderation results.
// Since moderation results are immutable, this just verifies the resource ID exists
// and ensures all state values are properly preserved.
func resourceOpenAIModerationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// The moderation API doesn't provide a way to retrieve a specific moderation result.
	// Instead, we just use the stored state to re-populate the Terraform state.

	// But we still ensure the client is properly configured
	_, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// The ID is already set, and the data should be populated from state,
	// so there's nothing to read.
	return nil
}

// resourceOpenAIModerationDelete removes moderation results.
// Note: Moderation results cannot be deleted through the API.
// This function only removes the resource from the Terraform state.
func resourceOpenAIModerationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Moderation resources are essentially stateless from OpenAI's perspective.
	// There's nothing to delete on the server side, so we just remove the resource from state.
	d.SetId("")
	return nil
}

// customizeDiffModeration handles differences between model versions
// Specifically, it ignores differences when the configuration specifies
// a generic model version (like "text-moderation-latest") but the API returns
// a specific model version (like "text-moderation-007")
func customizeDiffModeration(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// Skip if this is a new resource (not yet created)
	if d.Id() == "" {
		return nil
	}

	// Get the state and config values - oldValue is from state, newValue is from config
	oldValue, newValue := d.GetChange("model")

	// Convert to strings
	oldModel, oldOk := oldValue.(string)
	newModel, newOk := newValue.(string)

	// Skip if we couldn't get proper values
	if !oldOk || !newOk {
		return nil
	}

	// Debug logs to help troubleshoot
	fmt.Printf("DEBUG: Comparing models - old(state): %s, new(config): %s\n", oldModel, newModel)

	// If config specifies a generic model and state has a specific version, tell Terraform to ignore
	isGenericModel := newModel == "text-moderation-latest" ||
		newModel == "text-moderation-stable" ||
		newModel == ""

	isSpecificVersion := strings.HasPrefix(oldModel, "text-moderation-") &&
		oldModel != "text-moderation-latest" &&
		oldModel != "text-moderation-stable"

	if isGenericModel && isSpecificVersion {
		fmt.Printf("DEBUG: Suppressing model diff from %s to %s\n", oldModel, newModel)

		// This is the key line - Set the new value (from config) to be the old value (from state)
		// This makes Terraform think there's no difference
		_ = d.SetNew("model", oldModel)
		return nil
	}

	return nil
}

func getUrl(client *OpenAIClient, endpoint string) string {
	url := fmt.Sprintf("%s/v1/%s", client.APIURL, endpoint)
	if strings.Contains(client.APIURL, "/v1") {
		url = fmt.Sprintf("%s/%s", client.APIURL, endpoint)
	}
	return url
}
