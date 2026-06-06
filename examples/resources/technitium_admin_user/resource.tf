resource "technitium_admin_user" "ops" {
  username     = "ops-admin"
  password     = var.ops_password
  display_name = "Operations Admin"
}
