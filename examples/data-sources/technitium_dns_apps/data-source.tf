data "technitium_dns_apps" "all" {}

output "installed_apps" {
  value = data.technitium_dns_apps.all.apps[*].name
}
