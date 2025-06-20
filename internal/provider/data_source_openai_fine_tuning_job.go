package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOpenAIFineTuningJob() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAIFineTuningJobRead,
		Schema: map[string]*schema.Schema{
			"fine_tuning_job_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the fine-tuning job to retrieve",
			},
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
			"estimated_finish": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Estimated finish time in Unix timestamp format",
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
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization ID the model belongs to",
			},
			"error": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Error information if job failed",
			},
			"integrations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"wandb": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"project": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"entity": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"tags": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
				Description: "Integrations associated with the fine-tuning job",
			},
			"user_provided_suffix": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User-provided suffix for the fine-tuned model",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Metadata associated with the fine-tuning job",
			},
			"seed": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Seed for reproducibility",
			},
			"method": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"supervised": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hyperparameters": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"batch_size": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"learning_rate_multiplier": {
													Type:     schema.TypeFloat,
													Computed: true,
												},
												"n_epochs": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"dpo": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hyperparameters": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"beta": {
													Type:     schema.TypeFloat,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Description: "The fine-tuning method used",
			},
		},
	}
}

func dataSourceOpenAIFineTuningJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	jobID := d.Get("fine_tuning_job_id").(string)

	// Build the URL
	apiURL := fmt.Sprintf("%s/fine_tuning/jobs/%s", client.APIURL, jobID)

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
		// Handle 404 Not Found errors gracefully
		if resp.StatusCode == http.StatusNotFound {
			// Set the ID anyway so Terraform has something to reference
			d.SetId(jobID)

			// Create placeholder data for required fields
			if err := d.Set("object", "fine_tuning.job"); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("model", "unknown"); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("status", "unknown"); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("created_at", time.Now().Unix()); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("training_file", "file-unknown"); err != nil {
				return diag.FromErr(err)
			}

			// Return a warning instead of error
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Fine-tuning job not found",
					Detail:   fmt.Sprintf("Fine-tuning job with ID '%s' could not be found. This may be because it has been deleted or has expired. Using placeholder data.", jobID),
				},
			}
		}

		// For other error types, return the normal error
		return diag.FromErr(fmt.Errorf("API returned error: %s - %s", resp.Status, string(body)))
	}

	// Parse the response
	var fineTuningJob map[string]interface{}
	if err := json.Unmarshal(body, &fineTuningJob); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %s", err))
	}

	// Set the computed fields
	d.SetId(jobID)

	// Set all fields from the response
	for key, value := range fineTuningJob {
		// Handle lists specifically (like result_files)
		if key == "result_files" {
			if files, ok := value.([]interface{}); ok {
				resultFiles := make([]string, len(files))
				for i, file := range files {
					resultFiles[i] = file.(string)
				}
				if err := d.Set("result_files", resultFiles); err != nil {
					return diag.FromErr(fmt.Errorf("error setting result_files: %s", err))
				}
			}
		} else if key == "hyperparameters" {
			// Convert hyperparameters to map[string]interface{}
			if hyperparams, ok := value.(map[string]interface{}); ok {
				stringHyperparams := make(map[string]string)
				for k, v := range hyperparams {
					// Convert each value to string representation
					switch val := v.(type) {
					case string:
						stringHyperparams[k] = val
					case float64:
						stringHyperparams[k] = fmt.Sprintf("%v", val)
					case nil:
						stringHyperparams[k] = "null"
					default:
						stringHyperparams[k] = fmt.Sprintf("%v", val)
					}
				}
				if err := d.Set("hyperparameters", stringHyperparams); err != nil {
					return diag.FromErr(fmt.Errorf("error setting hyperparameters: %s", err))
				}
			}
		} else if key == "error" && value != nil {
			// Handle error object if present
			if errorData, ok := value.(map[string]interface{}); ok {
				errorMap := make(map[string]string)
				for k, v := range errorData {
					if str, ok := v.(string); ok {
						errorMap[k] = str
					} else {
						errorMap[k] = fmt.Sprintf("%v", v)
					}
				}
				if err := d.Set("error", errorMap); err != nil {
					return diag.FromErr(fmt.Errorf("error setting error field: %s", err))
				}
			}
		} else if key == "integrations" && value != nil {
			// Handle integrations specially
			if integrationsData, ok := value.([]interface{}); ok && len(integrationsData) > 0 {
				integrations := make([]map[string]interface{}, 0, len(integrationsData))

				for _, integItem := range integrationsData {
					if integMap, ok := integItem.(map[string]interface{}); ok {
						integration := make(map[string]interface{})

						// Copy basic fields
						if v, ok := integMap["id"]; ok {
							integration["id"] = v
						}
						if v, ok := integMap["type"]; ok {
							integration["type"] = v
						}

						// Handle wandb object if present
						if wandbData, ok := integMap["wandb"].(map[string]interface{}); ok {
							wandb := make([]map[string]interface{}, 1)
							wandbMap := make(map[string]interface{})

							if v, ok := wandbData["project"]; ok {
								wandbMap["project"] = v
							}
							if v, ok := wandbData["entity"]; ok {
								wandbMap["entity"] = v
							}
							if v, ok := wandbData["name"]; ok {
								wandbMap["name"] = v
							}

							// Handle tags array
							if tags, ok := wandbData["tags"].([]interface{}); ok {
								tagsList := make([]string, len(tags))
								for i, tag := range tags {
									tagsList[i] = tag.(string)
								}
								wandbMap["tags"] = tagsList
							}

							wandb[0] = wandbMap
							integration["wandb"] = wandb
						}

						integrations = append(integrations, integration)
					}
				}

				if err := d.Set("integrations", integrations); err != nil {
					return diag.FromErr(fmt.Errorf("error setting integrations: %s", err))
				}
			}
		} else if key == "method" && value != nil {
			// Handle method specially
			if methodMap, ok := value.(map[string]interface{}); ok {
				methods := make([]map[string]interface{}, 1)
				method := make(map[string]interface{})

				// Set the method type
				if typeVal, ok := methodMap["type"].(string); ok {
					method["type"] = typeVal

					// Handle supervised method
					if typeVal == "supervised" && methodMap["supervised"] != nil {
						if supervisedData, ok := methodMap["supervised"].(map[string]interface{}); ok {
							supervised := make([]map[string]interface{}, 1)
							supervisedMap := make(map[string]interface{})

							// Handle hyperparameters
							if hyperParams, ok := supervisedData["hyperparameters"].(map[string]interface{}); ok {
								hyperList := make([]map[string]interface{}, 1)
								hyperMap := make(map[string]interface{})

								if v, ok := hyperParams["batch_size"].(float64); ok {
									hyperMap["batch_size"] = int(v)
								}
								if v, ok := hyperParams["learning_rate_multiplier"].(float64); ok {
									hyperMap["learning_rate_multiplier"] = v
								}
								if v, ok := hyperParams["n_epochs"].(float64); ok {
									hyperMap["n_epochs"] = int(v)
								}

								hyperList[0] = hyperMap
								supervisedMap["hyperparameters"] = hyperList
							}

							supervised[0] = supervisedMap
							method["supervised"] = supervised
						}
					}

					// Handle DPO method
					if typeVal == "dpo" && methodMap["dpo"] != nil {
						if dpoData, ok := methodMap["dpo"].(map[string]interface{}); ok {
							dpo := make([]map[string]interface{}, 1)
							dpoMap := make(map[string]interface{})

							// Handle hyperparameters
							if hyperParams, ok := dpoData["hyperparameters"].(map[string]interface{}); ok {
								hyperList := make([]map[string]interface{}, 1)
								hyperMap := make(map[string]interface{})

								if v, ok := hyperParams["beta"].(float64); ok {
									hyperMap["beta"] = v
								}

								hyperList[0] = hyperMap
								dpoMap["hyperparameters"] = hyperList
							}

							dpo[0] = dpoMap
							method["dpo"] = dpo
						}
					}
				}

				methods[0] = method
				if err := d.Set("method", methods); err != nil {
					return diag.FromErr(fmt.Errorf("error setting method: %s", err))
				}
			}
		} else {
			// For other fields, set them directly
			if err := d.Set(key, value); err != nil {
				return diag.FromErr(fmt.Errorf("error setting %s: %s", key, err))
			}
		}
	}

	return diag.Diagnostics{}
}
