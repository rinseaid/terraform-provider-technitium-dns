package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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
		zone := r.URL.Query().Get("zone")
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
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"addedRecord": map[string]interface{}{
					"name": r.URL.Query().Get("domain"),
					"type": r.URL.Query().Get("type"),
				},
			},
		})
	})

	mux.HandleFunc("/api/zones/records/update", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"response": map[string]interface{}{
				"updatedRecord": map[string]interface{}{
					"name": r.URL.Query().Get("domain"),
					"type": r.URL.Query().Get("type"),
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

	return httptest.NewServer(mux)
}

func TestNewClient(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	t.Run("successful login", func(t *testing.T) {
		c, err := NewClient(srv.URL, "admin", "admin")
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
		_, err := NewClient(srv.URL, "admin", "wrong")
		if err == nil {
			t.Fatal("expected error for bad credentials, got nil")
		}
	})

	t.Run("unreachable server", func(t *testing.T) {
		_, err := NewClient("http://127.0.0.1:1", "admin", "admin")
		if err == nil {
			t.Fatal("expected error for unreachable server, got nil")
		}
	})
}

func TestListZones(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	resp, err := c.ListZones()
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

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	resp, err := c.CreateZone("new.example.com", "Primary")
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

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	_, err = c.DeleteZone("example.com")
	if err != nil {
		t.Fatalf("DeleteZone failed: %v", err)
	}
}

func TestGetRecords(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	resp, err := c.GetRecords("example.com", "", false)
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

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	params := url.Values{}
	params.Set("domain", "test.example.com")
	params.Set("type", "A")
	params.Set("ipAddress", "10.0.0.1")

	resp, err := c.AddRecord(params)
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

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	t.Run("missing domain", func(t *testing.T) {
		params := url.Values{}
		params.Set("type", "A")
		_, err := c.AddRecord(params)
		if err == nil {
			t.Fatal("expected error for missing domain")
		}
	})

	t.Run("missing type", func(t *testing.T) {
		params := url.Values{}
		params.Set("domain", "example.com")
		_, err := c.AddRecord(params)
		if err == nil {
			t.Fatal("expected error for missing type")
		}
	})
}

func TestUpdateRecord(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	params := url.Values{}
	params.Set("domain", "example.com")
	params.Set("type", "A")
	params.Set("ipAddress", "1.2.3.4")
	params.Set("newIpAddress", "5.6.7.8")

	_, err = c.UpdateRecord(params)
	if err != nil {
		t.Fatalf("UpdateRecord failed: %v", err)
	}
}

func TestDeleteRecord(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	params := url.Values{}
	params.Set("domain", "example.com")
	params.Set("type", "A")
	params.Set("ipAddress", "1.2.3.4")

	_, err = c.DeleteRecord(params)
	if err != nil {
		t.Fatalf("DeleteRecord failed: %v", err)
	}
}

func TestListDHCPScopes(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	resp, err := c.ListDHCPScopes()
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

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	resp, err := c.GetDHCPScope("Default")
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

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	params := url.Values{}
	params.Set("name", "Default")
	params.Set("startingAddress", "192.168.1.1")
	params.Set("endingAddress", "192.168.1.254")
	params.Set("subnetMask", "255.255.255.0")

	_, err = c.SetDHCPScope(params)
	if err != nil {
		t.Fatalf("SetDHCPScope failed: %v", err)
	}
}

func TestSetDHCPScope_MissingName(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	params := url.Values{}
	_, err = c.SetDHCPScope(params)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestDeleteDHCPScope(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	_, err = c.DeleteDHCPScope("Default")
	if err != nil {
		t.Fatalf("DeleteDHCPScope failed: %v", err)
	}
}

func TestAddReservedLease(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	params := url.Values{}
	params.Set("hardwareAddress", "00:11:22:33:44:55")
	params.Set("ipAddress", "192.168.1.100")
	_, err = c.AddReservedLease("Default", params)
	if err != nil {
		t.Fatalf("AddReservedLease failed: %v", err)
	}
}

func TestRemoveReservedLease(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	removeParams := url.Values{}
	removeParams.Set("hardwareAddress", "00:11:22:33:44:55")
	_, err = c.RemoveReservedLease("Default", removeParams)
	if err != nil {
		t.Fatalf("RemoveReservedLease failed: %v", err)
	}
}

func TestInvalidToken(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c, err := NewClient(srv.URL, "admin", "admin")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Corrupt the token to trigger invalid-token response
	c.token = "bad-token"
	_, err = c.ListZones()
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	if err.Error() != "session token is invalid or expired" {
		t.Errorf("unexpected error message: %v", err)
	}
}
