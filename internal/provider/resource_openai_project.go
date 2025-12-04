package provider

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceOpenAIProject defines the schema and CRUD operations for OpenAI projects.
// This resource allows users to manage OpenAI projects through Terraform, including
// creation, reading, updating, and deletion of projects, as well as importing existing ones.
func resourceOpenAIProject() *schema.Resource {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WARN] Recovered from panic in resourceOpenAIProject: %v", r)
		}
	}()

	resource := &schema.Resource{
		CreateContext: resourceOpenAIProjectCreate,
		ReadContext:   resourceOpenAIProjectRead,
		UpdateContext: resourceOpenAIProjectUpdate,
		DeleteContext: resourceOpenAIProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpenAIProjectImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the project",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the project was created",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the project (active, archived, etc.)",
			},
			"archived_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the project was archived, if applicable",
			},
		},
	}
	return resource
}

// resourceOpenAIProjectCreate handles the creation of a new OpenAI project.
func resourceOpenAIProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the client from the provider meta
	c, err := GetOpenAIClientWithAdminKey(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)

	log.Printf("[DEBUG] Creating OpenAI project with name: %s", name)

	// Create the project using the OpenAI API
	project, err := c.CreateProject(name)
	if err != nil {
		return diag.Errorf("error creating project: %s", err)
	}

	// Set ID
	d.SetId(project.ID)

	return resourceOpenAIProjectRead(ctx, d, meta)
}

// resourceOpenAIProjectRead retrieves the current state of an OpenAI project.
func resourceOpenAIProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the client from the provider meta
	c, err := GetOpenAIClientWithAdminKey(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Id()
	log.Printf("[DEBUG] Reading OpenAI project with ID: %s", projectID)

	// Get the project using the OpenAI API
	project, err := c.GetProject(projectID)
	if err != nil {
		return diag.Errorf("error reading project: %s", err)
	}

	log.Printf("[DEBUG] Successfully retrieved project from API: %s (status: %s)", project.Name, project.Status)

	// Set basic fields
	if err := d.Set("name", project.Name); err != nil {
		return diag.Errorf("error setting name: %s", err)
	}

	if project.Status != "" {
		if err := d.Set("status", project.Status); err != nil {
			return diag.Errorf("error setting status: %s", err)
		}
		log.Printf("[DEBUG] Set status to: %s", project.Status)
	}

	// Handle Unix timestamps for created_at and archived_at
	if project.CreatedAt != nil {
		createdTime := time.Unix(int64(*project.CreatedAt), 0)
		if err := d.Set("created_at", createdTime.Format(time.RFC3339)); err != nil {
			return diag.Errorf("error setting created_at: %s", err)
		}
		log.Printf("[DEBUG] Set created_at to: %s", createdTime.Format(time.RFC3339))
	}

	if project.ArchivedAt != nil {
		archivedTime := time.Unix(int64(*project.ArchivedAt), 0)
		if err := d.Set("archived_at", archivedTime.Format(time.RFC3339)); err != nil {
			return diag.Errorf("error setting archived_at: %s", err)
		}
		log.Printf("[DEBUG] Set archived_at to: %s", archivedTime.Format(time.RFC3339))
	}

	log.Printf("[DEBUG] OpenAI project read complete for ID: %s", projectID)
	return nil
}

// resourceOpenAIProjectUpdate modifies an existing OpenAI project.
func resourceOpenAIProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the client from the provider meta
	c, err := GetOpenAIClientWithAdminKey(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)

	log.Printf("[DEBUG] Updating OpenAI project with ID: %s, name: %s", d.Id(), name)

	// Update the project using the OpenAI API
	// Note: The API uses POST for updates, not PATCH
	_, err = c.UpdateProject(d.Id(), name)
	if err != nil {
		return diag.Errorf("error updating project: %s", err)
	}

	return resourceOpenAIProjectRead(ctx, d, meta)
}

// resourceOpenAIProjectDelete removes an OpenAI project.
func resourceOpenAIProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the client from the provider meta
	c, err := GetOpenAIClientWithAdminKey(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting (archiving) OpenAI project with ID: %s", d.Id())

	// Delete the project using the OpenAI API
	// Note: This actually archives the project by setting status to "archived"
	// and requires the project name to be included in the request
	if err := c.DeleteProject(d.Id()); err != nil {
		return diag.Errorf("error deleting project: %s", err)
	}

	d.SetId("")
	return nil
}

// resourceOpenAIProjectImport imports an existing OpenAI project into Terraform state.
func resourceOpenAIProjectImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Get the client from the provider meta
	c, err := GetOpenAIClientWithAdminKey(meta)
	if err != nil {
		return nil, err
	}

	projectID := d.Id()
	log.Printf("[DEBUG] Importing OpenAI project with ID: %s", projectID)

	// Get the project using the OpenAI API
	project, err := c.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("error reading project during import: %s", err)
	}

	log.Printf("[DEBUG] Successfully retrieved project from API: %s (status: %s)", project.Name, project.Status)

	// Set all fields
	if err := d.Set("name", project.Name); err != nil {
		return nil, fmt.Errorf("error setting name: %s", err)
	}

	if project.CreatedAt != nil {
		createdTime := time.Unix(int64(*project.CreatedAt), 0)
		if err := d.Set("created_at", createdTime.Format(time.RFC3339)); err != nil {
			return nil, fmt.Errorf("error setting created_at: %s", err)
		}
		log.Printf("[DEBUG] Set created_at to: %s", createdTime.Format(time.RFC3339))
	}

	if project.Status != "" {
		if err := d.Set("status", project.Status); err != nil {
			return nil, fmt.Errorf("error setting status: %s", err)
		}
		log.Printf("[DEBUG] Set status to: %s", project.Status)
	}

	if project.ArchivedAt != nil {
		archivedTime := time.Unix(int64(*project.ArchivedAt), 0)
		if err := d.Set("archived_at", archivedTime.Format(time.RFC3339)); err != nil {
			return nil, fmt.Errorf("error setting archived_at: %s", err)
		}
		log.Printf("[DEBUG] Set archived_at to: %s", archivedTime.Format(time.RFC3339))
	}

	log.Printf("[DEBUG] OpenAI project import complete for ID: %s", projectID)
	return []*schema.ResourceData{d}, nil
}
