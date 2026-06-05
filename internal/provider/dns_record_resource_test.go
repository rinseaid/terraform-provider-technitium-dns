package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckDNSRecordDestroy(s *terraform.State) error {
	return nil
}

func TestAccDNSRecordResource_A(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testarecord.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "host.testarecord.example"
  type   = "A"
  value  = "192.168.1.100"
  ttl    = 300
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testarecord.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "host.testarecord.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "192.168.1.100"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "disabled", "false"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testarecord.example:host.testarecord.example:A:192.168.1.100"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_CNAME(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testcname.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "alias.testcname.example"
  type   = "CNAME"
  value  = "target.testcname.example"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testcname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "alias.testcname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "CNAME"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "target.testcname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "3600"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testcname.example:alias.testcname.example:CNAME:target.testcname.example"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_MX(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testmx.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "testmx.example"
  type     = "MX"
  value    = "mail.testmx.example"
  ttl      = 3600
  priority = 10
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testmx.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "testmx.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "MX"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "mail.testmx.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "priority", "10"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testmx.example:testmx.example:MX:mail.testmx.example"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_TXT(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testtxt.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "testtxt.example"
  type   = "TXT"
  value  = "v=spf1 include:_spf.google.com ~all"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testtxt.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "testtxt.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "TXT"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "v=spf1 include:_spf.google.com ~all"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "3600"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testtxt.example:testtxt.example:TXT:v=spf1 include:_spf.google.com ~all"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_SRV(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testsrv.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "_sip._tcp.testsrv.example"
  type     = "SRV"
  value    = "sip.testsrv.example"
  ttl      = 3600
  priority = 10
  weight   = 60
  port     = 5060
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testsrv.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "_sip._tcp.testsrv.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "SRV"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "sip.testsrv.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "priority", "10"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "weight", "60"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "port", "5060"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_FWD(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testfwd.example"
  type = "Forwarder"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_dns_zone.test.name
  domain   = "testfwd.example"
  type     = "FWD"
  value    = "1.1.1.1"
  protocol = "Udp"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testfwd.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "testfwd.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "FWD"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "1.1.1.1"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "protocol", "Udp"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_updateValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testupdate.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "host.testupdate.example"
  type   = "A"
  value  = "10.0.0.1"
  ttl    = 300
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "10.0.0.1"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testupdate.example:host.testupdate.example:A:10.0.0.1"),
				),
			},
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testupdate.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "host.testupdate.example"
  type   = "A"
  value  = "10.0.0.2"
  ttl    = 600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "10.0.0.2"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "600"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testupdate.example:host.testupdate.example:A:10.0.0.2"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testimport.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "host.testimport.example"
  type   = "A"
  value  = "172.16.0.1"
  ttl    = 900
}
`,
			},
			{
				ResourceName:      "technitium_dns_record.test",
				ImportState:       true,
				ImportStateId:     "testimport.example:host.testimport.example:A:172.16.0.1",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSRecordsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testdsrecords.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "web.testdsrecords.example"
  type   = "A"
  value  = "203.0.113.50"
  ttl    = 1800
}

data "technitium_dns_records" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "web.testdsrecords.example"
  type   = "A"

  depends_on = [technitium_dns_record.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "zone", "testdsrecords.example"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.#", "1"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.0.domain", "web.testdsrecords.example"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.0.type", "A"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.0.value", "203.0.113.50"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.0.ttl", "1800"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.0.disabled", "false"),
				),
			},
		},
	})
}
