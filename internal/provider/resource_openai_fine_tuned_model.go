package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mkdev-me/terraform-provider-openai/internal/client"
)

// FineTuningJobResponse represents the API response for a fine-tuning job.
// It contains comprehensive information about the fine-tuning process, including job status,
// training details, and results. This structure captures all aspects of the fine-tuning job
// from creation to completion.
type FineTuningJobResponse struct {
	ID              string                `json:"id"`                         // Unique identifier for the fine-tuning job
	Object          string                `json:"object"`                     // Type of object (e.g., "fine_tuning.job")
	Model           string                `json:"model"`                      // Base model being fine-tuned
	CreatedAt       int                   `json:"created_at"`                 // Unix timestamp of job creation
	FinishedAt      int                   `json:"finished_at,omitempty"`      // Unix timestamp of job completion
	Status          string                `json:"status"`                     // Current status of the fine-tuning job
	TrainingFile    string                `json:"training_file"`              // ID of the training data file
	ValidationFile  string                `json:"validation_file,omitempty"`  // Optional ID of validation data file
	Hyperparameters FineTuningHyperparams `json:"hyperparameters"`            // Training hyperparameters
	ResultFiles     []string              `json:"result_files"`               // List of result file IDs
	TrainedTokens   int                   `json:"trained_tokens,omitempty"`   // Number of tokens processed
	FineTunedModel  string                `json:"fine_tuned_model,omitempty"` // ID of the resulting model
	Error           *FineTuningError      `json:"error,omitempty"`            // Error information if job failed
}

// FineTuningError represents an error that occurred during fine-tuning.
// It provides detailed information about what went wrong during the training process.
type FineTuningError struct {
	Message string `json:"message"` // Human-readable error message
	Type    string `json:"type"`    // Type of error (e.g., "validation_error")
	Code    string `json:"code"`    // Error code for programmatic handling
}

// FineTuningHyperparams represents the hyperparameters used for fine-tuning.
// These parameters control various aspects of the training process and can be
// customized to achieve different training objectives.
type FineTuningHyperparams struct {
	NEpochs                interface{} `json:"n_epochs,omitempty"`                 // Number of training epochs
	BatchSize              interface{} `json:"batch_size,omitempty"`               // Size of training batches
	LearningRateMultiplier interface{} `json:"learning_rate_multiplier,omitempty"` // Learning rate adjustment factor
}

// FineTuningJobRequest represents the request payload for creating a fine-tuning job.
// It specifies the model to fine-tune, training data, and optional parameters
// that control the fine-tuning process.
type FineTuningJobRequest struct {
	Model           string                 `json:"model"`                     // Base model to fine-tune
	TrainingFile    string                 `json:"training_file"`             // ID of the training data file
	ValidationFile  string                 `json:"validation_file,omitempty"` // Optional validation data file
	Hyperparameters *FineTuningHyperparams `json:"hyperparameters,omitempty"` // Optional training parameters
	Suffix          string                 `json:"suffix,omitempty"`          // Optional suffix for the fine-tuned model name
}

// MarshalJSON helps debug the JSON marshaling
func (hp *FineTuningHyperparams) MarshalJSON() ([]byte, error) {
	type Alias FineTuningHyperparams

	// For debugging purposes
	fmt.Printf("Marshaling hyperparameters. NEpochs type: %T, value: %v\n", hp.NEpochs, hp.NEpochs)

	// Special handling for when n_epochs is "auto" - use json.RawMessage for literal insertion
	// When using "auto", OpenAI requires ONLY n_epochs to be specified without other parameters
	if s, ok := hp.NEpochs.(string); ok && s == "auto" {
		fmt.Println("Special handling for 'auto' value - excluding other hyperparameters")

		// Create a map with ONLY n_epochs
		m := map[string]interface{}{
			"n_epochs": "auto", // Use string "auto" and let standard marshaling handle it
		}

		result, err := json.Marshal(m)
		fmt.Printf("Final JSON for hyperparameters with auto: %s\n", string(result))
		return result, err
	}

	// Standard marshaling for other cases
	return json.Marshal((*Alias)(hp))
}

