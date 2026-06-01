package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTSIGKeyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckTSIGKeyDestroy("test-key-basic."),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_tsig_key" "test" {
  key_name  = "test-key-basic."
  algorithm = "hmac-sha256"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_tsig_key.test", "key_name", "test-key-basic."),
					resource.TestCheckResourceAttr("technitium_tsig_key.test", "algorithm", "hmac-sha256"),
					resource.TestCheckResourceAttrSet("technitium_tsig_key.test", "shared_secret"),
				),
			},
		},
	})
}

func TestAccTSIGKeyResource_withSecret(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckTSIGKeyDestroy("test-key-secret."),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_tsig_key" "test" {
  key_name      = "test-key-secret."
  algorithm     = "hmac-sha256"
  shared_secret = "dGVzdHNlY3JldGtleTEyMzQ1Njc4OTAxMjM0NTY="
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_tsig_key.test", "key_name", "test-key-secret."),
					resource.TestCheckResourceAttr("technitium_tsig_key.test", "algorithm", "hmac-sha256"),
					resource.TestCheckResourceAttr("technitium_tsig_key.test", "shared_secret", "dGVzdHNlY3JldGtleTEyMzQ1Njc4OTAxMjM0NTY="),
				),
			},
		},
	})
}

func TestAccTSIGKeyResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckTSIGKeyDestroy("test-key-import."),
		Steps: []resource.TestStep{
			{
				Config: `
provider "technitium" {}

resource "technitium_tsig_key" "test" {
  key_name  = "test-key-import."
  algorithm = "hmac-sha256"
}
`,
			},
			{
				ResourceName:                         "technitium_tsig_key.test",
				ImportState:                          true,
				ImportStateId:                        "test-key-import.",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key_name",
			},
		},
	})
}

func testAccCheckTSIGKeyDestroy(keyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}

		response, err := c.GetDNSSettings()
		if err != nil {
			return fmt.Errorf("reading DNS settings for destroy check: %s", err)
		}

		rawKeys, ok := response["tsigKeys"].([]interface{})
		if !ok || rawKeys == nil {
			return nil
		}

		for _, raw := range rawKeys {
			entry, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			if name, ok := entry["keyName"].(string); ok && strings.EqualFold(name, keyName) {
				return fmt.Errorf("TSIG key %q still exists after destroy", keyName)
			}
		}

		return nil
	}
}
