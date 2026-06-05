package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDNSSettingsResource_import verifies that the singleton dns_settings resource
// can be imported using any arbitrary ID (the resource always resolves to "settings").
func TestAccDNSSettingsResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_dns_settings" "test" {
  dns_server_domain = "test.local"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "id", "settings"),
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "dns_server_domain", "test.local"),
				),
			},
			{
				ResourceName:      "technitium_dns_settings.test",
				ImportState:       true,
				ImportStateId:     "settings",
				ImportStateVerify: true,
				// These fields are write-only or not returned by the API, so they will
				// never round-trip cleanly through import.
				ImportStateVerifyIgnore: []string{
					"dns_tls_certificate_password",
					"web_service_tls_certificate_password",
					"server_proxy_password",
				},
			},
		},
	})
}

// TestAccZoneDnssecResource_update verifies the unsign+re-sign update cycle works
// when changing algorithm parameters (ECDSA/P256/NSEC → ECDSA/P384/NSEC3).
func TestAccZoneDnssecResource_update(t *testing.T) {
	zoneName := "dnssec-update-test.example"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckZoneDnssecDestroy(zoneName),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDnssecUpdateConfig(zoneName, "P256", "NSEC"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "zone", zoneName),
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "algorithm", "ECDSA"),
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "curve", "P256"),
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "nx_proof", "NSEC"),
					resource.TestCheckResourceAttrSet("technitium_zone_dnssec.test", "dnssec_status"),
				),
			},
			{
				Config: testAccZoneDnssecUpdateConfig(zoneName, "P384", "NSEC3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "zone", zoneName),
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "algorithm", "ECDSA"),
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "curve", "P384"),
					resource.TestCheckResourceAttr("technitium_zone_dnssec.test", "nx_proof", "NSEC3"),
					resource.TestCheckResourceAttrSet("technitium_zone_dnssec.test", "dnssec_status"),
				),
			},
		},
	})
}

func testAccZoneDnssecUpdateConfig(zone, curve, nxProof string) string {
	return fmt.Sprintf(`
provider "technitium" {}

resource "technitium_dns_zone" "test" {
  name = %[1]q
  type = "Primary"
}

resource "technitium_zone_dnssec" "test" {
  zone      = technitium_dns_zone.test.name
  algorithm = "ECDSA"
  curve     = %[2]q
  nx_proof  = %[3]q
}
`, zone, curve, nxProof)
}

// TestAccDHCPScopeResource_advancedFieldsFull verifies that a scope configured with
// exclusions, static_routes, ntp_servers, and vendor_info round-trips correctly.
func TestAccDHCPScopeResource_advancedFieldsFull(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPScopeDestroy("test-adv-full"),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name           = "test-adv-full"
  start_address  = "192.168.50.20"
  end_address    = "192.168.50.200"
  subnet_mask    = "255.255.255.0"
  router_address = "192.168.50.1"

  exclusions = "192.168.50.100|192.168.50.110"

  # Route 10.0.0.0/8 via the gateway
  static_routes = "10.0.0.0|255.0.0.0|192.168.50.1"

  # NTP servers specified as IP addresses
  ntp_servers = ["192.168.50.1", "192.168.50.2"]

  # Vendor info: class identifier + hex-encoded specific information
  vendor_info = "MSFT 5.0|01:04:00:00:00:00"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "name", "test-adv-full"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "exclusions", "192.168.50.100|192.168.50.110"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "static_routes", "10.0.0.0|255.0.0.0|192.168.50.1"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "ntp_servers.#", "2"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "ntp_servers.0", "192.168.50.1"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "ntp_servers.1", "192.168.50.2"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "vendor_info", "MSFT 5.0|01:04:00:00:00:00"),
				),
			},
		},
	})
}

// TestAccAdminUserResource_memberOfGroups creates a group first, then adds a user
// with member_of_groups pointing to it. Two-step: create group, then create user
// with membership.
func TestAccAdminUserResource_memberOfGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckAdminUserDestroy("tftest-grouped-user"),
			testAccCheckAdminGroupDestroy("tftest-user-group"),
		),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_admin_group" "test" {
  name        = "tftest-user-group"
  description = "Group for member_of_groups test"
}
`,
			},
			{
				Config: `
provider "technitium" {}

resource "technitium_admin_user" "test" {
  username         = "tftest-grouped-user"
  password         = "TestPass123!"
  display_name     = "Grouped User"
  member_of_groups = "tftest-user-group"
}

resource "technitium_admin_group" "test" {
  name        = "tftest-user-group"
  description = "Group for member_of_groups test"
  members     = "tftest-grouped-user"

  depends_on = [technitium_admin_user.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_admin_user.test", "member_of_groups", "tftest-user-group"),
					resource.TestCheckResourceAttr("technitium_admin_group.test", "members", "tftest-grouped-user"),
				),
			},
		},
	})
}

// TestAccAdminGroupResource_withMembers creates a user first, then creates a group
// with that user as a member. Two-step to avoid circular dependency.
func TestAccAdminGroupResource_withMembers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckAdminGroupDestroy("tftest-members-group"),
			testAccCheckAdminUserDestroy("tftest-member-user"),
		),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_admin_user" "test" {
  username     = "tftest-member-user"
  password     = "TestPass123!"
  display_name = "Member User"
}
`,
			},
			{
				Config: `
provider "technitium" {}

resource "technitium_admin_user" "test" {
  username         = "tftest-member-user"
  password         = "TestPass123!"
  display_name     = "Member User"
  member_of_groups = "tftest-members-group"
}

resource "technitium_admin_group" "test" {
  name        = "tftest-members-group"
  description = "Group for members test"
  members     = "tftest-member-user"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_admin_group.test", "members", "tftest-member-user"),
					resource.TestCheckResourceAttr("technitium_admin_user.test", "member_of_groups", "tftest-members-group"),
				),
			},
		},
	})
}
