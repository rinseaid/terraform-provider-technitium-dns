resource "technitium_admin_permission" "operators_zones" {
  section           = "Zones"
  user_permissions  = "ops-admin|true|true|false"
  group_permissions = "operators|true|true|false"
}
