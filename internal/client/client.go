package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client is an HTTP client for the Technitium DNS Server API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	UserAgent  string

	username string
	password string
	mu       sync.Mutex
}

const defaultTimeout = 30 * time.Second

// NewWithCredentials authenticates with the Technitium DNS server using
// username/password and returns a Client ready to make API calls.
func NewWithCredentials(baseURL, username, password string, timeout time.Duration) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	c := &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}

	if err := c.login(context.Background()); err != nil {
		return nil, err
	}
	return c, nil
}

// NewWithToken creates a Client using an existing API token. It validates the
// token by making a lightweight API call before returning.
func NewWithToken(baseURL, token string, timeout time.Duration) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	c := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
	_, err := c.doRequest(context.Background(), http.MethodGet, "settings/get", nil)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}
	return c, nil
}

// NewClient authenticates with the Technitium DNS server and returns a
// Client ready to make API calls. The baseURL should be the scheme, host,
// and port of the server (e.g. "http://192.168.1.1:5380").
func NewClient(baseURL, username, password string, timeout time.Duration) (*Client, error) {
	return NewWithCredentials(baseURL, username, password, timeout)
}

func (c *Client) login(ctx context.Context) error {
	params := url.Values{}
	params.Set("user", c.username)
	params.Set("pass", c.password)

	body := strings.NewReader(params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/user/login", body)
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading login response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parsing login response: %w", err)
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		errMsg, _ := result["errorMessage"].(string)
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return fmt.Errorf("login failed: %s", errMsg)
	}

	token, ok := result["token"].(string)
	if !ok || token == "" {
		return errors.New("login response missing token")
	}

	c.mu.Lock()
	c.token = token
	c.mu.Unlock()
	return nil
}

var errInvalidToken = errors.New("session token is invalid or expired")

// doRequest executes an API request, handling retries for GET and token refresh.
func (c *Client) doRequest(ctx context.Context, method, endpoint string, params url.Values) (map[string]interface{}, error) {
	result, err := c.doRequestWithRetry(ctx, method, endpoint, params)
	if err != nil && errors.Is(err, errInvalidToken) && c.username != "" && c.password != "" {
		if loginErr := c.login(ctx); loginErr != nil {
			return nil, fmt.Errorf("token expired and re-authentication failed: %w", loginErr)
		}
		return c.doRequestWithRetry(ctx, method, endpoint, params)
	}
	return result, err
}

const maxRetries = 3

func (c *Client) doRequestWithRetry(ctx context.Context, method, endpoint string, params url.Values) (map[string]interface{}, error) {
	fullURL := c.baseURL + "/api/" + strings.TrimLeft(endpoint, "/")
	isIdempotent := method == http.MethodGet

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if !isIdempotent {
				break
			}
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		var req *http.Request
		var err error

		switch method {
		case http.MethodGet:
			req, err = http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
			if err != nil {
				return nil, fmt.Errorf("creating request: %w", err)
			}
			if params != nil {
				req.URL.RawQuery = params.Encode()
			}
		case http.MethodPost:
			var reqBody io.Reader
			if params != nil {
				reqBody = strings.NewReader(params.Encode())
			}
			req, err = http.NewRequestWithContext(ctx, http.MethodPost, fullURL, reqBody)
			if err != nil {
				return nil, fmt.Errorf("creating request: %w", err)
			}
			if params != nil {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		default:
			return nil, fmt.Errorf("unsupported HTTP method: %s", method)
		}

		c.mu.Lock()
		req.Header.Set("Authorization", "Bearer "+c.token)
		c.mu.Unlock()

		if c.UserAgent != "" {
			req.Header.Set("User-Agent", c.UserAgent)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request to %s failed: %w", endpoint, err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response from %s: %w", endpoint, err)
			continue
		}

		if resp.StatusCode == 429 || resp.StatusCode == 500 ||
			resp.StatusCode == 502 || resp.StatusCode == 503 || resp.StatusCode == 504 {
			lastErr = fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, endpoint, truncateBody(respBody, 200))
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, endpoint, truncateBody(respBody, 200))
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
			return nil, errInvalidToken
		default:
			return nil, fmt.Errorf("unexpected API status from %s: %s", endpoint, status)
		}

		response, _ := result["response"].(map[string]interface{})
		return response, nil
	}

	return nil, lastErr
}

