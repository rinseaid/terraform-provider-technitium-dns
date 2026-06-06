package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// newTestServer sets up an httptest.Server that mimics the Technitium API.
// The login endpoint always succeeds and returns a fixed token. Other
// endpoints echo back a successful response with the request parameters
// embedded in the response for verification.
func newTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		user := r.FormValue("user")
		pass := r.FormValue("pass")
		if user == "admin" && pass == "admin" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"token":  "test-token-abc123",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "error",
			"errorMessage": "invalid credentials",
		})
	})

	mux.HandleFunc("/api/zones/list", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token-abc123" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "invalid-token",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"zones": []interface{}{
					map[string]interface{}{
						"name":     "example.com",
						"type":     "Primary",
						"disabled": false,
					},
					map[string]interface{}{
						"name":     "test.org",
						"type":     "Primary",
						"disabled": false,
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/zones/create", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		zone := r.FormValue("zone")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"domain": zone,
			},
		})
	})

	mux.HandleFunc("/api/zones/delete", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/zones/records/get", func(w http.ResponseWriter, r *http.Request) {
		domain := r.URL.Query().Get("domain")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"zone": map[string]interface{}{
					"name": domain,
					"type": "Primary",
				},
				"records": []interface{}{
					map[string]interface{}{
						"name": domain,
						"type": "A",
						"ttl":  3600,
						"rData": map[string]interface{}{
							"ipAddress": "1.2.3.4",
						},
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/zones/records/add", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"addedRecord": map[string]interface{}{
					"name": r.FormValue("domain"),
					"type": r.FormValue("type"),
				},
			},
		})
	})

	mux.HandleFunc("/api/zones/records/update", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"updatedRecord": map[string]interface{}{
					"name": r.FormValue("domain"),
					"type": r.FormValue("type"),
				},
			},
		})
	})

	mux.HandleFunc("/api/zones/records/delete", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/dhcp/scopes/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"scopes": []interface{}{
					map[string]interface{}{
						"name":            "Default",
						"enabled":         true,
						"startingAddress": "192.168.1.1",
						"endingAddress":   "192.168.1.254",
						"subnetMask":      "255.255.255.0",
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/dhcp/scopes/get", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"name":            r.URL.Query().Get("name"),
				"startingAddress": "192.168.1.1",
				"endingAddress":   "192.168.1.254",
				"subnetMask":      "255.255.255.0",
			},
		})
	})

	mux.HandleFunc("/api/dhcp/scopes/set", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/dhcp/scopes/delete", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/dhcp/scopes/addReservedLease", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/dhcp/scopes/removeReservedLease", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/zones/enable", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/zones/disable", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/zones/options/get", func(w http.ResponseWriter, r *http.Request) {
		zone := r.URL.Query().Get("zone")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"name":     zone,
				"type":     "Primary",
				"disabled": false,
			},
		})
	})

	mux.HandleFunc("/api/zones/options/set", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/dhcp/scopes/enable", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/dhcp/scopes/disable", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/dhcp/leases/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"leases": []interface{}{
					map[string]interface{}{
						"scope":           "Default",
						"type":            "Reserved",
						"hardwareAddress": "00:11:22:33:44:55",
						"address":         "192.168.1.100",
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/allowed/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"zones": []interface{}{
					map[string]interface{}{
						"name": "example.com",
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/allowed/add", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/allowed/delete", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/blocked/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"zones": []interface{}{
					map[string]interface{}{
						"name": "malware.example",
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/blocked/add", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/blocked/delete", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/settings/get", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"dnsServerDomain": "dns.example.com",
				"recursion":       "Allow",
				"tsigKeys": []interface{}{
					map[string]interface{}{
						"keyName":       "test-key.",
						"sharedSecret":  "c2VjcmV0",
						"algorithmName": "hmac-sha256",
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/settings/set", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/settings/getTsigKeyNames", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"tsigKeyNames": []interface{}{"test-key."},
			},
		})
	})

	mux.HandleFunc("/api/zones/catalogs/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"catalogZoneNames": []interface{}{"catalog.example"},
			},
		})
	})

	mux.HandleFunc("/api/apps/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"apps": []interface{}{
					map[string]interface{}{
						"name":    "Failover",
						"version": "1.0",
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/apps/getConfig", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"config": "{}",
			},
		})
	})

	mux.HandleFunc("/api/apps/setConfig", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	// DNSSEC endpoints
	mux.HandleFunc("/api/zones/dnssec/sign", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/zones/dnssec/unsign", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/zones/dnssec/properties/get", func(w http.ResponseWriter, r *http.Request) {
		zone := r.URL.Query().Get("zone")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"zone":            zone,
				"algorithm":       "ECDSA",
				"curve":           "P256",
				"dnsKeyTtl":       float64(86400),
				"zskRolloverDays": float64(30),
				"nxProof":         "NSEC",
			},
		})
	})

	// Admin users endpoints
	mux.HandleFunc("/api/admin/users/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"username":    "admin",
						"displayName": "Administrator",
						"disabled":    false,
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/admin/users/create", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"username":    r.FormValue("user"),
				"displayName": r.FormValue("displayName"),
			},
		})
	})

	mux.HandleFunc("/api/admin/users/get", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"username":              r.URL.Query().Get("user"),
				"displayName":           "Test User",
				"disabled":              false,
				"sessionTimeoutSeconds": float64(1800),
				"memberOfGroups": []interface{}{
					map[string]interface{}{"name": "Administrators"},
				},
			},
		})
	})

	mux.HandleFunc("/api/admin/users/set", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/admin/users/delete", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	// Admin groups endpoints
	mux.HandleFunc("/api/admin/groups/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"groups": []interface{}{
					map[string]interface{}{
						"name":        "Administrators",
						"description": "Built-in admin group",
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/admin/groups/create", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"name":        r.FormValue("group"),
				"description": r.FormValue("description"),
			},
		})
	})

	mux.HandleFunc("/api/admin/groups/get", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"name":        r.URL.Query().Get("group"),
				"description": "Test group",
				"members": []interface{}{
					map[string]interface{}{"username": "admin"},
				},
			},
		})
	})

	mux.HandleFunc("/api/admin/groups/set", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	mux.HandleFunc("/api/admin/groups/delete", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	// Admin permissions endpoints
	mux.HandleFunc("/api/admin/permissions/get", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"section": r.URL.Query().Get("section"),
				"userPermissions": []interface{}{
					map[string]interface{}{
						"username":  "admin",
						"canView":   true,
						"canModify": true,
						"canDelete": true,
					},
				},
				"groupPermissions": []interface{}{
					map[string]interface{}{
						"name":      "Administrators",
						"canView":   true,
						"canModify": true,
						"canDelete": true,
					},
				},
			},
		})
	})

	mux.HandleFunc("/api/admin/permissions/set", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})

	return httptest.NewServer(mux)
}

