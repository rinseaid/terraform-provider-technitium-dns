resource "technitium_dns_settings" "server" {
  dns_server_domain     = "dns.example.com"
  dnssec_validation     = true
  recursion             = "AllowOnlyForPrivateNetworks"
  enable_blocking       = true
  blocking_type         = "NxDomain"
  forwarders            = ["1.1.1.1", "8.8.8.8"]
  forwarder_protocol    = "Udp"
  cache_maximum_entries = 10000
  enable_logging        = true
  log_queries           = false
  max_log_file_days     = 30
}
