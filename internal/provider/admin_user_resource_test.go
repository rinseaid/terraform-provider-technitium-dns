package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAdminUserResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAdminUserDestroy("tftest-user"),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminUserConfig("tftest-user", "TestPass123!", "TF Test User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_admin_user.test", "username", "tftest-user"),
					resource.TestCheckResourceAttr("technitium_admin_user.test", "display_name", "TF Test User"),
					resource.TestCheckResourceAttr("technitium_admin_user.test", "disabled", "false"),
				),
			},
			{
				Config: testAccAdminUserConfig("tftest-user", "TestPass123!", "Updated Name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_admin_user.test", "display_name", "Updated Name"),
				),
			},
		},
	})
}

func TestAccAdminUserResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAdminUserDestroy("tftest-import-user"),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminUserConfig("tftest-import-user", "TestPass123!", "Import Test"),
			},
			{
				ResourceName:                         "technitium_admin_user.test",
				ImportState:                          true,
				ImportStateId:                        "tftest-import-user",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "username",
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func testAccAdminUserConfig(username, password, displayName string) string {
	return fmt.Sprintf(`
provider "technitium" {}

resource "technitium_admin_user" "test" {
  username     = %q
  password     = %q
  display_name = %q
}
`, username, password, displayName)
}

func testAccCheckAdminUserDestroy(username string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}

		_, err = c.GetUserDetails(username)
		if err == nil {
			return fmt.Errorf("admin user %q still exists after destroy", username)
		}

		return nil
	}
}
