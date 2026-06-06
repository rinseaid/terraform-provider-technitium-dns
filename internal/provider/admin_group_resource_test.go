package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAdminGroupResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAdminGroupDestroy("tftest-group"),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminGroupConfig("tftest-group", "Test group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_admin_group.test", "name", "tftest-group"),
					resource.TestCheckResourceAttr("technitium_admin_group.test", "description", "Test group"),
				),
			},
			{
				Config: testAccAdminGroupConfig("tftest-group", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_admin_group.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccAdminGroupResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAdminGroupDestroy("tftest-import-group"),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminGroupConfig("tftest-import-group", "Import test"),
			},
			{
				ResourceName:                         "technitium_admin_group.test",
				ImportState:                          true,
				ImportStateId:                        "tftest-import-group",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}

func testAccAdminGroupConfig(name, description string) string {
	return fmt.Sprintf(`
provider "technitium" {}

resource "technitium_admin_group" "test" {
  name        = %q
  description = %q
}
`, name, description)
}

func testAccCheckAdminGroupDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}

		_, err = c.GetGroupDetails(context.Background(), name)
		if err == nil {
			return fmt.Errorf("admin group %q still exists after destroy", name)
		}

		return nil
	}
}
