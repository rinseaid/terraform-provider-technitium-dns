package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSRecordResource_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "disabled-rec.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "off.disabled-rec.example"
  type     = "A"
  value    = "10.0.0.1"
  disabled = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "disabled", "true"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "10.0.0.1"),
				),
			},
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "disabled-rec.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "off.disabled-rec.example"
  type     = "A"
  value    = "10.0.0.1"
  disabled = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "disabled", "false"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_comments(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "comments-rec.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "noted.comments-rec.example"
  type     = "A"
  value    = "10.0.0.2"
  comments = "managed by terraform"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "comments", "managed by terraform"),
				),
			},
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "comments-rec.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "noted.comments-rec.example"
  type     = "A"
  value    = "10.0.0.2"
  comments = "updated comment"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "comments", "updated comment"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_multipleA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "multi-a.example"
  type = "Primary"
}

resource "technitium_dns_record" "first" {
  zone   = technitium_dns_zone.test.name
  domain = "host.multi-a.example"
  type   = "A"
  value  = "10.0.0.1"
}

resource "technitium_dns_record" "second" {
  zone   = technitium_dns_zone.test.name
  domain = "host.multi-a.example"
  type   = "A"
  value  = "10.0.0.2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.first", "value", "10.0.0.1"),
					resource.TestCheckResourceAttr("technitium_dns_record.second", "value", "10.0.0.2"),
					resource.TestCheckResourceAttr("technitium_dns_record.first", "id", "multi-a.example:host.multi-a.example:A:10.0.0.1"),
					resource.TestCheckResourceAttr("technitium_dns_record.second", "id", "multi-a.example:host.multi-a.example:A:10.0.0.2"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_MX_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "mxupd.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "mxupd.example"
  type     = "MX"
  value    = "mail1.mxupd.example"
  priority = 10
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "mail1.mxupd.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "priority", "10"),
				),
			},
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "mxupd.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "mxupd.example"
  type     = "MX"
  value    = "mail2.mxupd.example"
  priority = 20
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "mail2.mxupd.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "priority", "20"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_SRV_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "srvupd.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "_sip._tcp.srvupd.example"
  type     = "SRV"
  value    = "sip1.srvupd.example"
  priority = 10
  weight   = 60
  port     = 5060
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "sip1.srvupd.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "port", "5060"),
				),
			},
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "srvupd.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "_sip._tcp.srvupd.example"
  type     = "SRV"
  value    = "sip2.srvupd.example"
  priority = 20
  weight   = 40
  port     = 5061
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "sip2.srvupd.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "priority", "20"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "weight", "40"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "port", "5061"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_DNAME_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "dnameupd.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "sub.dnameupd.example"
  type   = "DNAME"
  value  = "target1.example"
  dname  = "target1.example"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "target1.example"),
				),
			},
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "dnameupd.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "sub.dnameupd.example"
  type   = "DNAME"
  value  = "target2.example"
  dname  = "target2.example"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "target2.example"),
				),
			},
		},
	})
}

func TestAccProviderConfig_invalidCredentials(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {
  server_url = "http://localhost:5380"
  username   = "admin"
  password   = "wrong-password"
}

resource "technitium_dns_zone" "test" {
  name = "should-fail.example"
  type = "Primary"
}
`,
				ExpectError: regexp.MustCompile(`login failed|Unable to Create`),
			},
		},
	})
}

func TestAccProviderConfig_invalidToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {
  server_url = "http://localhost:5380"
  api_token  = "invalid-token-value"
}

resource "technitium_dns_zone" "test" {
  name = "should-fail.example"
  type = "Primary"
}
`,
				ExpectError: regexp.MustCompile(`token validation failed|Unable to Create`),
			},
		},
	})
}

func TestAccDHCPScopeResource_leaseTimePrecision(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPScopeDestroy("lease-precision"),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dhcp_scope" "test" {
  name          = "lease-precision"
  start_address = "10.88.0.1"
  end_address   = "10.88.0.254"
  subnet_mask   = "255.255.255.0"
  lease_time    = 90060
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dhcp_scope.test", "lease_time", "90060"),
				),
			},
		},
	})
}

func TestAccDHCPScopeResource_leaseTimeRejectsSeconds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dhcp_scope" "test" {
  name          = "lease-bad"
  start_address = "10.89.0.1"
  end_address   = "10.89.0.254"
  subnet_mask   = "255.255.255.0"
  lease_time    = 90061
}
`,
				ExpectError: regexp.MustCompile(`must be divisible by 60`),
			},
		},
	})
}

func TestAccDHCPScopesDataSource_verifyContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDHCPScopeDestroy("ds-verify-scope"),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dhcp_scope" "test" {
  name          = "ds-verify-scope"
  start_address = "10.77.0.1"
  end_address   = "10.77.0.254"
  subnet_mask   = "255.255.255.0"
}

data "technitium_dhcp_scopes" "all" {
  depends_on = [technitium_dhcp_scope.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("data.technitium_dhcp_scopes.all", "scopes.#", regexpAtLeast1),
				),
			},
		},
	})
}

func TestAccAllowedZonesDataSource_verifyContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_allowed_zone" "test" {
  domain = "ds-allowed-verify.example"
}

data "technitium_allowed_zones" "all" {
  depends_on = [technitium_allowed_zone.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.technitium_allowed_zones.all", "zones.#"),
				),
			},
		},
	})
}

func TestAccBlockedZonesDataSource_verifyContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_blocked_zone" "test" {
  domain = "ds-blocked-verify.example"
}

data "technitium_blocked_zones" "all" {
  depends_on = [technitium_blocked_zone.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.technitium_blocked_zones.all", "zones.#"),
				),
			},
		},
	})
}

func TestAccDNSSettingsDataSource_verifyContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
data "technitium_dns_settings" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.technitium_dns_settings.test", "dns_server_domain", "test.local"),
					resource.TestCheckResourceAttrSet("data.technitium_dns_settings.test", "recursion"),
					resource.TestCheckResourceAttrSet("data.technitium_dns_settings.test", "enable_blocking"),
				),
			},
		},
	})
}
