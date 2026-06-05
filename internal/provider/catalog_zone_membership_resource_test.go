package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckCatalogZoneMembershipDestroy(zoneName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}

		opts, err := c.GetZoneOptions(zoneName)
		if err != nil {
			return nil
		}

		if catalog, ok := opts["catalog"].(string); ok && catalog != "" {
			return fmt.Errorf("zone %q still has catalog membership %q after destroy", zoneName, catalog)
		}

		return nil
	}
}

func TestAccCatalogZoneMembershipResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckCatalogZoneMembershipDestroy("test-member.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogZoneMembershipConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_catalog_zone_membership.test", "zone", "test-member.example"),
					resource.TestCheckResourceAttr("technitium_catalog_zone_membership.test", "catalog_zone", "test-catalog.example"),
				),
			},
		},
	})
}

func TestAccCatalogZoneMembershipResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckCatalogZoneMembershipDestroy("test-member-upd.example"),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_dns_zone" "catalog1" {
  name = "test-catalog-upd1.example"
  type = "Catalog"
}

resource "technitium_dns_zone" "catalog2" {
  name = "test-catalog-upd2.example"
  type = "Catalog"
}

resource "technitium_dns_zone" "member" {
  name = "test-member-upd.example"
  type = "Primary"
}

resource "technitium_catalog_zone_membership" "test" {
  zone         = technitium_dns_zone.member.name
  catalog_zone = technitium_dns_zone.catalog1.name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_catalog_zone_membership.test", "catalog_zone", "test-catalog-upd1.example"),
				),
			},
			{
				Config: `
provider "technitium" {}

resource "technitium_dns_zone" "catalog1" {
  name = "test-catalog-upd1.example"
  type = "Catalog"
}

resource "technitium_dns_zone" "catalog2" {
  name = "test-catalog-upd2.example"
  type = "Catalog"
}

resource "technitium_dns_zone" "member" {
  name = "test-member-upd.example"
  type = "Primary"
}

resource "technitium_catalog_zone_membership" "test" {
  zone         = technitium_dns_zone.member.name
  catalog_zone = technitium_dns_zone.catalog2.name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_catalog_zone_membership.test", "catalog_zone", "test-catalog-upd2.example"),
				),
			},
		},
	})
}

func TestAccCatalogZoneMembershipResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckCatalogZoneMembershipDestroy("test-member.example"),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogZoneMembershipConfig(),
			},
			{
				ResourceName:                         "technitium_catalog_zone_membership.test",
				ImportState:                          true,
				ImportStateId:                        "test-member.example",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "zone",
			},
		},
	})
}

func testAccCatalogZoneMembershipConfig() string {
	return `
provider "technitium" {}

resource "technitium_dns_zone" "catalog" {
  name = "test-catalog.example"
  type = "Catalog"
}

resource "technitium_dns_zone" "member" {
  name = "test-member.example"
  type = "Primary"
}

resource "technitium_catalog_zone_membership" "test" {
  zone         = technitium_dns_zone.member.name
  catalog_zone = technitium_dns_zone.catalog.name
}
`
}