// resourceOpenAIFineTunedModel defines the schema and CRUD operations for OpenAI fine-tuned models.
// This resource allows users to create, manage, and monitor fine-tuning jobs for OpenAI models.
// It provides comprehensive control over the fine-tuning process and access to training results.
func resourceOpenAIFineTunedModel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenAIFineTunedModelCreate,
		ReadContext:   resourceOpenAIFineTunedModelRead,
		UpdateContext: resourceOpenAIFineTunedModelUpdate,
		DeleteContext: resourceOpenAIFineTunedModelDelete,
		Schema: map[string]*schema.Schema{
			"model": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the base model to fine-tune",
			},
			"training_file": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the file containing training data",
			},
			"validation_file": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The ID of the file containing validation data",
			},
			"hyperparameters": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Hyperparameters for the fine-tuning job",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"n_epochs": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "4",
							Description: "Number of epochs to train for. Can be an integer or 'auto'",
						},
						"batch_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "Batch size to use for training",
						},
						"learning_rate_multiplier": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ValidateFunc: validation.FloatBetween(0.01, 10.0),
							Description:  "Learning rate multiplier to use for training",
						},
					},
				},
			},
			"suffix": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "A suffix to append to the fine-tuned model name",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the fine-tuning job was created",
			},
			"finished_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp for when the fine-tuning job completed",
			},
			"fine_tuned_model": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the fine-tuned model",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the fine-tuning job",
			},
			"completion_window": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Time in seconds to wait for job to complete during creation. 0 means don't wait.",
			},
		},
	}
}

// resourceOpenAIFineTunedModelCreate initiates a new fine-tuning job.
// It handles the creation of a fine-tuning job, including validation of inputs,
// submission of the job to OpenAI's API, and monitoring of the training process.
func resourceOpenAIFineTunedModelCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Construct the request directly from Go structs
	request := FineTuningJobRequest{
		Model:        d.Get("model").(string),
		TrainingFile: d.Get("training_file").(string),
	}

	if v, ok := d.GetOk("validation_file"); ok {
		request.ValidationFile = v.(string)
	}

	if v, ok := d.GetOk("suffix"); ok {
		request.Suffix = v.(string)
	}

	// Hyperparameters handling
	if hyperparameters, ok := d.GetOk("hyperparameters"); ok && len(hyperparameters.([]interface{})) > 0 {
		hp := hyperparameters.([]interface{})[0].(map[string]interface{})
		request.Hyperparameters = &FineTuningHyperparams{}

		if v, ok := hp["n_epochs"]; ok {
			nepochs := v.(string)
			if nepochs == "auto" {
				request.Hyperparameters.NEpochs = "auto"
			} else if nepochsInt, err := strconv.Atoi(nepochs); err == nil {
				request.Hyperparameters.NEpochs = nepochsInt
			} else {
				return diag.FromErr(fmt.Errorf("invalid value for n_epochs: %s", nepochs))
			}
		}

		if v, ok := hp["batch_size"]; ok {
			request.Hyperparameters.BatchSize = v.(int)
		}

		if v, ok := hp["learning_rate_multiplier"]; ok {
			// Convert to float64 if specified in terraform
			request.Hyperparameters.LearningRateMultiplier = v.(float64)
		}
	}

	// Marshal request into JSON
	reqBody, err := json.Marshal(request)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshaling request: %v", err))
	}

	// More detailed debugging
	fmt.Printf("\n-------------- REQUEST JSON START --------------\n")
	fmt.Printf("%s\n", string(reqBody))
	fmt.Printf("--------------- REQUEST JSON END ---------------\n")

	// For extra clarity, unmarshal and marshal again to see how Go interprets it
	var debugMap map[string]interface{}
	if err := json.Unmarshal(reqBody, &debugMap); err == nil {
		if hpMap, ok := debugMap["hyperparameters"].(map[string]interface{}); ok {
			fmt.Printf("n_epochs in hyperparameters after unmarshal: type=%T, value=%v\n",
				hpMap["n_epochs"], hpMap["n_epochs"])
		}
	}

	// Prepare HTTP request
	url := fmt.Sprintf("%s/fine_tuning/jobs", client.APIURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request: %v", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Send HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request: %v", err))
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response: %v", err))
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing error response: %v, status code: %d, body: %s",
				err, resp.StatusCode, string(respBody)))
		}
		return diag.FromErr(fmt.Errorf("error creating fine-tuning job: %s - %s",
			errorResponse.Error.Type, errorResponse.Error.Message))
	}

	// Parse success response
	var jobResponse FineTuningJobResponse
	if err := json.Unmarshal(respBody, &jobResponse); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
	}

	// Set resource ID and attributes
	d.SetId(jobResponse.ID)
	if err := d.Set("created_at", jobResponse.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", jobResponse.Status); err != nil {
		return diag.FromErr(err)
	}

	// Handle completion window logic as before
	if completionWindow, ok := d.GetOk("completion_window"); ok && completionWindow.(int) > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(completionWindow.(int))*time.Second)
		defer cancel()

		fineTuningJob, err := waitForFineTuningJobCompletion(timeoutCtx, client, jobResponse.ID)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("status", fineTuningJob.Status); err != nil {
			return diag.FromErr(err)
		}
		if fineTuningJob.FinishedAt != 0 {
			if err := d.Set("finished_at", fineTuningJob.FinishedAt); err != nil {
				return diag.FromErr(err)
			}
		}
		if fineTuningJob.FineTunedModel != "" {
			if err := d.Set("fine_tuned_model", fineTuningJob.FineTunedModel); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return diag.Diagnostics{}
}

