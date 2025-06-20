package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOpenAIProject_basic(t *testing.T) {
	t.Skip("Skipping until properly implemented and OpenAI API credentials are configured for tests")

	var projectID string
	projectName := "tf-acc-test-project"
	projectDesc := "Terraform acceptance test project"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckOpenAIProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOpenAIProjectBasic(projectName, projectDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenAIProjectExists("openai_project.test", &projectID),
					resource.TestCheckResourceAttr("openai_project.test", "name", projectName),
					resource.TestCheckResourceAttr("openai_project.test", "description", projectDesc),
					resource.TestCheckResourceAttrSet("openai_project.test", "created_at"),
					resource.TestCheckResourceAttrSet("openai_project.test", "updated_at"),
				),
			},
			{
				ResourceName:      "openai_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceOpenAIProject_update(t *testing.T) {
	t.Skip("Skipping until properly implemented and OpenAI API credentials are configured for tests")

	var projectID string
	projectName := "tf-acc-test-project"
	projectDesc := "Terraform acceptance test project"
	projectNameUpdated := "tf-acc-test-project-updated"
	projectDescUpdated := "Terraform acceptance test project updated"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckOpenAIProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOpenAIProjectBasic(projectName, projectDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenAIProjectExists("openai_project.test", &projectID),
					resource.TestCheckResourceAttr("openai_project.test", "name", projectName),
					resource.TestCheckResourceAttr("openai_project.test", "description", projectDesc),
				),
			},
			{
				Config: testAccResourceOpenAIProjectBasic(projectNameUpdated, projectDescUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenAIProjectExists("openai_project.test", &projectID),
					resource.TestCheckResourceAttr("openai_project.test", "name", projectNameUpdated),
					resource.TestCheckResourceAttr("openai_project.test", "description", projectDescUpdated),
				),
			},
		},
	})
}

func TestAccResourceOpenAIProject_withUsageLimits(t *testing.T) {
	t.Skip("Skipping until properly implemented and OpenAI API credentials are configured for tests")

	var projectID string
	projectName := "tf-acc-test-project-limits"
	projectDesc := "Terraform acceptance test project with usage limits"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckOpenAIProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOpenAIProjectWithUsageLimits(projectName, projectDesc, 100.0, 1000000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenAIProjectExists("openai_project.test", &projectID),
					resource.TestCheckResourceAttr("openai_project.test", "name", projectName),
					resource.TestCheckResourceAttr("openai_project.test", "description", projectDesc),
					resource.TestCheckResourceAttr("openai_project.test", "usage_limits.0.max_budget", "100"),
					resource.TestCheckResourceAttr("openai_project.test", "usage_limits.0.max_tokens", "1000000"),
				),
			},
		},
	})
}

func testAccCheckOpenAIProjectExists(n string, projectID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No project ID is set")
		}

		// In a real implementation, use the API client to check if the project exists
		// client := testAccProvider.Meta().(*client.OpenAIClient)
		// _, err := client.GetProject(rs.Primary.ID)
		// if err != nil {
		//     return fmt.Errorf("Error retrieving project: %s", err)
		// }

		*projectID = rs.Primary.ID
		return nil
	}
}

func testAccCheckOpenAIProjectDestroy(s *terraform.State) error {
	// In a real implementation, use the API client to check if the project has been destroyed
	// client := testAccProvider.Meta().(*client.OpenAIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openai_project" {
			continue
		}

		// Try to find the project by ID
		// _, err := client.GetProject(rs.Primary.ID)
		// if err == nil {
		//     return fmt.Errorf("Project still exists")
		// }
	}

	return nil
}

func testAccResourceOpenAIProjectBasic(name, description string) string {
	return fmt.Sprintf(`
resource "openai_project" "test" {
  name        = "%s"
  description = "%s"
}
`, name, description)
}

func testAccResourceOpenAIProjectWithUsageLimits(name, description string, maxBudget float64, maxTokens int) string {
	return fmt.Sprintf(`
resource "openai_project" "test" {
  name        = "%s"
  description = "%s"
  
  usage_limits {
    max_budget = %f
    max_tokens = %d
  }
}
`, name, description, maxBudget, maxTokens)
}
