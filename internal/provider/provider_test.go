package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/mkdev-me/terraform-provider-openai/internal/client"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"openai": providerserver.NewProtocol6WithError(NewFrameworkProvider("test")()),
}

func init() {
	// Provider initialization is now handled in testAccProtoV6ProviderFactories
}

// testAccPreCheck ensures the necessary environment variables are set for acceptance tests
func testAccPreCheck(t *testing.T) {
	// Verify that required environment variables are set for acceptance tests
	if v := os.Getenv("OPENAI_API_KEY"); v == "" {
		t.Fatal("OPENAI_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("OPENAI_ORGANIZATION_ID"); v == "" {
		t.Fatal("OPENAI_ORGANIZATION_ID must be set for acceptance tests")
	}
}

// testClient returns a client for use in unit tests
func testClient() *client.OpenAIClient {
	return client.NewClient(
		os.Getenv("OPENAI_API_KEY"),
		os.Getenv("OPENAI_ORGANIZATION_ID"),
		"https://api.openai.com/v1",
	)
}
