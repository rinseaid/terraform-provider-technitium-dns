resource "technitium_dns_record" "web" {
  zone   = technitium_dns_zone.example.name
  domain = "www.example.com"
  type   = "A"
  value  = "203.0.113.50"
  ttl    = 300
}

resource "technitium_dns_record" "mail" {
  zone     = technitium_dns_zone.example.name
  domain   = "example.com"
  type     = "MX"
  value    = "mail.example.com"
  priority = 10
}
