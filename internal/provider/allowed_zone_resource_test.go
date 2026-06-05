package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAllowedZoneResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAllowedZoneDestroy("allow-basic-test.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedZoneConfig("allow-basic-test.example"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_allowed_zone.test", "domain", "allow-basic-test.example"),
				),
			},
		},
	})
}

func TestAccAllowedZoneResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAllowedZoneDestroy("allow-import-test.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedZoneConfig("allow-import-test.example"),
			},
			{
				ResourceName:                         "technitium_allowed_zone.test",
				ImportState:                          true,
				ImportStateId:                        "allow-import-test.example",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "domain",
			},
		},
	})
}

func TestAccAllowedZonesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAllowedZoneDestroy("allow-datasource-test.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedZonesDataSourceConfig("allow-datasource-test.example"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_allowed_zone.test", "domain", "allow-datasource-test.example"),
					resource.TestCheckResourceAttrSet("data.technitium_allowed_zones.all", "zones.#"),
				),
			},
		},
	})
}

func testAccAllowedZoneConfig(domain string) string {
	return fmt.Sprintf(`
provider "technitium" {}

resource "technitium_allowed_zone" "test" {
  domain = %q
}
`, domain)
}

func testAccAllowedZonesDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
provider "technitium" {}

resource "technitium_allowed_zone" "test" {
  domain = %q
}

data "technitium_allowed_zones" "all" {
  depends_on = [technitium_allowed_zone.test]
}
`, domain)
}

func testAccCheckAllowedZoneDestroy(domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}

		result, err := c.ListAllowedZones("")
		if err != nil {
			return fmt.Errorf("error checking allowed zone destroy: %s", err)
		}

		if zoneList, ok := result["zones"].([]interface{}); ok {
			for _, entry := range zoneList {
				if z, ok := entry.(map[string]interface{}); ok {
					if name, ok := z["name"].(string); ok && name == domain {
						return fmt.Errorf("allowed zone %q still exists after destroy", domain)
					}
				}
			}
		}

		return nil
	}
}
