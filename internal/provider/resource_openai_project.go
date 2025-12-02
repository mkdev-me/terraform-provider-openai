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
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The title of the project",
			},
			"created": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Timestamp when the project was created",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the project (active, archived, etc.)",
			},
			"archived_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Timestamp when the project was archived, if applicable",
			},
			"is_initial": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this is the initial project",
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

	title := d.Get("title").(string)

	log.Printf("[DEBUG] Creating OpenAI project with title: %s", title)

	// Create the project using the OpenAI API
	project, err := c.CreateProject(title)
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

	log.Printf("[DEBUG] Successfully retrieved project from API: %s (status: %s)", project.Title, project.Status)

	// Set basic fields
	if err := d.Set("title", project.Title); err != nil {
		return diag.Errorf("error setting title: %s", err)
	}

	if err := d.Set("status", project.Status); err != nil {
		return diag.Errorf("error setting status: %s", err)
	}
	log.Printf("[DEBUG] Set status to: %s", project.Status)

	if err := d.Set("created", project.Created); err != nil {
		return diag.Errorf("error setting created: %s", err)
	}
	log.Printf("[DEBUG] Set created to: %d", project.Created)

	if project.ArchivedAt != nil {
		if err := d.Set("archived_at", *project.ArchivedAt); err != nil {
			return diag.Errorf("error setting archived_at: %s", err)
		}
		log.Printf("[DEBUG] Set archived_at to: %d", *project.ArchivedAt)
	}

	if err := d.Set("is_initial", project.IsInitial); err != nil {
		return diag.Errorf("error setting is_initial: %s", err)
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

	title := d.Get("title").(string)

	log.Printf("[DEBUG] Updating OpenAI project with ID: %s, title: %s", d.Id(), title)

	// Update the project using the OpenAI API
	_, err = c.UpdateProject(d.Id(), title)
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

	log.Printf("[DEBUG] Successfully retrieved project from API: %s (status: %s)", project.Title, project.Status)

	// Set all fields
	if err := d.Set("title", project.Title); err != nil {
		return nil, fmt.Errorf("error setting title: %s", err)
	}

	if err := d.Set("created", project.Created); err != nil {
		return nil, fmt.Errorf("error setting created: %s", err)
	}
	log.Printf("[DEBUG] Set created to: %d", project.Created)

	if err := d.Set("status", project.Status); err != nil {
		return nil, fmt.Errorf("error setting status: %s", err)
	}
	log.Printf("[DEBUG] Set status to: %s", project.Status)

	if project.ArchivedAt != nil {
		if err := d.Set("archived_at", *project.ArchivedAt); err != nil {
			return nil, fmt.Errorf("error setting archived_at: %s", err)
		}
		log.Printf("[DEBUG] Set archived_at to: %d", *project.ArchivedAt)
	}

	if err := d.Set("is_initial", project.IsInitial); err != nil {
		return nil, fmt.Errorf("error setting is_initial: %s", err)
	}

	log.Printf("[DEBUG] OpenAI project import complete for ID: %s", projectID)
	return []*schema.ResourceData{d}, nil
}
