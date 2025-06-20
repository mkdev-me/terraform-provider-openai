package main

import (
	"github.com/fjcorp/terraform-provider-openai/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// main is the entry point for the Terraform provider.
// It uses the Terraform plugin SDK to serve the provider implementation.
func main() {
	// Serve the provider
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.Provider,
	})
}