func TestNewClient(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	t.Run("successful login", func(t *testing.T) {
		c, err := NewClient(srv.URL, "admin", "admin", 0)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if c.token != "test-token-abc123" {
			t.Errorf("expected token 'test-token-abc123', got '%s'", c.token)
		}
		if c.baseURL != srv.URL {
			t.Errorf("expected baseURL '%s', got '%s'", srv.URL, c.baseURL)
		}
	})

	t.Run("bad credentials", func(t *testing.T) {
		_, err := NewClient(srv.URL, "admin", "wrong", 0)
		if err == nil {
			t.Fatal("expected error for bad credentials, got nil")
		}
	})

	t.Run("unreachable server", func(t *testing.T) {
		_, err := NewClient("http://127.0.0.1:1", "admin", "admin", 0)
		if err == nil {
			t.Fatal("expected error for unreachable server, got nil")
		}
	})
}

func TestListZones(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.ListZones(ctx)
	if err != nil {
		t.Fatalf("ListZones failed: %v", err)
	}

	zones, ok := resp["zones"].([]interface{})
	if !ok {
		t.Fatal("response missing 'zones' array")
	}
	if len(zones) != 2 {
		t.Errorf("expected 2 zones, got %d", len(zones))
	}

	first, _ := zones[0].(map[string]interface{})
	if name, _ := first["name"].(string); name != "example.com" {
		t.Errorf("expected first zone 'example.com', got '%s'", name)
	}
}

func TestCreateZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.CreateZone(ctx, "new.example.com", "Primary")
	if err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	domain, _ := resp["domain"].(string)
	if domain != "new.example.com" {
		t.Errorf("expected domain 'new.example.com', got '%s'", domain)
	}
}

func TestDeleteZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.DeleteZone(ctx, "example.com")
	if err != nil {
		t.Fatalf("DeleteZone failed: %v", err)
	}
}

func TestGetRecords(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.GetRecords(ctx, "example.com", "", false)
	if err != nil {
		t.Fatalf("GetRecords failed: %v", err)
	}

	records, ok := resp["records"].([]interface{})
	if !ok {
		t.Fatal("response missing 'records' array")
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}
}

func TestAddRecord(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("domain", "test.example.com")
	params.Set("type", "A")
	params.Set("ipAddress", "10.0.0.1")

	resp, err := c.AddRecord(ctx, params)
	if err != nil {
		t.Fatalf("AddRecord failed: %v", err)
	}

	added, _ := resp["addedRecord"].(map[string]interface{})
	if name, _ := added["name"].(string); name != "test.example.com" {
		t.Errorf("expected record name 'test.example.com', got '%s'", name)
	}
}

func TestAddRecord_MissingParams(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()

	t.Run("missing domain", func(t *testing.T) {
		params := url.Values{}
		params.Set("type", "A")
		_, err := c.AddRecord(ctx, params)
		if err == nil {
			t.Fatal("expected error for missing domain")
		}
	})

	t.Run("missing type", func(t *testing.T) {
		params := url.Values{}
		params.Set("domain", "example.com")
		_, err := c.AddRecord(ctx, params)
		if err == nil {
			t.Fatal("expected error for missing type")
		}
	})
}

func TestUpdateRecord(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("domain", "example.com")
	params.Set("type", "A")
	params.Set("ipAddress", "1.2.3.4")
	params.Set("newIpAddress", "5.6.7.8")

	_, err = c.UpdateRecord(ctx, params)
	if err != nil {
		t.Fatalf("UpdateRecord failed: %v", err)
	}
}

func TestDeleteRecord(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("domain", "example.com")
	params.Set("type", "A")
	params.Set("ipAddress", "1.2.3.4")

	_, err = c.DeleteRecord(ctx, params)
	if err != nil {
		t.Fatalf("DeleteRecord failed: %v", err)
	}
}

func TestListDHCPScopes(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.ListDHCPScopes(ctx)
	if err != nil {
		t.Fatalf("ListDHCPScopes failed: %v", err)
	}

	scopes, ok := resp["scopes"].([]interface{})
	if !ok {
		t.Fatal("response missing 'scopes' array")
	}
	if len(scopes) != 1 {
		t.Errorf("expected 1 scope, got %d", len(scopes))
	}
}

func TestGetDHCPScope(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.GetDHCPScope(ctx, "Default")
	if err != nil {
		t.Fatalf("GetDHCPScope failed: %v", err)
	}

	name, _ := resp["name"].(string)
	if name != "Default" {
		t.Errorf("expected scope name 'Default', got '%s'", name)
	}
}

func TestSetDHCPScope(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("name", "Default")
	params.Set("startingAddress", "192.168.1.1")
	params.Set("endingAddress", "192.168.1.254")
	params.Set("subnetMask", "255.255.255.0")

	_, err = c.SetDHCPScope(ctx, params)
	if err != nil {
		t.Fatalf("SetDHCPScope failed: %v", err)
	}
}

