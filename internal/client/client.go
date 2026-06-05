package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is an HTTP client for the Technitium DNS Server API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewWithCredentials authenticates with the Technitium DNS server using
// username/password and returns a Client ready to make API calls.
func NewWithCredentials(baseURL, username, password string) (*Client, error) {
	return NewClient(baseURL, username, password)
}

// NewWithToken creates a Client using an existing API token.
func NewWithToken(baseURL, token string) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	return c, nil
}

// NewClient authenticates with the Technitium DNS server and returns a
// Client ready to make API calls. The baseURL should be the scheme, host,
// and port of the server (e.g. "http://192.168.1.1:5380").
func NewClient(baseURL, username, password string) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")

	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	params := url.Values{}
	params.Set("user", username)
	params.Set("pass", password)

	resp, err := c.httpClient.PostForm(baseURL+"/api/user/login", params)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading login response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing login response: %w", err)
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		errMsg, _ := result["errorMessage"].(string)
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("login failed: %s", errMsg)
	}

	token, ok := result["token"].(string)
	if !ok || token == "" {
		return nil, errors.New("login response missing token")
	}

	c.token = token
	return c, nil
}

// doRequest executes an HTTP request against the Technitium API and returns
// the parsed "response" object from the JSON body. Auth is provided via the
// Authorization header. Parameters are sent as query string for GET or as
// form data for POST.
func (c *Client) doRequest(method, endpoint string, params url.Values) (map[string]interface{}, error) {
	var req *http.Request
	var err error

	fullURL := c.baseURL + "/api/" + strings.TrimLeft(endpoint, "/")

	switch method {
	case http.MethodGet:
		req, err = http.NewRequest(http.MethodGet, fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		if params != nil {
			req.URL.RawQuery = params.Encode()
		}
	case http.MethodPost:
		var body io.Reader
		if params != nil {
			body = strings.NewReader(params.Encode())
		}
		req, err = http.NewRequest(http.MethodPost, fullURL, body)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		if params != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %w", endpoint, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response from %s: %w", endpoint, err)
	}

	status, _ := result["status"].(string)
	switch status {
	case "ok":
		// success
	case "error":
		errMsg, _ := result["errorMessage"].(string)
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("API error from %s: %s", endpoint, errMsg)
	case "invalid-token":
		return nil, errors.New("session token is invalid or expired")
	default:
		return nil, fmt.Errorf("unexpected API status from %s: %s", endpoint, status)
	}

	response, _ := result["response"].(map[string]interface{})
	return response, nil
}

// ---------------------------------------------------------------------------
// Zone CRUD
// ---------------------------------------------------------------------------

// ListZones returns all authoritative zones hosted on the server.
func (c *Client) ListZones() (map[string]interface{}, error) {
	return c.doRequest(http.MethodGet, "zones/list", nil)
}

// CreateZone creates a new authoritative zone. The zoneType must be one of
// Primary, Secondary, Stub, Forwarder, SecondaryForwarder, Catalog, or
// SecondaryCatalog. Extra params are passed through (e.g. forwarder, protocol
// for Forwarder zones).
func (c *Client) CreateZone(zone, zoneType string, extra ...url.Values) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	params.Set("type", zoneType)
	for _, e := range extra {
		for k, vs := range e {
			for _, v := range vs {
				params.Set(k, v)
			}
		}
	}
	return c.doRequest(http.MethodGet, "zones/create", params)
}

// GetZoneOptions returns zone-specific options.
func (c *Client) GetZoneOptions(zone string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	return c.doRequest(http.MethodGet, "zones/options/get", params)
}

// SetZoneOptions updates zone-specific options.
func (c *Client) SetZoneOptions(zone string, params url.Values) (map[string]interface{}, error) {
	params.Set("zone", zone)
	return c.doRequest(http.MethodGet, "zones/options/set", params)
}

// DeleteZone permanently deletes an authoritative zone.
func (c *Client) DeleteZone(zone string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	return c.doRequest(http.MethodGet, "zones/delete", params)
}

// EnableZone enables an authoritative zone.
func (c *Client) EnableZone(zone string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	return c.doRequest(http.MethodGet, "zones/enable", params)
}

