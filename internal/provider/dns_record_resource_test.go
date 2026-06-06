package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckDNSRecordDestroy(s *terraform.State) error {
	c, err := testAccClientFromEnv()
	if err != nil {
		return fmt.Errorf("creating client for destroy check: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "technitium_dns_record" {
			continue
		}

		domain := rs.Primary.Attributes["domain"]
		zone := rs.Primary.Attributes["zone"]
		recType := rs.Primary.Attributes["type"]
		value := rs.Primary.Attributes["value"]

		response, err := c.GetRecords(context.Background(), domain, zone, false)
		if err != nil {
			continue
		}

		records, ok := response["records"].([]interface{})
		if !ok {
			continue
		}

		for _, item := range records {
			rec, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if stringFromMap(rec, "type") == recType && recordValueFromRData(rec, recType) == value {
				return fmt.Errorf("DNS record %s (type %s, value %s) still exists in zone %s after destroy",
					domain, recType, value, zone)
			}
		}
	}

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
  ttl      = 0
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

func TestAccDNSRecordResource_AAAA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testaaaa.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "host.testaaaa.example"
  type   = "AAAA"
  value  = "2001:db8::1"
  ttl    = 300
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testaaaa.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "host.testaaaa.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "AAAA"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "2001:db8::1"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "disabled", "false"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testaaaa.example:host.testaaaa.example:AAAA:2001:db8::1"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_NS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testns.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "testns.example"
  type   = "NS"
  value  = "ns1.testns.example"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testns.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "testns.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "NS"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "ns1.testns.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "3600"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "disabled", "false"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testns.example:testns.example:NS:ns1.testns.example"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_PTR(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "1.168.192.in-addr.arpa"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "100.1.168.192.in-addr.arpa"
  type   = "PTR"
  value  = "host.testptr.example"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "1.168.192.in-addr.arpa"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "100.1.168.192.in-addr.arpa"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "PTR"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "host.testptr.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "3600"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "disabled", "false"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "1.168.192.in-addr.arpa:100.1.168.192.in-addr.arpa:PTR:host.testptr.example"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_CAA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testcaa.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "testcaa.example"
  type   = "CAA"
  value  = "letsencrypt.org"
  flags  = 0
  tag    = "issue"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testcaa.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "testcaa.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "CAA"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "letsencrypt.org"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "flags", "0"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "tag", "issue"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "3600"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "disabled", "false"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_ANAME(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testaname.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone  = technitium_dns_zone.test.name
  domain = "testaname.example"
  type   = "ANAME"
  value  = "target.testaname.example"
  aname  = "target.testaname.example"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testaname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "testaname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "ANAME"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "target.testaname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "aname", "target.testaname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "3600"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "disabled", "false"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_DNAME(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testdname.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "sub.testdname.example"
  type   = "DNAME"
  value  = "otherdomain.example"
  dname  = "otherdomain.example"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", "testdname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "domain", "sub.testdname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "DNAME"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "otherdomain.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "dname", "otherdomain.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "3600"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "disabled", "false"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_updateCNAME(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testupdatecname.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "alias.testupdatecname.example"
  type   = "CNAME"
  value  = "original.testupdatecname.example"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "original.testupdatecname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testupdatecname.example:alias.testupdatecname.example:CNAME:original.testupdatecname.example"),
				),
			},
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testupdatecname.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "alias.testupdatecname.example"
  type   = "CNAME"
  value  = "updated.testupdatecname.example"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "updated.testupdatecname.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testupdatecname.example:alias.testupdatecname.example:CNAME:updated.testupdatecname.example"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_updateCAA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testupdatecaa.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "testupdatecaa.example"
  type   = "CAA"
  value  = "letsencrypt.org"
  flags  = 0
  tag    = "issue"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "letsencrypt.org"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "flags", "0"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "tag", "issue"),
				),
			},
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testupdatecaa.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "testupdatecaa.example"
  type   = "CAA"
  value  = "sectigo.com"
  flags  = 128
  tag    = "issuewild"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "sectigo.com"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "flags", "128"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "tag", "issuewild"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_updateNS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testupdatens.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "testupdatens.example"
  type   = "NS"
  value  = "ns1.testupdatens.example"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "ns1.testupdatens.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testupdatens.example:testupdatens.example:NS:ns1.testupdatens.example"),
				),
			},
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testupdatens.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone   = technitium_dns_zone.test.name
  domain = "testupdatens.example"
  type   = "NS"
  value  = "ns2.testupdatens.example"
  ttl    = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "value", "ns2.testupdatens.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "id", "testupdatens.example:testupdatens.example:NS:ns2.testupdatens.example"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_NAPTR(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testnaptr.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone              = technitium_dns_zone.test.name
  domain            = "testnaptr.example"
  type              = "NAPTR"
  value             = "_sip._udp.testnaptr.example"
  naptr_order       = 100
  naptr_preference  = 10
  naptr_flags       = "S"
  naptr_services    = "SIP+D2U"
  naptr_replacement = "_sip._udp.testnaptr.example"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "NAPTR"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "naptr_order", "100"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "naptr_preference", "10"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "naptr_flags", "S"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "naptr_services", "SIP+D2U"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "naptr_replacement", "_sip._udp.testnaptr.example"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_SSHFP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testsshfp.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone                  = technitium_dns_zone.test.name
  domain                = "host.testsshfp.example"
  type                  = "SSHFP"
  value                 = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
  sshfp_algorithm       = 4
  sshfp_fingerprint_type = 2
  sshfp_fingerprint     = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "SSHFP"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "sshfp_algorithm", "4"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "sshfp_fingerprint_type", "2"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "sshfp_fingerprint", "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_TLSA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testtlsa.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone                             = technitium_dns_zone.test.name
  domain                           = "_443._tcp.testtlsa.example"
  type                             = "TLSA"
  value                            = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
  tlsa_certificate_usage           = "DANE-EE"
  tlsa_selector                    = "SPKI"
  tlsa_matching_type               = "SHA2-256"
  tlsa_certificate_association_data = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "TLSA"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "tlsa_certificate_usage", "DANE-EE"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "tlsa_selector", "SPKI"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "tlsa_matching_type", "SHA2-256"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "tlsa_certificate_association_data", "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_URI(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testuri.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone         = technitium_dns_zone.test.name
  domain       = "_ftp._tcp.testuri.example"
  type         = "URI"
  value        = "ftp://ftp.testuri.example/public"
  uri_priority = 10
  uri_weight   = 1
  uri          = "ftp://ftp.testuri.example/public"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "URI"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "uri_priority", "10"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "uri_weight", "1"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "uri", "ftp://ftp.testuri.example/public"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_DS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testds.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone           = technitium_dns_zone.test.name
  domain         = "sub.testds.example"
  type           = "DS"
  value          = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
  ds_key_tag     = 12345
  ds_algorithm   = 13
  ds_digest_type = 2
  ds_digest      = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "DS"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ds_key_tag", "12345"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ds_algorithm", "13"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ds_digest_type", "2"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ds_digest", "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_SVCB(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testsvcb.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone            = technitium_dns_zone.test.name
  domain          = "_dns.testsvcb.example"
  type            = "SVCB"
  value           = "dns.testsvcb.example"
  svc_priority    = 1
  svc_target_name = "dns.testsvcb.example"
  svc_params      = "alpn|dot"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "SVCB"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "svc_priority", "1"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "svc_target_name", "dns.testsvcb.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "svc_params", "alpn|dot"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_HTTPS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "testhttps.example"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone            = technitium_dns_zone.test.name
  domain          = "testhttps.example"
  type            = "HTTPS"
  value           = "cdn.testhttps.example"
  svc_priority    = 1
  svc_target_name = "cdn.testhttps.example"
  svc_params      = "alpn|h2,h3"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "HTTPS"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "svc_priority", "1"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "svc_target_name", "cdn.testhttps.example"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "svc_params", "alpn|h2,h3"),
				),
			},
		},
	})
}

func TestAccDNSZoneResource_forwarder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "fwdzone-test.example"
  type = "Forwarder"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "name", "fwdzone-test.example"),
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "type", "Forwarder"),
				),
			},
		},
	})
}

func TestAccDNSZoneResource_catalog(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
resource "technitium_dns_zone" "test" {
  name = "catalog-test.example"
  type = "Catalog"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "name", "catalog-test.example"),
					resource.TestCheckResourceAttr("technitium_dns_zone.test", "type", "Catalog"),
				),
			},
		},
	})
}