func TestSetDHCPScope_MissingName(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	_, err = c.SetDHCPScope(ctx, params)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestDeleteDHCPScope(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.DeleteDHCPScope(ctx, "Default")
	if err != nil {
		t.Fatalf("DeleteDHCPScope failed: %v", err)
	}
}

func TestAddReservedLease(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("hardwareAddress", "00:11:22:33:44:55")
	params.Set("ipAddress", "192.168.1.100")
	_, err = c.AddReservedLease(ctx, "Default", params)
	if err != nil {
		t.Fatalf("AddReservedLease failed: %v", err)
	}
}

func TestRemoveReservedLease(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	removeParams := url.Values{}
	removeParams.Set("hardwareAddress", "00:11:22:33:44:55")
	_, err = c.RemoveReservedLease(ctx, "Default", removeParams)
	if err != nil {
		t.Fatalf("RemoveReservedLease failed: %v", err)
	}
}

func TestNewWithToken(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewWithToken(srv.URL, "test-token-abc123", 0)
	if err != nil {
		t.Fatalf("NewWithToken failed: %v", err)
	}
	if c.token != "test-token-abc123" {
		t.Errorf("expected token 'test-token-abc123', got '%s'", c.token)
	}

	ctx := context.Background()
	resp, err := c.ListZones(ctx)
	if err != nil {
		t.Fatalf("ListZones with token auth failed: %v", err)
	}
	zones, ok := resp["zones"].([]interface{})
	if !ok || len(zones) != 2 {
		t.Errorf("expected 2 zones, got %v", resp)
	}
}

func TestNewWithToken_TrailingSlash(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewWithToken(srv.URL+"/", "test-token-abc123", 0)
	if err != nil {
		t.Fatalf("NewWithToken failed: %v", err)
	}
	if c.baseURL != srv.URL {
		t.Errorf("expected baseURL '%s', got '%s'", srv.URL, c.baseURL)
	}
}

func TestNewWithCredentials(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewWithCredentials(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("NewWithCredentials failed: %v", err)
	}
	if c.token != "test-token-abc123" {
		t.Errorf("expected token 'test-token-abc123', got '%s'", c.token)
	}
}

func TestEnableZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.EnableZone(ctx, "example.com")
	if err != nil {
		t.Fatalf("EnableZone failed: %v", err)
	}
}

func TestDisableZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.DisableZone(ctx, "example.com")
	if err != nil {
		t.Fatalf("DisableZone failed: %v", err)
	}
}

func TestGetZoneOptions(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.GetZoneOptions(ctx, "example.com")
	if err != nil {
		t.Fatalf("GetZoneOptions failed: %v", err)
	}

	name, _ := resp["name"].(string)
	if name != "example.com" {
		t.Errorf("expected zone name 'example.com', got '%s'", name)
	}
}

func TestSetZoneOptions(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("zoneTransfer", "Deny")
	_, err = c.SetZoneOptions(ctx, "example.com", params)
	if err != nil {
		t.Fatalf("SetZoneOptions failed: %v", err)
	}
}

func TestEnableDHCPScope(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.EnableDHCPScope(ctx, "Default")
	if err != nil {
		t.Fatalf("EnableDHCPScope failed: %v", err)
	}
}

func TestDisableDHCPScope(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.DisableDHCPScope(ctx, "Default")
	if err != nil {
		t.Fatalf("DisableDHCPScope failed: %v", err)
	}
}

func TestListDHCPLeases(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.ListDHCPLeases(ctx, "Default")
	if err != nil {
		t.Fatalf("ListDHCPLeases failed: %v", err)
	}

	leases, ok := resp["leases"].([]interface{})
	if !ok {
		t.Fatal("response missing 'leases' array")
	}
	if len(leases) != 1 {
		t.Errorf("expected 1 lease, got %d", len(leases))
	}
}

func TestListDHCPLeases_NoScope(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.ListDHCPLeases(ctx, "")
	if err != nil {
		t.Fatalf("ListDHCPLeases with empty scope failed: %v", err)
	}
}

func TestListAllowedZones(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.ListAllowedZones(ctx, "")
	if err != nil {
		t.Fatalf("ListAllowedZones failed: %v", err)
	}

	zones, ok := resp["zones"].([]interface{})
	if !ok {
		t.Fatal("response missing 'zones' array")
	}
	if len(zones) != 1 {
		t.Errorf("expected 1 allowed zone, got %d", len(zones))
	}
}

func TestListAllowedZones_WithDomain(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.ListAllowedZones(ctx, "example.com")
	if err != nil {
		t.Fatalf("ListAllowedZones with domain failed: %v", err)
	}
}

func TestAllowZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.AllowZone(ctx, "example.com")
	if err != nil {
		t.Fatalf("AllowZone failed: %v", err)
	}
}

func TestDeleteAllowedZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.DeleteAllowedZone(ctx, "example.com")
	if err != nil {
		t.Fatalf("DeleteAllowedZone failed: %v", err)
	}
}

func TestListBlockedZones(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.ListBlockedZones(ctx, "")
	if err != nil {
		t.Fatalf("ListBlockedZones failed: %v", err)
	}

	zones, ok := resp["zones"].([]interface{})
	if !ok {
		t.Fatal("response missing 'zones' array")
	}
	if len(zones) != 1 {
		t.Errorf("expected 1 blocked zone, got %d", len(zones))
	}
}

func TestBlockZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.BlockZone(ctx, "malware.example")
	if err != nil {
		t.Fatalf("BlockZone failed: %v", err)
	}
}

func TestDeleteBlockedZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.DeleteBlockedZone(ctx, "malware.example")
	if err != nil {
		t.Fatalf("DeleteBlockedZone failed: %v", err)
	}
}

func TestGetDNSSettings(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.GetDNSSettings(ctx)
	if err != nil {
		t.Fatalf("GetDNSSettings failed: %v", err)
	}

	domain, _ := resp["dnsServerDomain"].(string)
	if domain != "dns.example.com" {
		t.Errorf("expected dnsServerDomain 'dns.example.com', got '%s'", domain)
	}
}

func TestSetDNSSettings(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("preferIPv6", "true")
	_, err = c.SetDNSSettings(ctx, params)
	if err != nil {
		t.Fatalf("SetDNSSettings failed: %v", err)
	}
}

func TestGetTSIGKeyNames(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.GetTSIGKeyNames(ctx)
	if err != nil {
		t.Fatalf("GetTSIGKeyNames failed: %v", err)
	}

	names, ok := resp["tsigKeyNames"].([]interface{})
	if !ok {
		t.Fatal("response missing 'tsigKeyNames' array")
	}
	if len(names) != 1 {
		t.Errorf("expected 1 TSIG key name, got %d", len(names))
	}
}

func TestListCatalogZones(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.ListCatalogZones(ctx)
	if err != nil {
		t.Fatalf("ListCatalogZones failed: %v", err)
	}

	names, ok := resp["catalogZoneNames"].([]interface{})
	if !ok {
		t.Fatal("response missing 'catalogZoneNames' array")
	}
	if len(names) != 1 {
		t.Errorf("expected 1 catalog zone, got %d", len(names))
	}
}

func TestAddReservedLease_MissingParams(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()

	t.Run("missing hardwareAddress", func(t *testing.T) {
		params := url.Values{}
		params.Set("ipAddress", "192.168.1.100")
		_, err := c.AddReservedLease(ctx, "Default", params)
		if err == nil {
			t.Fatal("expected error for missing hardwareAddress")
		}
	})

	t.Run("missing ipAddress", func(t *testing.T) {
		params := url.Values{}
		params.Set("hardwareAddress", "00:11:22:33:44:55")
		_, err := c.AddReservedLease(ctx, "Default", params)
		if err == nil {
			t.Fatal("expected error for missing ipAddress")
		}
	})
}

func TestRemoveReservedLease_MissingParams(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	_, err = c.RemoveReservedLease(ctx, "Default", params)
	if err == nil {
		t.Fatal("expected error for missing hardwareAddress")
	}
}

func TestUpdateRecord_MissingParams(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()

	t.Run("missing domain", func(t *testing.T) {
		params := url.Values{}
		params.Set("type", "A")
		_, err := c.UpdateRecord(ctx, params)
		if err == nil {
			t.Fatal("expected error for missing domain")
		}
	})

	t.Run("missing type", func(t *testing.T) {
		params := url.Values{}
		params.Set("domain", "example.com")
		_, err := c.UpdateRecord(ctx, params)
		if err == nil {
			t.Fatal("expected error for missing type")
		}
	})
}