// DisableZone disables an authoritative zone.
func (c *Client) DisableZone(zone string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	return c.doRequest(http.MethodGet, "zones/disable", params)
}

// ---------------------------------------------------------------------------
// Record CRUD
// ---------------------------------------------------------------------------

// GetRecords returns all records for a domain within an authoritative zone.
// Set listZone to true to list all records in the zone.
func (c *Client) GetRecords(domain, zone string, listZone bool) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("domain", domain)
	if zone != "" {
		params.Set("zone", zone)
	}
	if listZone {
		params.Set("listZone", "true")
	}
	return c.doRequest(http.MethodGet, "zones/records/get", params)
}

// AddRecord adds a resource record to an authoritative zone. The params
// must contain at minimum "domain" and "type". Additional fields depend on
// the record type (e.g. "ipAddress" for A/AAAA, "cname" for CNAME).
func (c *Client) AddRecord(params url.Values) (map[string]interface{}, error) {
	if params.Get("domain") == "" {
		return nil, errors.New("AddRecord: domain parameter is required")
	}
	if params.Get("type") == "" {
		return nil, errors.New("AddRecord: type parameter is required")
	}
	return c.doRequest(http.MethodGet, "zones/records/add", params)
}

// UpdateRecord updates an existing record in an authoritative zone. The
// params must contain "domain" and "type" at minimum. Additional fields
// identify the existing record and specify new values.
func (c *Client) UpdateRecord(params url.Values) (map[string]interface{}, error) {
	if params.Get("domain") == "" {
		return nil, errors.New("UpdateRecord: domain parameter is required")
	}
	if params.Get("type") == "" {
		return nil, errors.New("UpdateRecord: type parameter is required")
	}
	return c.doRequest(http.MethodGet, "zones/records/update", params)
}

// DeleteRecord deletes a record from an authoritative zone. The params must
// contain "domain" and "type" at minimum. Additional fields identify which
// specific record to delete within the record set.
func (c *Client) DeleteRecord(params url.Values) (map[string]interface{}, error) {
	if params.Get("domain") == "" {
		return nil, errors.New("DeleteRecord: domain parameter is required")
	}
	if params.Get("type") == "" {
		return nil, errors.New("DeleteRecord: type parameter is required")
	}
	return c.doRequest(http.MethodGet, "zones/records/delete", params)
}

// ---------------------------------------------------------------------------
// DHCP Scope
// ---------------------------------------------------------------------------

// ListDHCPScopes returns all DHCP scopes on the server.
func (c *Client) ListDHCPScopes() (map[string]interface{}, error) {
	return c.doRequest(http.MethodGet, "dhcp/scopes/list", nil)
}

// GetDHCPScope returns the full configuration of a DHCP scope.
func (c *Client) GetDHCPScope(name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("name", name)
	return c.doRequest(http.MethodGet, "dhcp/scopes/get", params)
}

// SetDHCPScope creates or updates a DHCP scope. The params must contain
// "name" at minimum. When creating a new scope, "startingAddress",
// "endingAddress", and "subnetMask" are required.
func (c *Client) SetDHCPScope(params url.Values) (map[string]interface{}, error) {
	if params.Get("name") == "" {
		return nil, errors.New("SetDHCPScope: name parameter is required")
	}
	return c.doRequest(http.MethodPost, "dhcp/scopes/set", params)
}

// DeleteDHCPScope permanently deletes a DHCP scope.
func (c *Client) DeleteDHCPScope(name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("name", name)
	return c.doRequest(http.MethodGet, "dhcp/scopes/delete", params)
}

// EnableDHCPScope enables a DHCP scope to allow lease allocation.
func (c *Client) EnableDHCPScope(name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("name", name)
	return c.doRequest(http.MethodGet, "dhcp/scopes/enable", params)
}

// DisableDHCPScope disables a DHCP scope.
func (c *Client) DisableDHCPScope(name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("name", name)
	return c.doRequest(http.MethodGet, "dhcp/scopes/disable", params)
}

// ---------------------------------------------------------------------------
// DHCP Lease
// ---------------------------------------------------------------------------

