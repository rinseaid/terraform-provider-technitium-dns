package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSSettingsResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_dns_settings" "test" {
  log_queries = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "id", "settings"),
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "log_queries", "false"),
					resource.TestCheckResourceAttrSet("technitium_dns_settings.test", "dns_server_domain"),
					resource.TestCheckResourceAttrSet("technitium_dns_settings.test", "recursion"),
					resource.TestCheckResourceAttrSet("technitium_dns_settings.test", "enable_blocking"),
				),
			},
		},
	})
}

func TestAccDNSSettingsResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_dns_settings" "test" {
  cache_negative_record_ttl = 300
  log_queries               = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "cache_negative_record_ttl", "300"),
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "log_queries", "false"),
				),
			},
			{
				Config: `
provider "technitium" {}

resource "technitium_dns_settings" "test" {
  cache_negative_record_ttl = 600
  log_queries               = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "cache_negative_record_ttl", "600"),
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "log_queries", "true"),
				),
			},
		},
	})
}

func TestAccDNSSettingsResource_blocking(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_dns_settings" "test" {
  enable_blocking = true
  blocking_type   = "NxDomain"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "enable_blocking", "true"),
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "blocking_type", "NxDomain"),
				),
			},
		},
	})
}

func TestAccDNSSettingsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

data "technitium_dns_settings" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.technitium_dns_settings.test", "dns_server_domain"),
					resource.TestCheckResourceAttrSet("data.technitium_dns_settings.test", "recursion"),
					resource.TestCheckResourceAttrSet("data.technitium_dns_settings.test", "enable_blocking"),
				),
			},
		},
	})
}

func TestAccDNSSettingsResource_forwarders(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_dns_settings" "test" {
  forwarders         = ["1.1.1.1", "8.8.8.8"]
  forwarder_protocol = "Udp"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "forwarders.#", "2"),
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "forwarders.0", "1.1.1.1"),
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "forwarders.1", "8.8.8.8"),
					resource.TestCheckResourceAttr("technitium_dns_settings.test", "forwarder_protocol", "Udp"),
				),
			},
		},
	})
}
