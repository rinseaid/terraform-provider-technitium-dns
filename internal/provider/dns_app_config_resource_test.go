package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccDnsAppConfigPreCheck(t *testing.T) {
	t.Helper()
	c, err := testAccClientFromEnv()
	if err != nil {
		t.Skipf("skipping: cannot create client: %s", err)
	}
	resp, err := c.ListApps()
	if err != nil {
		t.Skipf("skipping: cannot list apps: %s", err)
	}
	apps, _ := resp["apps"].([]interface{})
	if len(apps) == 0 {
		t.Skip("skipping: no DNS apps installed on test server")
	}
}

func testAccFirstAppName(t *testing.T) string {
	t.Helper()
	c, err := testAccClientFromEnv()
	if err != nil {
		t.Fatalf("cannot create client: %s", err)
	}
	resp, err := c.ListApps()
	if err != nil {
		t.Fatalf("cannot list apps: %s", err)
	}
	apps, _ := resp["apps"].([]interface{})
	if len(apps) == 0 {
		t.Fatal("no apps installed")
	}
	first, _ := apps[0].(map[string]interface{})
	name, _ := first["name"].(string)
	return name
}

func TestAccDnsAppConfigResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccDnsAppConfigPreCheck(t)
	appName := testAccFirstAppName(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDnsAppConfigHCL(appName, "{}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_dns_app_config.test", "name", appName),
				),
			},
		},
	})
}

func TestAccDnsAppConfigResource_import(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
	testAccDnsAppConfigPreCheck(t)
	appName := testAccFirstAppName(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDnsAppConfigHCL(appName, "{}"),
			},
			{
				ResourceName:      "technitium_dns_app_config.test",
				ImportState:       true,
				ImportStateId:     appName,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDnsAppConfigHCL(name, config string) string {
	return fmt.Sprintf(`
provider "technitium" {}

resource "technitium_dns_app_config" "test" {
  name   = %q
  config = %q
}
`, name, config)
}
