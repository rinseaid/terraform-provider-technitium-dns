resource "technitium_dns_zone" "catalog" {
  name = "catalog.example.com"
  type = "Catalog"
}

resource "technitium_dns_zone" "member" {
  name = "member.example.com"
  type = "Primary"
}

resource "technitium_catalog_zone_membership" "member" {
  zone         = technitium_dns_zone.member.name
  catalog_zone = technitium_dns_zone.catalog.name
}
