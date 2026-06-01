# terraform-provider-technitium-dns

Terraform/OpenTofu provider for Technitium DNS Server.

## Project Details

- Provider name: terraform-provider-technitium-dns
- Provider type name: technitium (resource prefix)
- Registry: planned for OpenTofu registry under rinseaid/technitium-dns
- API docs reference: APIDOCS.md in repo root (downloaded from TechnitiumSoftware/DnsServer)
- Built with Terraform Plugin Framework (not SDKv2)
- Target Technitium DNS Server v14.x API
- Deployment: Burrito in forge K8s cluster manages OpenTofu layers
- Repo: forgejo.rinseaid.uk/rinseaid/terraform-provider-technitium-dns (private)

## Resources

- technitium_dns_zone
- technitium_dns_record
- technitium_dns_settings (singleton)
- technitium_tsig_key
- technitium_catalog_zone_membership
- technitium_dhcp_scope
- technitium_dhcp_reserved_lease
- technitium_allowed_zone
- technitium_blocked_zone

## Data Sources

- technitium_dns_zones
- technitium_dns_records
- technitium_dns_settings
- technitium_dhcp_scopes
- technitium_dhcp_leases
- technitium_allowed_zones
- technitium_blocked_zones

## Resource Pattern

Each resource follows: schema definition, Create/Read/Update/Delete methods, import support.

## API Authentication

POST /api/user/login with username+password, returns token for subsequent requests.
All API calls are HTTP GET/POST to /api/<endpoint> with Authorization Bearer header.
Provider supports env var fallback: TECHNITIUM_SERVER_URL, TECHNITIUM_USERNAME, TECHNITIUM_PASSWORD, TECHNITIUM_API_TOKEN.

## TSIG Key Wire Format

TSIG keys are managed via the settings API. The `tsigKeys` parameter uses pipe-delimited format: `name1|secret1|algo1|name2|secret2|algo2`. The API normalizes key names with a trailing dot (FQDN); the provider strips this on read-back to match user input.

## Build

```
make build        # compile binary
make install      # install to local plugin directory
make test         # run unit tests
make testcover    # run unit tests with coverage
make testacc      # run acceptance tests (requires TF_ACC=1)
make testacccover # run acceptance tests with coverage
make fmt          # format Go and Terraform files
make lint         # run golangci-lint
make docs         # generate provider documentation
make clean        # remove binary
```
