Terraform/OpenTofu provider for Technitium DNS Server v14.x. Manages DNS zones, records, DHCP scopes, TSIG keys, DNSSEC signing, server settings, admin users/groups, and allow/block lists through the Technitium HTTP API.

## What's included

- **14 resources** and **8 data sources**, all with import support
- **21 DNS record types:** A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, CAA, SOA, FWD, APP, ANAME, DNAME, NAPTR, SSHFP, TLSA, URI, DS, SVCB, HTTPS
- DHCP scope management with reserved leases, exclusions, static routes, PXE boot options
- TSIG key management through the settings API (pipe-delimited wire format)
- DNSSEC zone signing with ECDSA, RSA, and EdDSA algorithms
- Admin user/group/permission management
- Allow and block zone lists
- Configurable HTTP timeout and automatic retry with exponential backoff

## Testing

241 tests validated against a live Technitium v14 instance. Coverage includes create, read, update, delete, and import for each resource, plus edge cases: disabled records, lowercase hex normalization, multiple records per domain, idempotent deletes, and provider misconfiguration detection.

## Provider configuration

```hcl
provider "technitium" {
  server_url = "http://192.168.1.1:5380"
  api_token  = var.technitium_token
}
```

All attributes support environment variable fallback: `TECHNITIUM_SERVER_URL`, `TECHNITIUM_API_TOKEN`, `TECHNITIUM_USERNAME`, `TECHNITIUM_PASSWORD`.

Built with the Terraform Plugin Framework. Protocol version 6.