// ---------------------------------------------------------------------------
// Zone CRUD
// ---------------------------------------------------------------------------

func (c *Client) ListZones(ctx context.Context) (map[string]interface{}, error) {
	return c.doRequest(ctx, http.MethodGet, "zones/list", nil)
}

func (c *Client) CreateZone(ctx context.Context, zone, zoneType string, extra ...url.Values) (map[string]interface{}, error) {
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
	return c.doRequest(ctx, http.MethodPost, "zones/create", params)
}

func (c *Client) GetZoneOptions(ctx context.Context, zone string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	return c.doRequest(ctx, http.MethodGet, "zones/options/get", params)
}

func (c *Client) SetZoneOptions(ctx context.Context, zone string, params url.Values) (map[string]interface{}, error) {
	merged := cloneValues(params)
	merged.Set("zone", zone)
	return c.doRequest(ctx, http.MethodPost, "zones/options/set", merged)
}

func (c *Client) DeleteZone(ctx context.Context, zone string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	return c.doRequest(ctx, http.MethodPost, "zones/delete", params)
}

func (c *Client) EnableZone(ctx context.Context, zone string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	return c.doRequest(ctx, http.MethodPost, "zones/enable", params)
}

func (c *Client) DisableZone(ctx context.Context, zone string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	return c.doRequest(ctx, http.MethodPost, "zones/disable", params)
}

// ---------------------------------------------------------------------------
// Record CRUD
// ---------------------------------------------------------------------------

func (c *Client) GetRecords(ctx context.Context, domain, zone string, listZone bool) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("domain", domain)
	if zone != "" {
		params.Set("zone", zone)
	}
	if listZone {
		params.Set("listZone", "true")
	}
	return c.doRequest(ctx, http.MethodGet, "zones/records/get", params)
}

func (c *Client) AddRecord(ctx context.Context, params url.Values) (map[string]interface{}, error) {
	if params.Get("domain") == "" {
		return nil, errors.New("AddRecord: domain parameter is required")
	}
	if params.Get("type") == "" {
		return nil, errors.New("AddRecord: type parameter is required")
	}
	return c.doRequest(ctx, http.MethodPost, "zones/records/add", params)
}

func (c *Client) UpdateRecord(ctx context.Context, params url.Values) (map[string]interface{}, error) {
	if params.Get("domain") == "" {
		return nil, errors.New("UpdateRecord: domain parameter is required")
	}
	if params.Get("type") == "" {
		return nil, errors.New("UpdateRecord: type parameter is required")
	}
	return c.doRequest(ctx, http.MethodPost, "zones/records/update", params)
}

func (c *Client) DeleteRecord(ctx context.Context, params url.Values) (map[string]interface{}, error) {
	if params.Get("domain") == "" {
		return nil, errors.New("DeleteRecord: domain parameter is required")
	}
	if params.Get("type") == "" {
		return nil, errors.New("DeleteRecord: type parameter is required")
	}
	return c.doRequest(ctx, http.MethodPost, "zones/records/delete", params)
}

// ---------------------------------------------------------------------------
// DHCP Scope
// ---------------------------------------------------------------------------

func (c *Client) ListDHCPScopes(ctx context.Context) (map[string]interface{}, error) {
	return c.doRequest(ctx, http.MethodGet, "dhcp/scopes/list", nil)
}

func (c *Client) GetDHCPScope(ctx context.Context, name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("name", name)
	return c.doRequest(ctx, http.MethodGet, "dhcp/scopes/get", params)
}

func (c *Client) SetDHCPScope(ctx context.Context, params url.Values) (map[string]interface{}, error) {
	if params.Get("name") == "" {
		return nil, errors.New("SetDHCPScope: name parameter is required")
	}
	return c.doRequest(ctx, http.MethodPost, "dhcp/scopes/set", params)
}

func (c *Client) DeleteDHCPScope(ctx context.Context, name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("name", name)
	return c.doRequest(ctx, http.MethodPost, "dhcp/scopes/delete", params)
}

func (c *Client) EnableDHCPScope(ctx context.Context, name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("name", name)
	return c.doRequest(ctx, http.MethodPost, "dhcp/scopes/enable", params)
}

