package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceOpenAITextToSpeech provides a data source to retrieve OpenAI text-to-speech details.
// This data source verifies the existence of a previously generated speech file.
func dataSourceOpenAITextToSpeech() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpenAITextToSpeechRead,
		Schema: map[string]*schema.Schema{
			"file_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path to the text-to-speech audio file to verify.",
			},
			"exists": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the speech file exists.",
			},
			"file_size_bytes": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size of the speech file in bytes.",
			},
			"last_modified": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp when the speech file was last modified.",
			},
		},
	}
}

// dataSourceOpenAITextToSpeechRead reads information about an existing OpenAI text-to-speech file.
// It verifies the file exists and returns metadata about it.
func dataSourceOpenAITextToSpeechRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	filePath := d.Get("file_path").(string)

	// Check if the file exists and get stats
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, but return valid response with exists=false
			d.SetId(fmt.Sprintf("tts_file_%d", time.Now().Unix()))
			if err := d.Set("exists", false); err != nil {
				return diag.FromErr(fmt.Errorf("error setting 'exists' attribute: %v", err))
			}
			if err := d.Set("file_size_bytes", 0); err != nil {
				return diag.FromErr(fmt.Errorf("error setting 'file_size_bytes' attribute: %v", err))
			}
			if err := d.Set("last_modified", 0); err != nil {
				return diag.FromErr(fmt.Errorf("error setting 'last_modified' attribute: %v", err))
			}
			return diags
		}
		// Other errors
		return diag.FromErr(fmt.Errorf("error accessing file %s: %v", filePath, err))
	}

	// File exists, set values
	d.SetId(fmt.Sprintf("tts_file_%d", time.Now().Unix()))

	if err := d.Set("exists", true); err != nil {
		return diag.FromErr(fmt.Errorf("error setting 'exists' attribute: %v", err))
	}

	if err := d.Set("file_size_bytes", fileInfo.Size()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting 'file_size_bytes' attribute: %v", err))
	}

	if err := d.Set("last_modified", fileInfo.ModTime().Unix()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting 'last_modified' attribute: %v", err))
	}

	return diags
}