func TestDeleteRecord_MissingParams(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()

	t.Run("missing domain", func(t *testing.T) {
		params := url.Values{}
		params.Set("type", "A")
		_, err := c.DeleteRecord(ctx, params)
		if err == nil {
			t.Fatal("expected error for missing domain")
		}
	})

	t.Run("missing type", func(t *testing.T) {
		params := url.Values{}
		params.Set("domain", "example.com")
		_, err := c.DeleteRecord(ctx, params)
		if err == nil {
			t.Fatal("expected error for missing type")
		}
	})
}

func TestListApps(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.ListApps(ctx)
	if err != nil {
		t.Fatalf("ListApps failed: %v", err)
	}

	apps, ok := resp["apps"].([]interface{})
	if !ok {
		t.Fatal("response missing 'apps' array")
	}
	if len(apps) != 1 {
		t.Errorf("expected 1 app, got %d", len(apps))
	}
}

func TestGetAppConfig(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.GetAppConfig(ctx, "Failover")
	if err != nil {
		t.Fatalf("GetAppConfig failed: %v", err)
	}
}

func TestSetAppConfig(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.SetAppConfig(ctx, "Failover", "{}")
	if err != nil {
		t.Fatalf("SetAppConfig failed: %v", err)
	}
}

func TestInvalidToken(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	// Use NewWithToken so no credentials are stored for auto-refresh
	c, err := NewWithToken(srv.URL, "test-token-abc123", 0)
	if err != nil {
		t.Fatalf("NewWithToken failed: %v", err)
	}

	// Corrupt the token to trigger invalid-token response
	c.token = "bad-token"
	ctx := context.Background()
	_, err = c.ListZones(ctx)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	if err.Error() != "session token is invalid or expired" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDoRequest_MalformedJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"token":  "test-token",
		})
	})
	mux.HandleFunc("/api/zones/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.ListZones(ctx)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestDoRequest_ErrorWithMessage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"token":  "test-token",
		})
	})
	mux.HandleFunc("/api/zones/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "error",
			"errorMessage": "zone not found",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.ListZones(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "API error from zones/list: zone not found" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDoRequest_ErrorWithoutMessage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"token":  "test-token",
		})
	})
	mux.HandleFunc("/api/zones/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.ListZones(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "API error from zones/list: unknown error" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDoRequest_UnexpectedStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"token":  "test-token",
		})
	})
	mux.HandleFunc("/api/zones/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "weird-status",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.ListZones(ctx)
	if err == nil {
		t.Fatal("expected error for unexpected status")
	}
	if err.Error() != "unexpected API status from zones/list: weird-status" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDoRequest_ServerDown(t *testing.T) {
	srv := httptest.NewServer(http.NewServeMux())
	c := &Client{
		baseURL: srv.URL,
		token:   "test-token",
		httpClient: &http.Client{
			Timeout: 1 * time.Second,
		},
	}
	srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.doRequestWithRetry(ctx, http.MethodGet, "zones/list", nil)
	if err == nil {
		t.Fatal("expected error for closed server")
	}
}

func TestDoRequest_POST(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"token":  "test-token",
		})
	})
	mux.HandleFunc("/api/dhcp/scopes/set", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/x-www-form-urlencoded" {
			t.Errorf("expected form content type, got %s", ct)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("name", "Test")
	params.Set("startingAddress", "10.0.0.1")
	params.Set("endingAddress", "10.0.0.254")
	params.Set("subnetMask", "255.255.255.0")

	_, err = c.SetDHCPScope(ctx, params)
	if err != nil {
		t.Fatalf("SetDHCPScope (POST) failed: %v", err)
	}
}

func TestDoRequest_NilResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"token":  "test-token",
		})
	})
	mux.HandleFunc("/api/zones/delete", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.DeleteZone(ctx, "example.com")
	if err != nil {
		t.Fatalf("DeleteZone failed: %v", err)
	}
	if resp != nil {
		t.Errorf("expected nil response, got %v", resp)
	}
}

func TestNewClient_LoginMissingToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	_, err := NewClient(srv.URL, "admin", "admin", 0)
	if err == nil {
		t.Fatal("expected error for missing token in login response")
	}
}

func TestNewClient_LoginMalformedJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	_, err := NewClient(srv.URL, "admin", "admin", 0)
	if err == nil {
		t.Fatal("expected error for malformed login response")
	}
}