func (c *Client) DisableDHCPScope(ctx context.Context, name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("name", name)
	return c.doRequest(ctx, http.MethodPost, "dhcp/scopes/disable", params)
}

// ---------------------------------------------------------------------------
// DHCP Lease
// ---------------------------------------------------------------------------

func (c *Client) ListDHCPLeases(ctx context.Context, scope string) (map[string]interface{}, error) {
	params := url.Values{}
	if scope != "" {
		params.Set("name", scope)
	}
	return c.doRequest(ctx, http.MethodGet, "dhcp/leases/list", params)
}

func (c *Client) AddReservedLease(ctx context.Context, scopeName string, leaseParams url.Values) (map[string]interface{}, error) {
	if leaseParams.Get("hardwareAddress") == "" {
		return nil, errors.New("AddReservedLease: hardwareAddress parameter is required")
	}
	if leaseParams.Get("ipAddress") == "" {
		return nil, errors.New("AddReservedLease: ipAddress parameter is required")
	}
	merged := cloneValues(leaseParams)
	merged.Set("name", scopeName)
	return c.doRequest(ctx, http.MethodPost, "dhcp/scopes/addReservedLease", merged)
}

func (c *Client) RemoveReservedLease(ctx context.Context, scopeName string, leaseParams url.Values) (map[string]interface{}, error) {
	if leaseParams.Get("hardwareAddress") == "" {
		return nil, errors.New("RemoveReservedLease: hardwareAddress parameter is required")
	}
	merged := cloneValues(leaseParams)
	merged.Set("name", scopeName)
	return c.doRequest(ctx, http.MethodPost, "dhcp/scopes/removeReservedLease", merged)
}

// ---------------------------------------------------------------------------
// Allowed / Blocked Zones
// ---------------------------------------------------------------------------

func (c *Client) ListAllowedZones(ctx context.Context, domain string) (map[string]interface{}, error) {
	params := url.Values{}
	if domain != "" {
		params.Set("domain", domain)
	}
	return c.doRequest(ctx, http.MethodGet, "allowed/list", params)
}

func (c *Client) AllowZone(ctx context.Context, domain string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("domain", domain)
	return c.doRequest(ctx, http.MethodPost, "allowed/add", params)
}

func (c *Client) DeleteAllowedZone(ctx context.Context, domain string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("domain", domain)
	return c.doRequest(ctx, http.MethodPost, "allowed/delete", params)
}

func (c *Client) ListBlockedZones(ctx context.Context, domain string) (map[string]interface{}, error) {
	params := url.Values{}
	if domain != "" {
		params.Set("domain", domain)
	}
	return c.doRequest(ctx, http.MethodGet, "blocked/list", params)
}

func (c *Client) BlockZone(ctx context.Context, domain string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("domain", domain)
	return c.doRequest(ctx, http.MethodPost, "blocked/add", params)
}

