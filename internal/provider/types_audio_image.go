package provider

// TranscriptionResponseFramework represents the API response for audio transcriptions.
type TranscriptionResponseFramework struct {
	Text     string             `json:"text"`
	Duration float64            `json:"duration,omitempty"`
	Segments []SegmentFramework `json:"segments,omitempty"`
	// For verbose_json
	Language string `json:"language,omitempty"`
	Task     string `json:"task,omitempty"`
}

// TranslationResponseFramework represents the API response for audio translations.
type TranslationResponseFramework struct {
	Text     string             `json:"text"`
	Duration float64            `json:"duration,omitempty"`
	Segments []SegmentFramework `json:"segments,omitempty"`
}

// SegmentFramework represents a single segment of the audio transcription/translation.
type SegmentFramework struct {
	ID               int     `json:"id"`
	Seek             int     `json:"seek"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	Tokens           []int   `json:"tokens"`
	Temperature      float64 `json:"temperature"`
	AvgLogprob       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
}

// TextToSpeechRequest represents the request payload for text-to-speech conversion.
type TextToSpeechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
	Instructions   string  `json:"instructions,omitempty"`
}

// ImageGenerationResponseFramework represents the API response for image generation.
type ImageGenerationResponseFramework struct {
	Created int64                          `json:"created"`
	Data    []ImageGenerationDataFramework `json:"data"`
}

// ImageGenerationDataFramework represents a single generated image.
type ImageGenerationDataFramework struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// ImageGenerationRequest represents the request payload for generating images.
type ImageGenerationRequest struct {
	Model          string `json:"model,omitempty"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	Size           string `json:"size,omitempty"`
	Style          string `json:"style,omitempty"`
	User           string `json:"user,omitempty"`
}

// ImageEditResponseFramework represents the API response for image editing.
type ImageEditResponseFramework struct {
	Created int64                    `json:"created"`
	Data    []ImageEditDataFramework `json:"data"`
}

// ImageEditDataFramework represents a single edited image.
type ImageEditDataFramework struct {
	URL     string `json:"url,omitempty"`
	B64JSON string `json:"b64_json,omitempty"`
}

// ImageVariationResponseFramework represents the API response for image variations.
type ImageVariationResponseFramework struct {
	Created int64                         `json:"created"`
	Data    []ImageVariationDataFramework `json:"data"`
}

// ImageVariationDataFramework represents a single image variation.
type ImageVariationDataFramework struct {
	URL     string `json:"url,omitempty"`
	B64JSON string `json:"b64_json,omitempty"`
}
