package provider

import (
	"os"
	"testing"

	"github.com/mkdev-me/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testAccProvider is the Terraform provider used for tests
var testAccProvider *schema.Provider

// testAccProviderFactories is a map of providers used for acceptance tests
var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	testAccProvider = Provider()
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"openai": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
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