func (c *Client) DeleteBlockedZone(ctx context.Context, domain string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("domain", domain)
	return c.doRequest(ctx, http.MethodPost, "blocked/delete", params)
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

func (c *Client) GetDNSSettings(ctx context.Context) (map[string]interface{}, error) {
	return c.doRequest(ctx, http.MethodGet, "settings/get", nil)
}

func (c *Client) SetDNSSettings(ctx context.Context, params url.Values) (map[string]interface{}, error) {
	return c.doRequest(ctx, http.MethodPost, "settings/set", params)
}

func (c *Client) GetTSIGKeyNames(ctx context.Context) (map[string]interface{}, error) {
	return c.doRequest(ctx, http.MethodGet, "settings/getTsigKeyNames", nil)
}

func (c *Client) ListCatalogZones(ctx context.Context) (map[string]interface{}, error) {
	return c.doRequest(ctx, http.MethodGet, "zones/catalogs/list", nil)
}

// ---------------------------------------------------------------------------
// DNS Apps
// ---------------------------------------------------------------------------

func (c *Client) ListApps(ctx context.Context) (map[string]interface{}, error) {
	return c.doRequest(ctx, http.MethodGet, "apps/list", nil)
}

func (c *Client) GetAppConfig(ctx context.Context, name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("name", name)
	return c.doRequest(ctx, http.MethodGet, "apps/getConfig", params)
}

func (c *Client) SetAppConfig(ctx context.Context, name string, config string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("name", name)
	params.Set("config", config)
	return c.doRequest(ctx, http.MethodPost, "apps/setConfig", params)
}

// ---------------------------------------------------------------------------
// DNSSEC
// ---------------------------------------------------------------------------

func (c *Client) SignZone(ctx context.Context, zone string, params url.Values) (map[string]interface{}, error) {
	merged := cloneValues(params)
	merged.Set("zone", zone)
	return c.doRequest(ctx, http.MethodPost, "zones/dnssec/sign", merged)
}

func (c *Client) UnsignZone(ctx context.Context, zone string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	return c.doRequest(ctx, http.MethodPost, "zones/dnssec/unsign", params)
}

func (c *Client) GetDNSSECProperties(ctx context.Context, zone string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("zone", zone)
	return c.doRequest(ctx, http.MethodGet, "zones/dnssec/properties/get", params)
}

// ---------------------------------------------------------------------------
// Admin Users
// ---------------------------------------------------------------------------

func (c *Client) ListUsers(ctx context.Context) (map[string]interface{}, error) {
	return c.doRequest(ctx, http.MethodGet, "admin/users/list", nil)
}

func (c *Client) CreateUser(ctx context.Context, username, password string, extra ...url.Values) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("user", username)
	params.Set("pass", password)
	for _, e := range extra {
		for k, vs := range e {
			for _, v := range vs {
				params.Set(k, v)
			}
		}
	}
	return c.doRequest(ctx, http.MethodPost, "admin/users/create", params)
}

func (c *Client) GetUserDetails(ctx context.Context, username string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("user", username)
	params.Set("includeGroups", "true")
	return c.doRequest(ctx, http.MethodGet, "admin/users/get", params)
}

func (c *Client) SetUserDetails(ctx context.Context, username string, params url.Values) (map[string]interface{}, error) {
	merged := cloneValues(params)
	merged.Set("user", username)
	return c.doRequest(ctx, http.MethodPost, "admin/users/set", merged)
}

func (c *Client) DeleteUser(ctx context.Context, username string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("user", username)
	return c.doRequest(ctx, http.MethodPost, "admin/users/delete", params)
}

// ---------------------------------------------------------------------------
// Admin Groups
// ---------------------------------------------------------------------------

func (c *Client) ListGroups(ctx context.Context) (map[string]interface{}, error) {
	return c.doRequest(ctx, http.MethodGet, "admin/groups/list", nil)
}

func (c *Client) CreateGroup(ctx context.Context, name string, description string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("group", name)
	if description != "" {
		params.Set("description", description)
	}
	return c.doRequest(ctx, http.MethodPost, "admin/groups/create", params)
}

func (c *Client) GetGroupDetails(ctx context.Context, name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("group", name)
	params.Set("includeUsers", "true")
	return c.doRequest(ctx, http.MethodGet, "admin/groups/get", params)
}

func (c *Client) SetGroupDetails(ctx context.Context, name string, params url.Values) (map[string]interface{}, error) {
	merged := cloneValues(params)
	merged.Set("group", name)
	return c.doRequest(ctx, http.MethodPost, "admin/groups/set", merged)
}

func (c *Client) DeleteGroup(ctx context.Context, name string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("group", name)
	return c.doRequest(ctx, http.MethodPost, "admin/groups/delete", params)
}

// ---------------------------------------------------------------------------
// Admin Permissions
// ---------------------------------------------------------------------------

func (c *Client) GetPermissions(ctx context.Context, section string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("section", section)
	return c.doRequest(ctx, http.MethodGet, "admin/permissions/get", params)
}

func (c *Client) SetPermissions(ctx context.Context, section string, params url.Values) (map[string]interface{}, error) {
	merged := cloneValues(params)
	merged.Set("section", section)
	return c.doRequest(ctx, http.MethodPost, "admin/permissions/set", merged)
}

func truncateBody(body []byte, maxLen int) string {
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "..."
}

func cloneValues(src url.Values) url.Values {
	dst := make(url.Values, len(src))
	for k, vs := range src {
		dst[k] = append([]string(nil), vs...)
	}
	return dst
}
