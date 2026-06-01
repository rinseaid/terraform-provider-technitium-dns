# terraform-provider-technitium

Terraform/OpenTofu provider for Technitium DNS Server.

## Project Details

- Provider name: terraform-provider-technitium
- Registry: planned for OpenTofu registry under rinseaid/technitium
- API docs reference: APIDOCS.md in repo root (downloaded from TechnitiumSoftware/DnsServer)
- Built with Terraform Plugin Framework (not SDKv2)
- Target Technitium DNS Server v14.x API
- Deployment: Burrito in forge K8s cluster manages OpenTofu layers

## Existing Providers (Reference)

- kenske/terraform-provider-technitium (DHCP focus)
- darkhonor/terraform-provider-technitium (DNS focus)

## Resources to Implement

- technitium_dns_zone
- technitium_dns_record
- technitium_dhcp_scope
- technitium_dhcp_reserved_lease
- technitium_allowed_zone
- technitium_blocked_zone
- technitium_dns_settings

## Data Sources to Implement

- technitium_dhcp_lease

## Resource Pattern

Each resource follows: schema definition, Create/Read/Update/Delete methods, import support.

## API Authentication

POST /api/user/login with username+password, returns token for subsequent requests.
All API calls are HTTP GET/POST to /api/<endpoint> with token parameter.

## Build

```
make build      # compile binary
make install    # install to local plugin directory
make test       # run unit tests
make testacc    # run acceptance tests (requires TF_ACC=1)
make fmt        # format Go and Terraform files
make lint       # run golangci-lint
make docs       # generate provider documentation
make clean      # remove binary
```
