package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var regexpAtLeast1 = regexp.MustCompile(`^[1-9][0-9]*$`)

func TestAccDNSRecordResource_SSHFP_lowercase(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "sshfplc.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone                   = technitium_dns_zone.test.name
  domain                 = "host.sshfplc.example"
  type                   = "SSHFP"
  value                  = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
  sshfp_algorithm        = 4
  sshfp_fingerprint_type = 2
  sshfp_fingerprint      = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "sshfp_fingerprint", "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_NAPTR_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "naptrimport.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone              = technitium_dns_zone.test.name
  domain            = "naptrimport.example"
  type              = "NAPTR"
  value             = "_sip._udp.naptrimport.example"
  naptr_order       = 100
  naptr_preference  = 10
  naptr_flags       = "S"
  naptr_services    = "SIP+D2U"
  naptr_replacement = "_sip._udp.naptrimport.example"
}
`,
			},
			{
				ResourceName:      "technitium_dns_record.test",
				ImportState:       true,
				ImportStateId:     "naptrimport.example:naptrimport.example:NAPTR:_sip._udp.naptrimport.example",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSRecordResource_SSHFP_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "sshfpimport.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone                   = technitium_dns_zone.test.name
  domain                 = "host.sshfpimport.example"
  type                   = "SSHFP"
  value                  = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
  sshfp_algorithm        = 4
  sshfp_fingerprint_type = 2
  sshfp_fingerprint      = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
}
`,
			},
			{
				ResourceName:      "technitium_dns_record.test",
				ImportState:       true,
				ImportStateId:     "sshfpimport.example:host.sshfpimport.example:SSHFP:0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSRecordResource_TLSA_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "tlsaimport.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone                              = technitium_dns_zone.test.name
  domain                            = "_443._tcp.tlsaimport.example"
  type                              = "TLSA"
  value                             = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
  tlsa_certificate_usage            = "DANE-EE"
  tlsa_selector                     = "SPKI"
  tlsa_matching_type                = "SHA2-256"
  tlsa_certificate_association_data = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
}
`,
			},
			{
				ResourceName:      "technitium_dns_record.test",
				ImportState:       true,
				ImportStateId:     "tlsaimport.example:_443._tcp.tlsaimport.example:TLSA:0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSRecordResource_URI_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "uriimport.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone         = technitium_dns_zone.test.name
  domain       = "_ftp._tcp.uriimport.example"
  type         = "URI"
  value        = "ftp://ftp.uriimport.example/pub"
  uri_priority = 10
  uri_weight   = 1
  uri          = "ftp://ftp.uriimport.example/pub"
}
`,
			},
			{
				ResourceName:      "technitium_dns_record.test",
				ImportState:       true,
				ImportStateId:     "uriimport.example:_ftp._tcp.uriimport.example:URI:ftp://ftp.uriimport.example/pub",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSRecordResource_DS_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "dsimport.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone           = technitium_dns_zone.test.name
  domain         = "sub.dsimport.example"
  type           = "DS"
  value          = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
  ds_key_tag     = 12345
  ds_algorithm   = 13
  ds_digest_type = 2
  ds_digest      = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
}
`,
			},
			{
				ResourceName:      "technitium_dns_record.test",
				ImportState:       true,
				ImportStateId:     "dsimport.example:sub.dsimport.example:DS:0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSRecordResource_SVCB_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "svcbimport.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone            = technitium_dns_zone.test.name
  domain          = "_dns.svcbimport.example"
  type            = "SVCB"
  value           = "dns.svcbimport.example"
  svc_priority    = 1
  svc_target_name = "dns.svcbimport.example"
  svc_params      = "alpn|dot"
}
`,
			},
			{
				ResourceName:      "technitium_dns_record.test",
				ImportState:       true,
				ImportStateId:     "svcbimport.example:_dns.svcbimport.example:SVCB:dns.svcbimport.example",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSRecordResource_HTTPS_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "httpsimport.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone            = technitium_dns_zone.test.name
  domain          = "httpsimport.example"
  type            = "HTTPS"
  value           = "cdn.httpsimport.example"
  svc_priority    = 1
  svc_target_name = "cdn.httpsimport.example"
  svc_params      = "alpn|h2,h3"
}
`,
			},
			{
				ResourceName:      "technitium_dns_record.test",
				ImportState:       true,
				ImportStateId:     "httpsimport.example:httpsimport.example:HTTPS:cdn.httpsimport.example",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSZonesDataSource_verifyContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "ds-verify.example"
  type = "Primary"
}

data "technitium_dns_zones" "all" {
  depends_on = [technitium_dns_zone.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.technitium_dns_zones.all", "zones.#"),
					resource.TestMatchResourceAttr("data.technitium_dns_zones.all", "zones.#", regexpAtLeast1),
				),
			},
		},
	})
}

func TestAccDNSRecordsDataSource_verifyContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "ds-recverify.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "host.ds-recverify.example"
  type   = "A"
  value  = "10.20.30.40"
  ttl    = 600
}

data "technitium_dns_records" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "host.ds-recverify.example"
  type   = "A"
  depends_on = [technitium_dns_record.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.#", "1"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.0.value", "10.20.30.40"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.0.ttl", "600"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.0.type", "A"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.0.domain", "host.ds-recverify.example"),
				),
			},
		},
	})
}

func TestAccCatalogZoneMembershipResource_verifyBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "catalog" {
  name = "cat-verify.example"
  type = "Catalog"
}

resource "technitium_dns_zone" "member" {
  name = "member-verify.example"
  type = "Primary"
}

resource "technitium_catalog_zone_membership" "test" {
  zone         = technitium_dns_zone.member.name
  catalog_zone = technitium_dns_zone.catalog.name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_catalog_zone_membership.test", "zone", "member-verify.example"),
					resource.TestCheckResourceAttr("technitium_catalog_zone_membership.test", "catalog_zone", "cat-verify.example"),
					resource.TestCheckResourceAttr("technitium_catalog_zone_membership.test", "id", "member-verify.example"),
				),
			},
		},
	})
}
