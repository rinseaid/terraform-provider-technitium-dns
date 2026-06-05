package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccZoneDnssecResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckZoneDnssecDestroy("dnssec-test.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDnssecConfig("dnssec-test.example"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "zone", "dnssec-test.example"),
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "algorithm", "ECDSA"),
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "curve", "P256"),
					resource.TestCheckResourceAttrSet("technitium_zone_dnssec.test", "dnssec_status"),
				),
			},
		},
	})
}

func TestAccZoneDnssecResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckZoneDnssecDestroy("dnssec-import-test.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDnssecConfig("dnssec-import-test.example"),
			},
			{
				ResourceName:                         "technitium_zone_dnssec.test",
				ImportState:                          true,
				ImportStateId:                        "dnssec-import-test.example",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "zone",
			},
		},
	})
}

func testAccZoneDnssecConfig(zone string) string {
	return fmt.Sprintf(`
provider "technitium" {}

resource "technitium_dns_zone" "test" {
  name = %[1]q
  type = "Primary"
}

resource "technitium_zone_dnssec" "test" {
  zone      = technitium_dns_zone.test.name
  algorithm = "ECDSA"
  curve     = "P256"
  nx_proof  = "NSEC"
}
`, zone)
}

func testAccCheckZoneDnssecDestroy(zone string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}

		opts, err := c.GetZoneOptions(zone)
		if err != nil {
			return nil
		}

		status, _ := opts["dnssecStatus"].(string)
		if status != "" && status != "Unsigned" {
			return fmt.Errorf("zone %q still has DNSSEC status %q after destroy", zone, status)
		}

		return nil
	}
}