// ListDHCPLeases returns all DHCP leases for a scope. The API endpoint
// returns leases across all scopes; callers should filter by scope name.
func (c *Client) ListDHCPLeases(scope string) (map[string]interface{}, error) {
	params := url.Values{}
	if scope != "" {
		params.Set("name", scope)
	}
	return c.doRequest(http.MethodGet, "dhcp/leases/list", params)
}

// AddReservedLease adds a reserved lease entry to the specified scope.
func (c *Client) AddReservedLease(scopeName string, leaseParams url.Values) (map[string]interface{}, error) {
	leaseParams.Set("name", scopeName)
	if leaseParams.Get("hardwareAddress") == "" {
		return nil, errors.New("AddReservedLease: hardwareAddress parameter is required")
	}
	if leaseParams.Get("ipAddress") == "" {
		return nil, errors.New("AddReservedLease: ipAddress parameter is required")
	}
	return c.doRequest(http.MethodGet, "dhcp/scopes/addReservedLease", leaseParams)
}

// RemoveReservedLease removes a reserved lease entry from the specified scope.
func (c *Client) RemoveReservedLease(scopeName string, leaseParams url.Values) (map[string]interface{}, error) {
	leaseParams.Set("name", scopeName)
	if leaseParams.Get("hardwareAddress") == "" {
		return nil, errors.New("RemoveReservedLease: hardwareAddress parameter is required")
	}
	return c.doRequest(http.MethodGet, "dhcp/scopes/removeReservedLease", leaseParams)
}

// ConvertToReservedLease converts a dynamic lease to a reserved lease.
func (c *Client) ConvertToReservedLease(scopeName, hardwareAddress string) (map[string]interface{}, error) {
	return nil, errors.New("ConvertToReservedLease: not implemented")
}

// ---------------------------------------------------------------------------
// Allowed / Blocked Zones
// ---------------------------------------------------------------------------

// ListAllowedZones returns all allowed zones. When domain is empty the root
// is listed, returning top-level zone entries.
func (c *Client) ListAllowedZones(domain string) (map[string]interface{}, error) {
	params := url.Values{}
	if domain != "" {
		params.Set("domain", domain)
	}
	return c.doRequest(http.MethodGet, "allowed/list", params)
}

// AllowZone adds a domain name to the Allowed Zones.
func (c *Client) AllowZone(domain string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("domain", domain)
	return c.doRequest(http.MethodGet, "allowed/add", params)
}

// DeleteAllowedZone removes a domain from the Allowed Zones.
func (c *Client) DeleteAllowedZone(domain string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("domain", domain)
	return c.doRequest(http.MethodGet, "allowed/delete", params)
}

// ListBlockedZones returns all blocked zones. When domain is empty the root
// is listed, returning top-level zone entries.
func (c *Client) ListBlockedZones(domain string) (map[string]interface{}, error) {
	params := url.Values{}
	if domain != "" {
		params.Set("domain", domain)
	}
	return c.doRequest(http.MethodGet, "blocked/list", params)
}

// BlockZone adds a domain name to the Blocked Zones.
func (c *Client) BlockZone(domain string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("domain", domain)
	return c.doRequest(http.MethodGet, "blocked/add", params)
}

// DeleteBlockedZone removes a domain from the Blocked Zones.
func (c *Client) DeleteBlockedZone(domain string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("domain", domain)
	return c.doRequest(http.MethodGet, "blocked/delete", params)
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

// GetDNSSettings returns the DNS server settings.
func (c *Client) GetDNSSettings() (map[string]interface{}, error) {
	return c.doRequest(http.MethodGet, "settings/get", nil)
}

// SetDNSSettings updates the DNS server settings.
func (c *Client) SetDNSSettings(params url.Values) (map[string]interface{}, error) {
	return c.doRequest(http.MethodGet, "settings/set", params)
}

// GetTSIGKeyNames returns TSIG key names configured on the server.
func (c *Client) GetTSIGKeyNames() (map[string]interface{}, error) {
	return c.doRequest(http.MethodGet, "settings/getTsigKeyNames", nil)
}

// ListCatalogZones returns the list of catalog zone names.
func (c *Client) ListCatalogZones() (map[string]interface{}, error) {
	return c.doRequest(http.MethodGet, "zones/catalogs/list", nil)
}
