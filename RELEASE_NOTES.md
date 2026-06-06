Bug fix and documentation release.

### State-loss fixes

- Read operations no longer drop state on transient API errors for zones, DNSSEC, admin users, groups, and permissions. Previously a temporary 500 or timeout would cause Terraform to think the resource was deleted.
- All resource deletes now check for "not found" before removing from state, preventing crashes on already-deleted resources.
- SVCB/HTTPS `svcParams` map keys are now sorted deterministically, eliminating perpetual plan diffs.
- DNSSEC write-only fields (key parameters) are preserved across reads instead of triggering perpetual diffs.
- Added missing RSAMD5 case to DNSSEC algorithm mapping.

### Client improvements

- Context propagation: all API calls now accept `context.Context` for proper Terraform cancellation support.
- POST requests are no longer retried (retrying non-idempotent operations could create duplicate resources).
- Automatic token refresh on `invalid-token` response when using username/password auth.
- HTTP 500 added to retryable status codes for GET requests.
- Login uses `NewRequestWithContext` instead of `PostForm` (respects context cancellation and custom User-Agent).
- `User-Agent` header set to `terraform-provider-technitium-dns/<version>`.

### Schema fixes

- Zone `type` attribute now uses `RequiresReplace` (in-place type changes are not supported by the API).
- `UseStateForUnknown` added to Optional+Computed record fields and `id`, reducing unnecessary plan noise.
- Admin user `password` uses `WriteOnly` (no longer stored in state).
- Provider credential fields check `IsUnknown` consistently.
- DHCP scope data source attribute names aligned with the resource schema.
- Forwarder zones support configurable address and protocol (previously hardcoded).
- DNS settings data source gains 10 fields that were present in the resource but missing from the data source.

### Documentation

- Fixed incorrect README examples: `admin_permission` used nonexistent attributes, `dns_app_config` used `app_name` instead of `name`.
- Added missing examples for `admin_user`, `admin_group`, `admin_permission`, `dns_app_config`, `zone_dnssec` resources and `dns_apps` data source.
- Regenerated all docs from current schema (provider `timeout` attribute now documented, validator descriptions visible).

### Build and CI

- Pinned golangci-lint to v2.1.6, added `gosec` and `noctx` linters.
- Added `go mod tidy` drift check and `govulncheck` to CI.
- Release workflow skips GPG signing gracefully when key is not configured.
- Removed duplicate `.github/workflows` (Forgejo is canonical CI).
- Removed dead code: `ConvertToReservedLease` stub and its test.
