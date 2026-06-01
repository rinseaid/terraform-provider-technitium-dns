data "technitium_dhcp_scopes" "all" {}

output "scope_names" {
  value = data.technitium_dhcp_scopes.all.scopes[*].name
}
