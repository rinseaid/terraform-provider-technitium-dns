resource "technitium_dns_zone" "example" {
  name = "example.com"
  type = "Primary"
}

resource "technitium_dns_zone" "secondary" {
  name                           = "secondary.example.com"
  type                           = "Secondary"
  primary_name_server_addresses  = "10.0.0.1,10.0.0.2"
  primary_zone_transfer_protocol = "Tcp"
}

resource "technitium_dns_zone" "secondary_catalog" {
  name                           = "catalog.example.com"
  type                           = "SecondaryCatalog"
  primary_name_server_addresses  = "10.0.0.1"
  primary_zone_transfer_protocol = "Tls"
  primary_zone_transfer_tsig_key = "xfer-key.example.com"
}
