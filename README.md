# terraform-provider-technitium-dns

Terraform/OpenTofu provider for [Technitium DNS Server](https://technitium.com/dns/). Manages zones, records, DHCP scopes, TSIG keys, server settings, and allow/block lists through the Technitium HTTP API.

Built with the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework). Targets Technitium DNS Server v14.x.

## Requirements

- [OpenTofu](https://opentofu.org/) >= 1.6 or [Terraform](https://www.terraform.io/) >= 1.0
- [Go](https://golang.org/) >= 1.25 (to build from source)
- A running Technitium DNS Server instance with API access

## Installation

### From source

```sh
git clone https://forgejo.rinseaid.uk/rinseaid/terraform-provider-technitium-dns.git
cd terraform-provider-technitium-dns
make install
```

This builds the binary and places it in `~/.terraform.d/plugins/` for local use.

### Registry

Planned for the OpenTofu registry under `rinseaid/technitium-dns`. Not yet published.

## Provider Configuration

```hcl
provider "technitium" {
  server_url = "http://192.168.1.1:5380"
  username   = "admin"
  password   = "admin"
}
```

Or use an API token instead of credentials:

```hcl
provider "technitium" {
  server_url = "http://192.168.1.1:5380"
  api_token  = "your-api-token"
}
```

All attributes can be set via environment variables:

| Attribute    | Environment Variable       |
|-------------|---------------------------|
| `server_url` | `TECHNITIUM_SERVER_URL`    |
| `username`   | `TECHNITIUM_USERNAME`      |
| `password`   | `TECHNITIUM_PASSWORD`      |
| `api_token`  | `TECHNITIUM_API_TOKEN`     |

## Resources

### technitium_dns_zone

```hcl
resource "technitium_dns_zone" "example" {
  name = "example.com"
  type = "Primary"
}
```

Zone types: `Primary`, `Secondary`, `Stub`, `Forwarder`, `SecondaryForwarder`, `Catalog`, `SecondaryCatalog`.

Zone options for transfer and access control:

```hcl
resource "technitium_dns_zone" "restricted" {
  name          = "internal.example.com"
  type          = "Primary"
  zone_transfer = "AllowOnlyZoneNameServers"
  notify        = "ZoneNameServers"
  query_access  = "AllowOnlyPrivateNetworks"
  update        = "Deny"
}
```

### technitium_dns_record

```hcl
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
```

SRV records support `priority`, `weight`, and `port` fields:

```hcl
resource "technitium_dns_record" "sip" {
  zone     = technitium_dns_zone.example.name
  domain   = "_sip._tcp.example.com"
  type     = "SRV"
  value    = "sip.example.com"
  priority = 10
  weight   = 60
  port     = 5060
}
```

FWD records configure zone-level forwarders (requires a Forwarder zone):

```hcl
resource "technitium_dns_record" "forwarder" {
  zone     = technitium_dns_zone.example.name
  domain   = "example.com"
  type     = "FWD"
  value    = "1.1.1.1"
  ttl      = 0
  protocol = "Udp"
}
```

Records are identified by `zone:domain:type:value`. Changing `zone`, `domain`, or `type` forces replacement.

### technitium_dns_settings

Singleton resource for server-wide DNS settings.

```hcl
resource "technitium_dns_settings" "server" {
  dns_server_domain     = "dns.example.com"
  dnssec_validation     = true
  recursion             = "AllowOnlyForPrivateNetworks"
  enable_blocking       = true
  blocking_type         = "NxDomain"
  forwarders            = ["1.1.1.1", "8.8.8.8"]
  forwarder_protocol    = "Udp"
  cache_maximum_entries = 10000
  enable_logging        = true
  log_queries           = false
  max_log_file_days     = 30
}
```

### technitium_tsig_key

```hcl
resource "technitium_tsig_key" "transfer" {
  key_name  = "xfer-key.example.com"
  algorithm = "hmac-sha256"
}
```

A shared secret is generated server-side and available as the `shared_secret` attribute.

### technitium_catalog_zone_membership

Associates a zone with a catalog zone.

```hcl
resource "technitium_catalog_zone_membership" "member" {
  zone         = technitium_dns_zone.member.name
  catalog_zone = technitium_dns_zone.catalog.name
}
```

### technitium_dhcp_scope

```hcl
resource "technitium_dhcp_scope" "lan" {
  name           = "LAN"
  start_address  = "192.168.1.100"
  end_address    = "192.168.1.200"
  subnet_mask    = "255.255.255.0"
  router_address = "192.168.1.1"
  dns_servers    = ["192.168.1.1"]
  domain_name    = "home.local"
  lease_time     = 86400
}
```

Advanced DHCP options (exclusions, DNS updates, PXE boot, search domains):

```hcl
resource "technitium_dhcp_scope" "advanced" {
  name                            = "Advanced"
  start_address                   = "10.0.0.100"
  end_address                     = "10.0.0.200"
  subnet_mask                     = "255.255.255.0"
  router_address                  = "10.0.0.1"
  exclusions                      = "10.0.0.150|10.0.0.160"
  domain_search_list              = ["home.local", "corp.local"]
  dns_updates                     = true
  dns_ttl                         = 300
  ignore_client_identifier_option = true
}
```

### technitium_dhcp_reserved_lease

```hcl
resource "technitium_dhcp_reserved_lease" "server" {
  scope_name       = technitium_dhcp_scope.lan.name
  hardware_address = "00:11:22:33:44:55"
  address          = "192.168.1.10"
  hostname         = "fileserver"
  comments         = "NAS reserved IP"
}
```

### technitium_allowed_zone / technitium_blocked_zone

```hcl
resource "technitium_allowed_zone" "trusted" {
  domain = "trusted.example.com"
}

resource "technitium_blocked_zone" "ads" {
  domain = "ads.example.com"
}
```

### technitium_dns_app_config

Manages configuration for a DNS application installed on the Technitium server.

```hcl
resource "technitium_dns_app_config" "blocker" {
  app_name = "Advanced Blocking"
  config   = jsonencode({ "enableBlocking" = true })
}
```

### technitium_admin_user

Manages an admin user account.

```hcl
resource "technitium_admin_user" "ops" {
  username = "ops-admin"
  password = "secure-password"
}
```

### technitium_admin_group

Manages an admin group.

```hcl
resource "technitium_admin_group" "operators" {
  name = "operators"
}
```

### technitium_admin_permission

Manages permissions for an admin group.

```hcl
resource "technitium_admin_permission" "operators_zones" {
  group   = technitium_admin_group.operators.name
  section = "Zones"
  access  = "Allow"
}
```

### technitium_zone_dnssec

Manages DNSSEC signing for a zone.

```hcl
resource "technitium_zone_dnssec" "example" {
  zone      = technitium_dns_zone.example.name
  algorithm = "ECDSA"
}
```

## Data Sources

| Data Source                    | Description                          |
|-------------------------------|--------------------------------------|
| `technitium_dns_zones`        | List all authoritative zones         |
| `technitium_dns_records`      | Query records by zone, domain, type  |
| `technitium_dns_settings`     | Read current server settings         |
| `technitium_dhcp_scopes`      | List all DHCP scopes                 |
| `technitium_dhcp_leases`      | List leases for a scope              |
| `technitium_allowed_zones`    | List allowed zone entries            |
| `technitium_blocked_zones`    | List blocked zone entries            |
| `technitium_dns_apps`         | List installed DNS applications      |

```hcl
data "technitium_dns_zones" "all" {}

output "zone_names" {
  value = data.technitium_dns_zones.all.zones[*].name
}
```

## Import

All resources support `terraform import`. Import IDs vary by resource:

```sh
# Zone
terraform import technitium_dns_zone.example example.com

# Record (zone:domain:type:value)
terraform import technitium_dns_record.web "example.com:www.example.com:A:203.0.113.50"

# DHCP scope
terraform import technitium_dhcp_scope.lan LAN

# DHCP reserved lease (scope_name:hardware_address)
terraform import technitium_dhcp_reserved_lease.server "LAN:00:11:22:33:44:55"

# Allowed/blocked zone
terraform import technitium_allowed_zone.trusted trusted.example.com
terraform import technitium_blocked_zone.ads ads.example.com
```

## Development

```sh
make build        # compile binary
make test         # run unit tests
make testacc      # run acceptance tests (needs TF_ACC=1 and a running Technitium instance)
make testcover    # unit tests with coverage report
make testacccover # acceptance tests with coverage report
make fmt          # format Go and Terraform files
make lint         # golangci-lint
make docs         # generate provider documentation (docs/)
make clean        # remove binary
```

### Acceptance Tests

Acceptance tests run against a live Technitium DNS Server. The CI pipeline starts one as a Docker service container automatically. For local testing:

```sh
docker run -d -p 5380:5380 \
  -e DNS_SERVER_DOMAIN=dns-server \
  -e DNS_SERVER_ADMIN_PASSWORD=admin \
  technitium/dns-server:latest

export TF_ACC=1
export TECHNITIUM_SERVER_URL=http://localhost:5380
export TECHNITIUM_USERNAME=admin
export TECHNITIUM_PASSWORD=admin
make testacc
```

### Project Layout

```
├── internal/
│   ├── client/       # HTTP client for the Technitium API
│   └── provider/     # Resource and data source implementations
├── examples/         # Example .tf files (used by doc generation)
├── docs/             # Generated provider documentation
├── .github/workflows/
│   ├── ci.yml        # Lint, unit tests, acceptance tests, coverage
│   └── release.yml   # GoReleaser on tag push
└── APIDOCS.md        # Technitium DNS Server API reference
```
