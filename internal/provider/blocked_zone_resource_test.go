package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccBlockedZoneResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckBlockedZoneDestroy("block-basic-test.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccBlockedZoneConfig("block-basic-test.example"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_blocked_zone.test", "domain", "block-basic-test.example"),
				),
			},
		},
	})
}

func TestAccBlockedZoneResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckBlockedZoneDestroy("block-import-test.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccBlockedZoneConfig("block-import-test.example"),
			},
			{
				ResourceName:                         "technitium_blocked_zone.test",
				ImportState:                          true,
				ImportStateId:                        "block-import-test.example",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "domain",
			},
		},
	})
}

func TestAccBlockedZonesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckBlockedZoneDestroy("block-datasource-test.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccBlockedZonesDataSourceConfig("block-datasource-test.example"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_blocked_zone.test", "domain", "block-datasource-test.example"),
					resource.TestCheckResourceAttrSet("data.technitium_blocked_zones.all", "zones.#"),
				),
			},
		},
	})
}

func testAccBlockedZoneConfig(domain string) string {
	return fmt.Sprintf(`
provider "technitium" {}

resource "technitium_blocked_zone" "test" {
  domain = %q
}
`, domain)
}

func testAccBlockedZonesDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
provider "technitium" {}

resource "technitium_blocked_zone" "test" {
  domain = %q
}

data "technitium_blocked_zones" "all" {
  depends_on = [technitium_blocked_zone.test]
}
`, domain)
}

func testAccCheckBlockedZoneDestroy(domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}

		result, err := c.ListBlockedZones("")
		if err != nil {
			return fmt.Errorf("error checking blocked zone destroy: %s", err)
		}

		if zoneList, ok := result["zones"].([]interface{}); ok {
			for _, entry := range zoneList {
				if z, ok := entry.(map[string]interface{}); ok {
					if name, ok := z["name"].(string); ok && name == domain {
						return fmt.Errorf("blocked zone %q still exists after destroy", domain)
					}
				}
			}
		}

		return nil
	}
}
