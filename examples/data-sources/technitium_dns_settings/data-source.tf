data "technitium_dns_settings" "current" {}

output "recursion_policy" {
  value = data.technitium_dns_settings.current.recursion
}

output "blocking_enabled" {
  value = data.technitium_dns_settings.current.enable_blocking
}
