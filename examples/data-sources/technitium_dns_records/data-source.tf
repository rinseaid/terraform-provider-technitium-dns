data "technitium_dns_records" "web" {
  zone   = "example.com"
  domain = "www.example.com"
  type   = "A"
}
