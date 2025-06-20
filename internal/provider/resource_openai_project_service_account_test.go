package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOpenAIProjectServiceAccount_basic(t *testing.T) {
	t.Skip("Skipping until properly implemented and OpenAI API credentials are configured for tests")

	var serviceAccountID string
	projectName := "tf-acc-test-project-svc"
	serviceAccountName := "tf-acc-test-service-account"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckOpenAIProjectServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOpenAIProjectServiceAccountBasic(projectName, serviceAccountName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenAIProjectServiceAccountExists("openai_project_service_account.test", &serviceAccountID),
					resource.TestCheckResourceAttr("openai_project_service_account.test", "name", serviceAccountName),
					resource.TestCheckResourceAttrSet("openai_project_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttrSet("openai_project_service_account.test", "created_at"),
				),
			},
			{
				ResourceName:      "openai_project_service_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceOpenAIProjectServiceAccount_withAPIKey(t *testing.T) {
	t.Skip("Skipping until properly implemented and OpenAI API credentials are configured for tests")

	var serviceAccountID string
	projectName := "tf-acc-test-project-svc-key"
	serviceAccountName := "tf-acc-test-service-account-with-key"
	apiKeyName := "tf-acc-test-api-key"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckOpenAIProjectServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOpenAIProjectServiceAccountWithAPIKey(projectName, serviceAccountName, apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenAIProjectServiceAccountExists("openai_project_service_account.test", &serviceAccountID),
					resource.TestCheckResourceAttr("openai_project_service_account.test", "name", serviceAccountName),
					resource.TestCheckResourceAttrSet("openai_project_service_account.test", "service_account_id"),
					resource.TestCheckResourceAttrSet("openai_project_service_account.test", "created_at"),
					resource.TestCheckResourceAttr("openai_project_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrSet("openai_project_api_key.test", "api_key_id"),
					resource.TestCheckResourceAttrSet("openai_project_api_key.test", "key"),
				),
			},
		},
	})
}

func TestAccResourceOpenAIProjectServiceAccount_moduleUsage(t *testing.T) {
	t.Skip("Skipping until properly implemented and OpenAI API credentials are configured for tests")

	var serviceAccountID string
	projectName := "tf-acc-test-project-svc-module"
	serviceAccountName := "tf-acc-test-module-service-account"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckOpenAIProjectServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOpenAIProjectServiceAccountModuleUsage(projectName, serviceAccountName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenAIProjectServiceAccountExistsModule("module.test_service_account", &serviceAccountID),
					resource.TestCheckResourceAttrSet("module.test_service_account.openai_project_service_account.this", "service_account_id"),
					resource.TestCheckResourceAttr("module.test_service_account.openai_project_service_account.this", "name", serviceAccountName),
					resource.TestCheckResourceAttrSet("module.test_service_account.openai_project_api_key.this[0]", "key"),
				),
			},
		},
	})
}

func testAccCheckOpenAIProjectServiceAccountExists(n string, serviceAccountID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No service account ID is set")
		}

		// In a real implementation, use the API client to check if the service account exists
		// client := testAccProvider.Meta().(*client.OpenAIClient)
		//
		// Extract project ID and service account ID from the composite ID
		// idParts := strings.Split(rs.Primary.ID, ":")
		// if len(idParts) != 2 {
		//     return fmt.Errorf("Invalid ID format: %s", rs.Primary.ID)
		// }
		// projectID := idParts[0]
		// saID := idParts[1]
		//
		// _, err := client.GetProjectServiceAccount(projectID, saID, "")
		// if err != nil {
		//     return fmt.Errorf("Error retrieving service account: %s", err)
		// }

		*serviceAccountID = rs.Primary.Attributes["service_account_id"]
		return nil
	}
}

func testAccCheckOpenAIProjectServiceAccountExistsModule(n string, serviceAccountID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// For modules, we need to look for the resources created by the module
		var saResource *terraform.ResourceState
		for _, res := range s.RootModule().Resources {
			if strings.HasSuffix(res.Type, "openai_project_service_account") {
				saResource = res
				break
			}
		}

		if saResource == nil {
			return fmt.Errorf("Service account resource not found in state")
		}

		if saResource.Primary.ID == "" {
			return fmt.Errorf("No service account ID is set")
		}

		// In a real implementation, verify with API call as in testAccCheckOpenAIProjectServiceAccountExists

		*serviceAccountID = saResource.Primary.Attributes["service_account_id"]
		return nil
	}
}

func testAccCheckOpenAIProjectServiceAccountDestroy(s *terraform.State) error {
	// In a real implementation, use the API client to check if the service account has been destroyed
	// client := testAccProvider.Meta().(*client.OpenAIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openai_project_service_account" {
			continue
		}

		// Extract project ID and service account ID from the composite ID
		// idParts := strings.Split(rs.Primary.ID, ":")
		// if len(idParts) != 2 {
		//     continue // Invalid ID format, skip this resource
		// }
		//
		// projectID := idParts[0]
		// saID := idParts[1]
		//
		// Try to find the service account by ID
		// _, err := client.GetProjectServiceAccount(projectID, saID, "")
		// if err == nil {
		//     return fmt.Errorf("Service account still exists")
		// }
	}

	return nil
}

func testAccResourceOpenAIProjectServiceAccountBasic(projectName, serviceAccountName string) string {
	return fmt.Sprintf(`
resource "openai_project" "test" {
  name        = "%s"
  description = "Terraform acceptance test project for service accounts"
}

resource "openai_project_service_account" "test" {
  project_id = openai_project.test.id
  name       = "%s"
}
`, projectName, serviceAccountName)
}

func testAccResourceOpenAIProjectServiceAccountWithAPIKey(projectName, serviceAccountName, apiKeyName string) string {
	return fmt.Sprintf(`
resource "openai_project" "test" {
  name        = "%s"
  description = "Terraform acceptance test project for service accounts with API key"
}

resource "openai_project_service_account" "test" {
  project_id = openai_project.test.id
  name       = "%s"
}

resource "openai_project_api_key" "test" {
  project_id         = openai_project.test.id
  name               = "%s"
  service_account_id = openai_project_service_account.test.service_account_id
}
`, projectName, serviceAccountName, apiKeyName)
}

func testAccResourceOpenAIProjectServiceAccountModuleUsage(projectName, serviceAccountName string) string {
	return fmt.Sprintf(`
resource "openai_project" "test" {
  name        = "%s"
  description = "Terraform acceptance test project for service account module"
}

module "test_service_account" {
  source     = "../../modules/service_account"
  project_id = openai_project.test.id
  name       = "%s"
}

output "service_account_id" {
  value = module.test_service_account.service_account_id
}

output "api_key" {
  value     = module.test_service_account.api_key
  sensitive = true
}
`, projectName, serviceAccountName)
}
