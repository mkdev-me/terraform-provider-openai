package provider

// FineTuningJobResponse represents the API response for fine-tuning jobs.
type FineTuningJobResponse struct {
	ID              string                   `json:"id"`
	Model           string                   `json:"model"`
	TrainingFile    string                   `json:"training_file"`
	ValidationFile  string                   `json:"validation_file,omitempty"`
	FineTunedModel  string                   `json:"fine_tuned_model,omitempty"`
	OrganizationID  string                   `json:"organization_id"`
	Status          string                   `json:"status"`
	CreatedAt       int64                    `json:"created_at"`
	FinishedAt      *int64                   `json:"finished_at,omitempty"`
	ResultFiles     []string                 `json:"result_files"`
	TrainedTokens   int64                    `json:"trained_tokens"`
	ValidationLoss  float64                  `json:"validation_loss,omitempty"`
	Hyperparameters *HyperparametersResponse `json:"hyperparameters,omitempty"`
	Integrations    []IntegrationResponse    `json:"integrations,omitempty"`
	Seed            int                      `json:"seed,omitempty"`
	DatasetID       string                   `json:"dataset_id,omitempty"` // New field often appearing
	Estimator       string                   `json:"estimator,omitempty"`
	Metadata        map[string]interface{}   `json:"metadata,omitempty"`
}

type HyperparametersResponse struct {
	NEpochs                interface{} `json:"n_epochs"`                 // Can be "auto" or int
	BatchSize              interface{} `json:"batch_size"`               // Can be "auto" or int
	LearningRateMultiplier interface{} `json:"learning_rate_multiplier"` // Can be "auto" or float
}

type IntegrationResponse struct {
	Type  string            `json:"type"`
	WandB *WandBIntegration `json:"wandb,omitempty"`
}

// FineTuningJobCreateRequest represents the request to create a fine-tuning job.
type FineTuningJobCreateRequest struct {
	Model          string                 `json:"model"`
	TrainingFile   string                 `json:"training_file"`
	ValidationFile string                 `json:"validation_file,omitempty"`
	Suffix         string                 `json:"suffix,omitempty"`
	Seed           int                    `json:"seed,omitempty"`
	Method         *FineTuningMethod      `json:"method,omitempty"`
	Integrations   []IntegrationRequest   `json:"integrations,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	// Legacy support via method normally, but API might accept hyperparameters at top level (deprecated)
	// We will stick to "method" as much as possible for new Framework resource.
}

type FineTuningMethod struct {
	Type       string            `json:"type"`
	Supervised *SupervisedMethod `json:"supervised,omitempty"`
	DPO        *DPOMethod        `json:"dpo,omitempty"`
}

type SupervisedMethod struct {
	Hyperparameters *SupervisedHyperparameters `json:"hyperparameters,omitempty"`
}

type SupervisedHyperparameters struct {
	NEpochs                interface{} `json:"n_epochs,omitempty"`
	BatchSize              interface{} `json:"batch_size,omitempty"`
	LearningRateMultiplier interface{} `json:"learning_rate_multiplier,omitempty"`
}

type DPOMethod struct {
	Hyperparameters *DPOHyperparameters `json:"hyperparameters,omitempty"`
}

type DPOHyperparameters struct {
	Beta                   interface{} `json:"beta,omitempty"` // Can be "auto" or float
	NEpochs                interface{} `json:"n_epochs,omitempty"`
	BatchSize              interface{} `json:"batch_size,omitempty"`
	LearningRateMultiplier interface{} `json:"learning_rate_multiplier,omitempty"`
}

type IntegrationRequest struct {
	Type  string            `json:"type"`
	WandB *WandBIntegration `json:"wandb,omitempty"`
}

type WandBIntegration struct {
	Project string   `json:"project"`
	Name    string   `json:"name,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}
