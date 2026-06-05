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

resource "technitium_dns_record" "sip" {
  zone     = technitium_dns_zone.example.name
  domain   = "_sip._tcp.example.com"
  type     = "SRV"
  value    = "sip.example.com"
  priority = 10
  weight   = 60
  port     = 5060
}

resource "technitium_dns_record" "forwarder" {
  zone     = technitium_dns_zone.example.name
  domain   = "example.com"
  type     = "FWD"
  value    = "1.1.1.1"
  protocol = "Udp"
}