func TestGetRecords_WithZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.GetRecords(ctx, "example.com", "example.com", true)
	if err != nil {
		t.Fatalf("GetRecords with zone and listZone failed: %v", err)
	}

	records, ok := resp["records"].([]interface{})
	if !ok {
		t.Fatal("response missing 'records' array")
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}
}

func TestListBlockedZones_WithDomain(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.ListBlockedZones(ctx, "malware.example")
	if err != nil {
		t.Fatalf("ListBlockedZones with domain failed: %v", err)
	}
}

func TestCreateZone_WithExtra(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	extra := url.Values{}
	extra.Set("forwarder", "this-server")
	extra.Set("protocol", "Udp")

	resp, err := c.CreateZone(ctx, "fwd.example.com", "Forwarder", extra)
	if err != nil {
		t.Fatalf("CreateZone with extra failed: %v", err)
	}

	domain, _ := resp["domain"].(string)
	if domain != "fwd.example.com" {
		t.Errorf("expected domain 'fwd.example.com', got '%s'", domain)
	}
}

// ---------------------------------------------------------------------------
// DNSSEC tests
// ---------------------------------------------------------------------------

func TestSignZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("algorithm", "ECDSA")
	params.Set("curve", "P256")
	_, err = c.SignZone(ctx, "example.com", params)
	if err != nil {
		t.Fatalf("SignZone failed: %v", err)
	}
}

func TestUnsignZone(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.UnsignZone(ctx, "example.com")
	if err != nil {
		t.Fatalf("UnsignZone failed: %v", err)
	}
}

func TestGetDNSSECProperties(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.GetDNSSECProperties(ctx, "example.com")
	if err != nil {
		t.Fatalf("GetDNSSECProperties failed: %v", err)
	}

	algo, _ := resp["algorithm"].(string)
	if algo != "ECDSA" {
		t.Errorf("expected algorithm 'ECDSA', got '%s'", algo)
	}

	curve, _ := resp["curve"].(string)
	if curve != "P256" {
		t.Errorf("expected curve 'P256', got '%s'", curve)
	}
}

// ---------------------------------------------------------------------------
// Admin user tests
// ---------------------------------------------------------------------------

