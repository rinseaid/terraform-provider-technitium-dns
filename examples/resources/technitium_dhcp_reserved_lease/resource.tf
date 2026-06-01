resource "technitium_dhcp_reserved_lease" "server" {
  scope_name       = technitium_dhcp_scope.lan.name
  hardware_address = "00:11:22:33:44:55"
  address          = "192.168.1.10"
  hostname         = "fileserver"
  comments         = "NAS reserved IP"
}
