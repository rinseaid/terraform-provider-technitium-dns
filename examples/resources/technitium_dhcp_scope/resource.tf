resource "technitium_dhcp_scope" "lan" {
  name           = "LAN"
  start_address  = "192.168.1.100"
  end_address    = "192.168.1.200"
  subnet_mask    = "255.255.255.0"
  router_address = "192.168.1.1"
  dns_servers    = ["192.168.1.1"]
  domain_name    = "home.local"
  lease_time     = 86400
  ntp_servers    = ["192.168.1.1"]
  static_routes  = "172.16.0.0|255.255.255.0|192.168.1.2"
}
