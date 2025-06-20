package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FineTuningJobsResponse represents the API response for listing fine-tuning jobs.
type FineTuningJobsResponse struct {
	Object  string              `json:"object"`   // Type of object, usually "list"
	Data    []FineTuningJobData `json:"data"`     // List of fine-tuning jobs
	HasMore bool                `json:"has_more"` // Whether there are more jobs to fetch
}

// FineTuningJobData represents a fine-tuning job entry in the list.
type FineTuningJobData struct {
	ID              string                `json:"id"`                         // Unique identifier for the fine-tuning job
	Object          string                `json:"object"`                     // Type of object (e.g., "fine_tuning.job")
	Model           string                `json:"model"`                      // Base model being fine-tuned
	CreatedAt       int                   `json:"created_at"`                 // Unix timestamp of job creation
	FinishedAt      *int                  `json:"finished_at,omitempty"`      // Unix timestamp of job completion
	Status          string                `json:"status"`                     // Current status of the fine-tuning job
	TrainingFile    string                `json:"training_file"`              // ID of the training data file
	ValidationFile  *string               `json:"validation_file,omitempty"`  // Optional ID of validation data file
	Hyperparameters FineTuningHyperparams `json:"hyperparameters"`            // Training hyperparameters
	ResultFiles     []string              `json:"result_files"`               // List of result file IDs
	TrainedTokens   *int                  `json:"trained_tokens,omitempty"`   // Number of tokens processed
	FineTunedModel  *string               `json:"fine_tuned_model,omitempty"` // ID of the resulting model
	Error           *FineTuningError      `json:"error,omitempty"`            // Error information if job failed
}

func dataSourceOpenAIFineTuningJobs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIFineTuningJobsRead,
		Schema: map[string]*schema.Schema{
			"after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier for the last job from the previous pagination request",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20,
				Description: "Number of fine-tuning jobs to retrieve (default: 20)",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Optional metadata filter in the format {key: value}",
			},
			"jobs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the fine-tuning job",
						},
						"object": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The object type, always 'fine_tuning.job'",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The base model being fine-tuned",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Unix timestamp of when the job was created",
						},
						"finished_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Unix timestamp of when the job finished",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of the fine-tuning job",
						},
						"training_file": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the training data file",
						},
						"validation_file": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Optional ID of validation data file",
						},
						"hyperparameters": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Hyperparameters used for the fine-tuning job",
						},
						"result_files": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of result file IDs",
						},
						"trained_tokens": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of tokens processed during training",
						},
						"fine_tuned_model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the resulting fine-tuned model",
						},
						"error": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Error information if job failed",
						},
					},
				},
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more fine-tuning jobs to retrieve",
			},
		},
	}
}

func dataSourceOpenAIFineTuningJobsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build the query parameters
	queryParams := url.Values{}

	if v, ok := d.GetOk("after"); ok {
		queryParams.Add("after", v.(string))
	}

	if v, ok := d.GetOk("limit"); ok {
		queryParams.Add("limit", strconv.Itoa(v.(int)))
	}

	// Handle metadata filter if provided
	if v, ok := d.GetOk("metadata"); ok {
		metadata := v.(map[string]interface{})
		for k, v := range metadata {
			queryParams.Add(fmt.Sprintf("metadata[%s]", k), v.(string))
		}
	}

	// Build the URL with query parameters
	apiURL := fmt.Sprintf("%s/fine_tuning/jobs", client.APIURL)
	if len(queryParams) > 0 {
		apiURL = fmt.Sprintf("%s?%s", apiURL, queryParams.Encode())
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %s", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %s", err))
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %s", err))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(body)))
	}

	// Parse the response
	var fineTuningJobs FineTuningJobsResponse
	if err := json.Unmarshal(body, &fineTuningJobs); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Create a unique ID for this data source
	if len(fineTuningJobs.Data) > 0 {
		d.SetId(fmt.Sprintf("fine-tuning-jobs-%d", fineTuningJobs.Data[0].CreatedAt))
	} else {
		// No jobs found, use a generic ID
		d.SetId(fmt.Sprintf("fine-tuning-jobs-%d", time.Now().Unix()))
	}

	// Set the has_more value
	if err := d.Set("has_more", fineTuningJobs.HasMore); err != nil {
		return diag.FromErr(fmt.Errorf("error setting has_more: %s", err))
	}

	// Convert the jobs data to the format expected by the schema
	jobsList := make([]map[string]interface{}, 0, len(fineTuningJobs.Data))
	for _, job := range fineTuningJobs.Data {
		jobMap := map[string]interface{}{
			"id":            job.ID,
			"object":        job.Object,
			"model":         job.Model,
			"created_at":    job.CreatedAt,
			"status":        job.Status,
			"training_file": job.TrainingFile,
			"result_files":  job.ResultFiles,
		}

		// Handle optional fields
		if job.FinishedAt != nil {
			jobMap["finished_at"] = *job.FinishedAt
		}

		if job.ValidationFile != nil {
			jobMap["validation_file"] = *job.ValidationFile
		}

		if job.TrainedTokens != nil {
			jobMap["trained_tokens"] = *job.TrainedTokens
		}

		if job.FineTunedModel != nil {
			jobMap["fine_tuned_model"] = *job.FineTunedModel
		}

		// Convert hyperparameters to a map
		hyperparams := make(map[string]interface{})
		hyperparamsJSON, err := json.Marshal(job.Hyperparameters)
		if err == nil {
			json.Unmarshal(hyperparamsJSON, &hyperparams)

			// Convert numeric values to strings for the Terraform schema
			stringHyperparams := make(map[string]interface{})
			for k, v := range hyperparams {
				switch val := v.(type) {
				case float64:
					// Convert float64 to string with minimal decimal places
					if val == float64(int(val)) {
						// If it's a whole number, format as integer
						stringHyperparams[k] = fmt.Sprintf("%d", int(val))
					} else {
						stringHyperparams[k] = fmt.Sprintf("%g", val)
					}
				case int:
					stringHyperparams[k] = fmt.Sprintf("%d", val)
				default:
					// Keep as is for strings and other types
					stringHyperparams[k] = v
				}
			}

			jobMap["hyperparameters"] = stringHyperparams
		}

		// Handle error if present
		if job.Error != nil {
			errorMap := map[string]interface{}{
				"message": job.Error.Message,
				"type":    job.Error.Type,
				"code":    job.Error.Code,
			}
			jobMap["error"] = errorMap
		}

		jobsList = append(jobsList, jobMap)
	}

	// Set the jobs list
	if err := d.Set("jobs", jobsList); err != nil {
		return diag.FromErr(fmt.Errorf("error setting jobs: %s", err))
	}

	return diag.Diagnostics{}
}
