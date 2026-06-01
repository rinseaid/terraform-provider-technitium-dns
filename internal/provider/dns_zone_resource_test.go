package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDNSZoneResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:            testAccCheckDNSZoneDestroy("testzone-basic.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneConfig("testzone-basic.example", "Primary", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "name", "testzone-basic.example"),
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "type", "Primary"),
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "disabled", "false"),
					resource.TestCheckResourceAttrSet("technitium_dns_zone.test", "dnssec_status"),
				),
			},
		},
	})
}

func TestAccDNSZoneResource_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:            testAccCheckDNSZoneDestroy("testzone-disabled.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneConfig("testzone-disabled.example", "Primary", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "name", "testzone-disabled.example"),
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "type", "Primary"),
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "disabled", "true"),
				),
			},
		},
	})
}

func TestAccDNSZoneResource_updateDisabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:            testAccCheckDNSZoneDestroy("testzone-update.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneConfig("testzone-update.example", "Primary", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "name", "testzone-update.example"),
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "disabled", "false"),
				),
			},
			{
				Config: testAccDNSZoneConfig("testzone-update.example", "Primary", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "name", "testzone-update.example"),
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "disabled", "true"),
				),
			},
		},
	})
}

func TestAccDNSZoneResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:            testAccCheckDNSZoneDestroy("testzone-import.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneConfig("testzone-import.example", "Primary", false),
			},
			{
				ResourceName:                         "technitium_dns_zone.test",
				ImportState:                          true,
				ImportStateId:                        "testzone-import.example",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}

func TestAccDNSZonesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:            testAccCheckDNSZoneDestroy("testzone-datasource.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZonesDataSourceConfig("testzone-datasource.example"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "name", "testzone-datasource.example"),
					resource.TestCheckResourceAttrSet("data.technitium_dns_zones.all", "zones.#"),
				),
			},
		},
	})
}

func testAccDNSZoneConfig(name, zoneType string, disabled bool) string {
	return fmt.Sprintf(`
resource "technitium_dns_zone" "test" {
  name     = %q
  type     = %q
  disabled = %t
}
`, name, zoneType, disabled)
}

func testAccDNSZonesDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "technitium_dns_zone" "test" {
  name = %q
  type = "Primary"
}

data "technitium_dns_zones" "all" {
  depends_on = [technitium_dns_zone.test]
}
`, name)
}

func testAccCheckDNSZoneDestroy(zoneName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}

		_, err = c.GetZoneOptions(zoneName)
		if err == nil {
			return fmt.Errorf("DNS zone %q still exists after destroy", zoneName)
		}

		return nil
	}
}
