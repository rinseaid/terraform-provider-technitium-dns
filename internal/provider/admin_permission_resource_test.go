package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAdminPermissionResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminPermissionConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_admin_permission.test", "section", "Dashboard"),
					resource.TestCheckResourceAttrSet("technitium_admin_permission.test", "user_permissions"),
				),
			},
		},
	})
}

func TestAccAdminPermissionResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminPermissionConfig(),
			},
			{
				ResourceName:                         "technitium_admin_permission.test",
				ImportState:                          true,
				ImportStateId:                        "Dashboard",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "section",
			},
		},
	})
}

func testAccAdminPermissionConfig() string {
	return `
provider "technitium" {}

resource "technitium_admin_permission" "test" {
  section          = "Dashboard"
  user_permissions = "admin|true|true|true"
}
`
}