// waitForFineTuningJobCompletion polls the OpenAI API until a fine-tuning job completes.
// It returns the final job status or an error if polling fails or times out.
// The caller should handle timeouts as appropriate for their use case.
func waitForFineTuningJobCompletion(ctx context.Context, client *client.OpenAIClient, jobID string) (*FineTuningJobResponse, error) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for fine-tuning job completion")
		case <-ticker.C:
			// Consultar el estado del trabajo
			job, err := getFineTuningJob(ctx, client, jobID)
			if err != nil {
				return nil, err
			}

			// Verificar si el trabajo ha finalizado
			if job.Status == "succeeded" || job.Status == "failed" || job.Status == "cancelled" {
				return job, nil
			}
		}
	}
}

// getFineTuningJob retrieves the current status and details of a fine-tuning job.
// It makes an API request to fetch the latest information from OpenAI.
func getFineTuningJob(ctx context.Context, client *client.OpenAIClient, jobID string) (*FineTuningJobResponse, error) {
	url := fmt.Sprintf("%s/fine_tuning/jobs/%s", client.APIURL, jobID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Establecer headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	// Añadir Organization ID si está presente
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Realizar la petición
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Leer la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Verificar si hubo un error
	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			// If API returns 404 with a non-standard error format, assume the fine-tune job doesn't exist
			if resp.StatusCode == http.StatusNotFound {
				return nil, fmt.Errorf("error getting fine-tuning job: fine-tune job not found (404)")
			}
			return nil, fmt.Errorf("error parsing error response: %v, status code: %d, body: %s",
				err, resp.StatusCode, string(respBody))
		}

		// Check specific error messages that indicate job doesn't exist
		if resp.StatusCode == http.StatusNotFound ||
			(errorResponse.Error.Type == "invalid_request_error" &&
				strings.Contains(errorResponse.Error.Message, "Could not find fine tune")) {
			return nil, fmt.Errorf("error getting fine-tuning job: fine-tune job not found (404)")
		}

		return nil, fmt.Errorf("error getting fine-tuning job: %s - %s",
			errorResponse.Error.Type, errorResponse.Error.Message)
	}

	// Parsear la respuesta
	var jobResponse FineTuningJobResponse
	if err := json.Unmarshal(respBody, &jobResponse); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &jobResponse, nil
}

