# Technitium DNS Provider

Terraform provider for managing [Technitium DNS Server](https://technitium.com/dns/) resources including DNS zones, DNS records, DHCP scopes, DHCP reserved leases, allowed zones, and blocked zones.

## Provider Configuration

```hcl
provider "technitium" {
  server_url = "http://192.168.1.1:5380"
  username   = "admin"
  password   = "admin"
}
```

Alternatively, authenticate with an API token:

```hcl
provider "technitium" {
  server_url = "http://192.168.1.1:5380"
  api_token  = "your-api-token"
}
```

### Environment Variables

All provider attributes can be set via environment variables. Environment variables are used as fallbacks when the corresponding attribute is not set in the provider block.

| Attribute    | Environment Variable       |
|--------------|----------------------------|
| `server_url` | `TECHNITIUM_SERVER_URL`    |
| `username`   | `TECHNITIUM_USERNAME`      |
| `password`   | `TECHNITIUM_PASSWORD`      |
| `api_token`  | `TECHNITIUM_API_TOKEN`     |

### Argument Reference

- `server_url` (Required) - URL of the Technitium DNS Server API (e.g. `http://localhost:5380`).
- `username` (Optional) - Username for authentication. Required if `api_token` is not set.
- `password` (Optional) - Password for authentication. Required if `api_token` is not set.
- `api_token` (Optional) - API token for authentication. Alternative to username/password.

## Resources

### technitium_dns_zone

Manages an authoritative DNS zone.

```hcl
resource "technitium_dns_zone" "example" {
  name = "example.com"
  type = "Primary"
}
```

### technitium_dns_record

Manages a DNS record within an authoritative zone.

```hcl
resource "technitium_dns_record" "web" {
  zone   = technitium_dns_zone.example.name
  domain = "www.example.com"
  type   = "A"
  value  = "192.168.1.100"
  ttl    = 3600
}

resource "technitium_dns_record" "mail" {
  zone     = technitium_dns_zone.example.name
  domain   = "example.com"
  type     = "MX"
  value    = "mail.example.com"
  priority = 10
  ttl      = 3600
}
```

### technitium_dhcp_scope

Manages a DHCP scope.

```hcl
resource "technitium_dhcp_scope" "lan" {
  name             = "LAN"
  starting_address = "192.168.1.100"
  ending_address   = "192.168.1.200"
  subnet_mask      = "255.255.255.0"
}
```

### technitium_dhcp_reserved_lease

Manages a reserved DHCP lease within a scope.

```hcl
resource "technitium_dhcp_reserved_lease" "printer" {
  scope_name       = technitium_dhcp_scope.lan.name
  hardware_address = "00:11:22:33:44:55"
  address          = "192.168.1.50"
  hostname         = "printer"
  comments         = "Office printer"
}
```

Import by composite ID (`scope_name:hardware_address`):

```shell
terraform import technitium_dhcp_reserved_lease.printer "LAN:00:11:22:33:44:55"
```

### technitium_allowed_zone

Manages an entry in the Allowed Zones list. Allowed zones bypass blocklists, ensuring the domain is always resolvable.

This resource supports create and delete only. Changing the domain triggers replacement.

```hcl
resource "technitium_allowed_zone" "google" {
  domain = "google.com"
}
```

Import by domain name:

```shell
terraform import technitium_allowed_zone.google google.com
```

### technitium_blocked_zone

Manages an entry in the Blocked Zones list. Blocked zones are explicitly denied resolution regardless of upstream or cache state.

This resource supports create and delete only. Changing the domain triggers replacement.

```hcl
resource "technitium_blocked_zone" "ads" {
  domain = "ads.example.com"
}
```

Import by domain name:

```shell
terraform import technitium_blocked_zone.ads ads.example.com
```

## Data Sources

### technitium_dns_zones

Lists all authoritative DNS zones.

```hcl
data "technitium_dns_zones" "all" {}

output "zone_names" {
  value = data.technitium_dns_zones.all.zones[*].name
}
```

### technitium_dns_records

Reads DNS records from an authoritative zone. Supports filtering by domain and record type.

```hcl
data "technitium_dns_records" "web" {
  zone   = "example.com"
  domain = "www.example.com"
  type   = "A"
}
```

### technitium_dhcp_scopes

Lists all DHCP scopes.

```hcl
data "technitium_dhcp_scopes" "all" {}
```

### technitium_allowed_zones

Lists all allowed zones.

```hcl
data "technitium_allowed_zones" "all" {}

output "allowed_domains" {
  value = data.technitium_allowed_zones.all.zones[*].domain
}
```

### technitium_blocked_zones

Lists all blocked zones.

```hcl
data "technitium_blocked_zones" "all" {}

output "blocked_domains" {
  value = data.technitium_blocked_zones.all.zones[*].domain
}
```
