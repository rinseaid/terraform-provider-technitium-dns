package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDHCPReservedLeaseResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPReservedLeaseDestroy("test-lease-basic", "00:11:22:33:44:01"),
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPReservedLeaseConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "scope_name", "test-lease-basic"),
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "hardware_address", "00:11:22:33:44:01"),
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "address", "10.110.0.20"),
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "id", "test-lease-basic:00:11:22:33:44:01"),
				),
			},
		},
	})
}

func TestAccDHCPReservedLeaseResource_withOptional(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPReservedLeaseDestroy("test-lease-opts", "00:11:22:33:44:02"),
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPReservedLeaseConfig_withOptional(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "scope_name", "test-lease-opts"),
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "hardware_address", "00:11:22:33:44:02"),
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "address", "10.111.0.20"),
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "hostname", "test-host"),
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "comments", "acceptance test reservation"),
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "id", "test-lease-opts:00:11:22:33:44:02"),
				),
			},
		},
	})
}

func TestAccDHCPReservedLeaseResource_updateAddress(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPReservedLeaseDestroy("test-lease-upd", "00:11:22:33:44:03"),
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPReservedLeaseConfig_updateAddress_initial(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "address", "10.112.0.20"),
				),
			},
			{
				Config: testAccDHCPReservedLeaseConfig_updateAddress_modified(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_reserved_lease.test", "address", "10.112.0.30"),
				),
			},
		},
	})
}

func TestAccDHCPReservedLeaseResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPReservedLeaseDestroy("test-lease-imp", "00:11:22:33:44:04"),
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPReservedLeaseConfig_import(),
			},
			{
				ResourceName:      "technitium_dhcp_reserved_lease.test",
				ImportState:       true,
				ImportStateId:     "test-lease-imp:00:11:22:33:44:04",
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDHCPReservedLeaseDestroy(scopeName, mac string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}

		scopeData, err := c.GetDHCPScope(scopeName)
		if err != nil {
			// Scope itself was deleted, so the lease is gone too.
			return nil
		}

		reservedLeases, ok := scopeData["reservedLeases"].([]interface{})
		if !ok {
			return nil
		}

		normalizedMAC := normalizeMAC(mac)
		for _, entry := range reservedLeases {
			lease := entry.(map[string]interface{})
			if normalizeMAC(lease["hardwareAddress"].(string)) == normalizedMAC {
				return fmt.Errorf("DHCP reserved lease %s:%s still exists after destroy", scopeName, mac)
			}
		}

		return nil
	}
}

func testAccDHCPReservedLeaseConfig_basic() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name          = "test-lease-basic"
  start_address = "10.110.0.10"
  end_address   = "10.110.0.50"
  subnet_mask   = "255.255.255.0"
}

resource "technitium_dhcp_reserved_lease" "test" {
  scope_name       = technitium_dhcp_scope.test.name
  hardware_address = "00:11:22:33:44:01"
  address          = "10.110.0.20"
}
`
}

func testAccDHCPReservedLeaseConfig_withOptional() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name          = "test-lease-opts"
  start_address = "10.111.0.10"
  end_address   = "10.111.0.50"
  subnet_mask   = "255.255.255.0"
}

resource "technitium_dhcp_reserved_lease" "test" {
  scope_name       = technitium_dhcp_scope.test.name
  hardware_address = "00:11:22:33:44:02"
  address          = "10.111.0.20"
  hostname         = "test-host"
  comments         = "acceptance test reservation"
}
`
}

func testAccDHCPReservedLeaseConfig_updateAddress_initial() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name          = "test-lease-upd"
  start_address = "10.112.0.10"
  end_address   = "10.112.0.50"
  subnet_mask   = "255.255.255.0"
}

resource "technitium_dhcp_reserved_lease" "test" {
  scope_name       = technitium_dhcp_scope.test.name
  hardware_address = "00:11:22:33:44:03"
  address          = "10.112.0.20"
}
`
}

func testAccDHCPReservedLeaseConfig_updateAddress_modified() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name          = "test-lease-upd"
  start_address = "10.112.0.10"
  end_address   = "10.112.0.50"
  subnet_mask   = "255.255.255.0"
}

resource "technitium_dhcp_reserved_lease" "test" {
  scope_name       = technitium_dhcp_scope.test.name
  hardware_address = "00:11:22:33:44:03"
  address          = "10.112.0.30"
}
`
}

func testAccDHCPReservedLeaseConfig_import() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name          = "test-lease-imp"
  start_address = "10.113.0.10"
  end_address   = "10.113.0.50"
  subnet_mask   = "255.255.255.0"
}

resource "technitium_dhcp_reserved_lease" "test" {
  scope_name       = technitium_dhcp_scope.test.name
  hardware_address = "00:11:22:33:44:04"
  address          = "10.113.0.20"
}
`
}
