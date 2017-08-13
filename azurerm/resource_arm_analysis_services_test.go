package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMAnalysisServices_basic(t *testing.T) {
	resourceName := "azurerm_analysis_services.test"
	ri := acctest.RandInt()
	config := testAccAzureRMAnalysisServices_basic(ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMAnalysisServicesDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMAnalysisServicesExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureRMAnalysisServices_prod(t *testing.T) {
	ctxProviders := make(map[string]terraform.ResourceProviderFactory)

	// add any fixed providers
	for k, p := range testAccProviders {
		ctxProviders[k] = terraform.ResourceProviderFactoryFixed(p)
	}

	opts := terraform.ContextOpts{Providers: ctxProviders}

	// A single state variable to track the lifecycle, starting with no state
	var state *terraform.State

	ri := acctest.RandInt()
	config := testAccAzureRMAnalysisServices_basic(ri)
	step := resource.TestStep{
		Config: config,
	}

	resource.ApplyConfig(opts, state, step)
}

func testCheckAzureRMAnalysisServicesExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		serverName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for Analysis Server: %s", serverName)
		}

		conn := testAccProvider.Meta().(*ArmClient).analysisClient

		resp, err := conn.GetDetails(resourceGroup, serverName)
		if err != nil {
			return fmt.Errorf("Bad: Get on analysisClient: %s", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Analysis Server %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMAnalysisServicesDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).analysisClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_analysis_services" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.GetDetails(resourceGroup, name)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Analysis Server still exists:\n%#v", resp.ServerProperties)
		}
	}

	return nil
}

func testAccAzureRMAnalysisServices_basic(rInt int) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
    name = "acctestRG-%d"
    location = "West US"
}

resource "azurerm_analysis_services" "test" {
    name = "acctestavset%d"
    location = "${azurerm_resource_group.test.location}"
	resource_group_name = "${azurerm_resource_group.test.name}"
	location            = "West US"
    sku_name            = "B1"
    sku_tier            = "Basic"
}
`, rInt, rInt)
}
