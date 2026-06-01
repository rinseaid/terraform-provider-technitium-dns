data "technitium_dns_zones" "all" {}

output "zone_names" {
  value = data.technitium_dns_zones.all.zones[*].name
}
