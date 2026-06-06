resource "technitium_zone_dnssec" "example" {
  zone      = technitium_dns_zone.example.name
  algorithm = "ECDSA"
  curve     = "P256"
  nx_proof  = "NSEC"
}