// resourceOpenAIFineTunedModelRead retrieves the current state of a fine-tuned model.
// It fetches the latest information about the model and updates the Terraform state.
func resourceOpenAIFineTunedModelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Obtener el ID del trabajo
	jobID := d.Id()

	// Si el ID está vacío, el recurso ya no existe
	if jobID == "" {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Consultar el estado del trabajo
	job, err := getFineTuningJob(ctx, client, jobID)
	if err != nil {
		// Si el trabajo no se encuentra, marcar el recurso como eliminado
		if strings.Contains(err.Error(), "404") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "Could not find fine tune") {
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(err)
	}

	// Actualizar el estado con los datos de la respuesta
	// Don't update the model field to prevent unnecessary recreations
	// The API returns specific model versions (e.g., gpt-3.5-turbo-0125) which would cause Terraform
	// to try to recreate the resource if it differs from the config (e.g., gpt-3.5-turbo)

	// d.Set("model", job.Model) - Commented out to preserve the original model name

	if err := d.Set("training_file", job.TrainingFile); err != nil {
		return diag.FromErr(err)
	}
	if job.ValidationFile != "" {
		if err := d.Set("validation_file", job.ValidationFile); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("created_at", job.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", job.Status); err != nil {
		return diag.FromErr(err)
	}

	if job.FinishedAt > 0 {
		if err := d.Set("finished_at", job.FinishedAt); err != nil {
			return diag.FromErr(err)
		}
	}

	if job.FineTunedModel != "" {
		if err := d.Set("fine_tuned_model", job.FineTunedModel); err != nil {
			return diag.FromErr(err)
		}
	}

	// No actualizamos los hiperparámetros ya que podrían ser diferentes de los proporcionados inicialmente

	return diag.Diagnostics{}
}

// resourceOpenAIFineTunedModelDelete handles the deletion of a fine-tuned model.
// This function only removes the resource from the Terraform state.
func resourceOpenAIFineTunedModelDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := GetOpenAIClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	// Obtener el nombre del modelo fine-tuned
	fineTunedModel := d.Get("fine_tuned_model").(string)

	// Si no hay un modelo fine-tuned, simplemente limpiar el ID y salir
	if fineTunedModel == "" {
		d.SetId("")
		return diag.Diagnostics{}
	}

	// Verificar si el trabajo aún está en progreso
	status := d.Get("status").(string)
	if status == "pending" || status == "running" {
		// Intentar cancelar el trabajo en curso
		jobID := d.Id()
		url := fmt.Sprintf("%s/fine_tuning/jobs/%s/cancel", client.APIURL, jobID)

		req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating request to cancel job: %v", err))
		}

		// Establecer headers
		req.Header.Set("Authorization", "Bearer "+client.APIKey)

		// Añadir Organization ID si está presente
		if client.OrganizationID != "" {
			req.Header.Set("OpenAI-Organization", client.OrganizationID)
		}

		// Realizar la petición
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error making request to cancel job: %v", err))
		}
		defer resp.Body.Close()

		// Si el trabajo no se pudo cancelar, continuar con la eliminación del modelo
		if resp.StatusCode != http.StatusOK {
			// Simplemente registrar el error pero continuar
			fmt.Printf("Warning: Could not cancel job %s (status code: %d)\n", jobID, resp.StatusCode)
		}
	}

	// Eliminar el modelo fine-tuned (nota: esto no siempre es posible con la API de OpenAI)
	url := fmt.Sprintf("%s/models/%s", client.APIURL, fineTunedModel)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating request to delete model: %v", err))
	}

	// Establecer headers
	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	// Añadir Organization ID si está presente
	if client.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", client.OrganizationID)
	}

	// Realizar la petición
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error making request to delete model: %v", err))
	}
	defer resp.Body.Close()

	// Si el modelo no se pudo eliminar, registrar el error pero continuar
	if resp.StatusCode != http.StatusOK {
		// Leer el cuerpo de la respuesta para obtener detalles del error
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("Warning: Could not delete model %s (status code: %d, body: %s)\n",
			fineTunedModel, resp.StatusCode, string(respBody))
	}

	// Limpiar el ID del recurso
	d.SetId("")
	return diag.Diagnostics{}
}

// resourceOpenAIFineTunedModelUpdate handles the update of a fine-tuned model resource.
// Currently, fine-tuned models cannot be updated after creation, so this simply
// re-reads the current state of the resource.
func resourceOpenAIFineTunedModelUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Fine-tuned models cannot be updated after creation
	// Just return the result of Read to refresh the state
	return resourceOpenAIFineTunedModelRead(ctx, d, m)
}
