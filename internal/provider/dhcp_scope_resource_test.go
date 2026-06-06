package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDHCPScopeResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPScopeDestroy("test-basic"),
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPScopeConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "name", "test-basic"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "start_address", "10.100.0.10"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "end_address", "10.100.0.50"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "subnet_mask", "255.255.255.0"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "router_address", "10.100.0.1"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "enabled", "true"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "lease_time", "86400"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "offer_delay", "0"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "ping_check", "false"),
				),
			},
		},
	})
}

func TestAccDHCPScopeResource_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPScopeDestroy("test-full"),
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPScopeConfig_full(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "name", "test-full"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "enabled", "true"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "start_address", "10.101.0.10"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "end_address", "10.101.0.100"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "subnet_mask", "255.255.255.0"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "router_address", "10.101.0.1"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "dns_servers.#", "2"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "dns_servers.0", "10.101.0.1"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "dns_servers.1", "1.1.1.1"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "domain_name", "test.local"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "lease_time", "43200"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "offer_delay", "500"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "ping_check", "true"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "ping_check_timeout", "1000"),
				),
			},
		},
	})
}

func TestAccDHCPScopeResource_advancedFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPScopeDestroy("test-advanced"),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name                            = "test-advanced"
  start_address                   = "10.110.0.10"
  end_address                     = "10.110.0.200"
  subnet_mask                     = "255.255.255.0"
  router_address                  = "10.110.0.1"
  dns_servers                     = ["10.110.0.1"]
  domain_name                     = "adv.local"
  exclusions                      = "10.110.0.50|10.110.0.60|10.110.0.100|10.110.0.110"
  domain_search_list              = ["adv.local", "search.local"]
  dns_updates                     = true
  dns_ttl                         = 300
  ping_check                      = true
  ping_check_retries              = 3
  ignore_client_identifier_option = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "name", "test-advanced"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "exclusions", "10.110.0.50|10.110.0.60|10.110.0.100|10.110.0.110"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "domain_search_list.#", "2"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "domain_search_list.0", "adv.local"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "domain_search_list.1", "search.local"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "dns_updates", "true"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "dns_ttl", "300"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "ping_check_retries", "3"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "ignore_client_identifier_option", "true"),
				),
			},
		},
	})
}

func TestAccDHCPScopeResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPScopeDestroy("test-update"),
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPScopeConfig_update_initial(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "name", "test-update"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "end_address", "10.102.0.50"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "lease_time", "86400"),
				),
			},
			{
				Config: testAccDHCPScopeConfig_update_modified(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "name", "test-update"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "end_address", "10.102.0.100"),
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "lease_time", "43200"),
				),
			},
		},
	})
}

func TestAccDHCPScopeResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPScopeDestroy("test-import"),
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPScopeConfig_import(),
			},
			{
				ResourceName:                         "technitium_dhcp_scope.test",
				ImportState:                          true,
				ImportStateId:                        "test-import",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}

func TestAccDHCPScopesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPScopeDestroy("test-ds"),
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPScopesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "name", "test-ds"),
					resource.TestCheckResourceAttrWith("data.technitium_dhcp_scopes.all", "scopes.#", func(value string) error {
						if value == "0" {
							return fmt.Errorf("expected at least 1 scope, got 0")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccDHCPLeasesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPScopeDestroy("test-leases-ds"),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name           = "test-leases-ds"
  start_address  = "10.105.0.10"
  end_address    = "10.105.0.50"
  subnet_mask    = "255.255.255.0"
  router_address = "10.105.0.1"
}

resource "technitium_dhcp_reserved_lease" "test" {
  scope_name       = technitium_dhcp_scope.test.name
  hardware_address = "00:11:22:33:44:55"
  address          = "10.105.0.20"
}

data "technitium_dhcp_leases" "test" {
  scope_name = technitium_dhcp_scope.test.name
  depends_on = [technitium_dhcp_reserved_lease.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.technitium_dhcp_leases.test", "leases.#"),
				),
			},
		},
	})
}

func testAccCheckDHCPScopeDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}

		_, err = c.GetDHCPScope(context.Background(), name)
		if err == nil {
			return fmt.Errorf("DHCP scope %q still exists after destroy", name)
		}

		return nil
	}
}

func testAccDHCPScopeConfig_basic() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name           = "test-basic"
  start_address  = "10.100.0.10"
  end_address    = "10.100.0.50"
  subnet_mask    = "255.255.255.0"
  router_address = "10.100.0.1"
}
`
}

func testAccDHCPScopeConfig_full() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name               = "test-full"
  start_address      = "10.101.0.10"
  end_address        = "10.101.0.100"
  subnet_mask        = "255.255.255.0"
  router_address     = "10.101.0.1"
  dns_servers        = ["10.101.0.1", "1.1.1.1"]
  domain_name        = "test.local"
  lease_time         = 43200
  offer_delay        = 500
  ping_check         = true
  ping_check_timeout = 1000
}
`
}

func testAccDHCPScopeConfig_update_initial() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name           = "test-update"
  start_address  = "10.102.0.10"
  end_address    = "10.102.0.50"
  subnet_mask    = "255.255.255.0"
  router_address = "10.102.0.1"
}
`
}

func testAccDHCPScopeConfig_update_modified() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name           = "test-update"
  start_address  = "10.102.0.10"
  end_address    = "10.102.0.100"
  subnet_mask    = "255.255.255.0"
  router_address = "10.102.0.1"
  lease_time     = 43200
}
`
}

func testAccDHCPScopeConfig_import() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name           = "test-import"
  start_address  = "10.103.0.10"
  end_address    = "10.103.0.50"
  subnet_mask    = "255.255.255.0"
  router_address = "10.103.0.1"
}
`
}

func testAccDHCPScopesDataSourceConfig() string {
	return `
provider "technitium" {}

resource "technitium_dhcp_scope" "test" {
  name           = "test-ds"
  start_address  = "10.104.0.10"
  end_address    = "10.104.0.50"
  subnet_mask    = "255.255.255.0"
  router_address = "10.104.0.1"
}

data "technitium_dhcp_scopes" "all" {
  depends_on = [technitium_dhcp_scope.test]
}
`
}