func TestListUsers(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	users, ok := resp["users"].([]interface{})
	if !ok {
		t.Fatal("response missing 'users' array")
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

func TestCreateUser(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.CreateUser(ctx, "testuser", "testpass")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	username, _ := resp["username"].(string)
	if username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", username)
	}
}

func TestCreateUser_WithExtra(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	extra := url.Values{}
	extra.Set("displayName", "Test User")

	resp, err := c.CreateUser(ctx, "testuser", "testpass", extra)
	if err != nil {
		t.Fatalf("CreateUser with extra failed: %v", err)
	}

	displayName, _ := resp["displayName"].(string)
	if displayName != "Test User" {
		t.Errorf("expected displayName 'Test User', got '%s'", displayName)
	}
}

func TestGetUserDetails(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.GetUserDetails(ctx, "admin")
	if err != nil {
		t.Fatalf("GetUserDetails failed: %v", err)
	}

	username, _ := resp["username"].(string)
	if username != "admin" {
		t.Errorf("expected username 'admin', got '%s'", username)
	}

	displayName, _ := resp["displayName"].(string)
	if displayName != "Test User" {
		t.Errorf("expected displayName 'Test User', got '%s'", displayName)
	}
}

func TestSetUserDetails(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("disabled", "true")
	_, err = c.SetUserDetails(ctx, "admin", params)
	if err != nil {
		t.Fatalf("SetUserDetails failed: %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.DeleteUser(ctx, "testuser")
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Admin group tests
// ---------------------------------------------------------------------------

func TestListGroups(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.ListGroups(ctx)
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}

	groups, ok := resp["groups"].([]interface{})
	if !ok {
		t.Fatal("response missing 'groups' array")
	}
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
}

func TestCreateGroup(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.CreateGroup(ctx, "Operators", "Operations team")
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	name, _ := resp["name"].(string)
	if name != "Operators" {
		t.Errorf("expected group name 'Operators', got '%s'", name)
	}
}

func TestCreateGroup_NoDescription(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.CreateGroup(ctx, "Operators", "")
	if err != nil {
		t.Fatalf("CreateGroup without description failed: %v", err)
	}
}

func TestGetGroupDetails(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.GetGroupDetails(ctx, "Administrators")
	if err != nil {
		t.Fatalf("GetGroupDetails failed: %v", err)
	}

	name, _ := resp["name"].(string)
	if name != "Administrators" {
		t.Errorf("expected group name 'Administrators', got '%s'", name)
	}

	description, _ := resp["description"].(string)
	if description != "Test group" {
		t.Errorf("expected description 'Test group', got '%s'", description)
	}
}

func TestSetGroupDetails(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("members", "admin,testuser")
	_, err = c.SetGroupDetails(ctx, "Administrators", params)
	if err != nil {
		t.Fatalf("SetGroupDetails failed: %v", err)
	}
}

func TestDeleteGroup(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.DeleteGroup(ctx, "Operators")
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Admin permission tests
// ---------------------------------------------------------------------------

func TestGetPermissions(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	resp, err := c.GetPermissions(ctx, "Zones")
	if err != nil {
		t.Fatalf("GetPermissions failed: %v", err)
	}

	section, _ := resp["section"].(string)
	if section != "Zones" {
		t.Errorf("expected section 'Zones', got '%s'", section)
	}

	userPerms, ok := resp["userPermissions"].([]interface{})
	if !ok {
		t.Fatal("response missing 'userPermissions' array")
	}
	if len(userPerms) != 1 {
		t.Errorf("expected 1 user permission, got %d", len(userPerms))
	}

	groupPerms, ok := resp["groupPermissions"].([]interface{})
	if !ok {
		t.Fatal("response missing 'groupPermissions' array")
	}
	if len(groupPerms) != 1 {
		t.Errorf("expected 1 group permission, got %d", len(groupPerms))
	}
}

func TestSetPermissions(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin", 0)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	ctx := context.Background()
	params := url.Values{}
	params.Set("userPermissions", "admin|true|true|true")
	params.Set("groupPermissions", "Administrators|true|true|true")
	_, err = c.SetPermissions(ctx, "Zones", params)
	if err != nil {
		t.Fatalf("SetPermissions failed: %v", err)
	}
}

func TestDoRequest_RetryOn503(t *testing.T) {
	attempts := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/api/settings/get", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"dnsServerDomain": "test",
			},
		})
	})
	mux.HandleFunc("/api/zones/list", func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("temporarily unavailable"))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"zones": []interface{}{},
			},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewWithToken(srv.URL, "test-token", 0)
	if err != nil {
		t.Fatalf("NewWithToken failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.ListZones(ctx)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts (2 retries + 1 success), got %d", attempts)
	}
}

func TestDoRequest_NonRetriableHTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/settings/get", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"dnsServerDomain": "test",
			},
		})
	})
	mux.HandleFunc("/api/zones/list", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewWithToken(srv.URL, "test-token", 0)
	if err != nil {
		t.Fatalf("NewWithToken failed: %v", err)
	}

	ctx := context.Background()
	_, err = c.ListZones(ctx)
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	expected := "HTTP 403 from zones/list: forbidden"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestMutatingEndpoints_UsePOST(t *testing.T) {
	var lastMethod string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/settings/get", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{"dnsServerDomain": "test"},
		})
	})
	mux.HandleFunc("/api/zones/create", func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{"domain": "test"},
		})
	})
	mux.HandleFunc("/api/zones/delete", func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"response": map[string]interface{}{},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewWithToken(srv.URL, "test-token", 0)
	if err != nil {
		t.Fatalf("NewWithToken failed: %v", err)
	}

	ctx := context.Background()
	_, _ = c.CreateZone(ctx, "test.example", "Primary")
	if lastMethod != http.MethodPost {
		t.Errorf("CreateZone: expected POST, got %s", lastMethod)
	}

	_, _ = c.DeleteZone(ctx, "test.example")
	if lastMethod != http.MethodPost {
		t.Errorf("DeleteZone: expected POST, got %s", lastMethod)
	}
}
