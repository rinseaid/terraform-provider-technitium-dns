package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0E-0E-0C-88-A4-B9", "0e:0e:0c:88:a4:b9"},
		{"0e:0e:0c:88:a4:b9", "0e:0e:0c:88:a4:b9"},
		{"0E:0E:0C:88:A4:B9", "0e:0e:0c:88:a4:b9"},
		{"0e.0e.0c.88.a4.b9", "0e:0e:0c:88:a4:b9"},
		{"00-11-22-33-44-55", "00:11:22:33:44:55"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeMAC(tt.input)
			if got != tt.want {
				t.Errorf("normalizeMAC(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRecordValueFromRData(t *testing.T) {
	tests := []struct {
		name       string
		recordType string
		rData      map[string]interface{}
		want       string
	}{
		{
			name:       "A record",
			recordType: "A",
			rData:      map[string]interface{}{"ipAddress": "192.168.1.1"},
			want:       "192.168.1.1",
		},
		{
			name:       "AAAA record",
			recordType: "AAAA",
			rData:      map[string]interface{}{"ipAddress": "::1"},
			want:       "::1",
		},
		{
			name:       "CNAME record",
			recordType: "CNAME",
			rData:      map[string]interface{}{"cname": "target.example.com"},
			want:       "target.example.com",
		},
		{
			name:       "NS record",
			recordType: "NS",
			rData:      map[string]interface{}{"nameServer": "ns1.example.com"},
			want:       "ns1.example.com",
		},
		{
			name:       "PTR record",
			recordType: "PTR",
			rData:      map[string]interface{}{"ptrName": "host.example.com"},
			want:       "host.example.com",
		},
		{
			name:       "MX record",
			recordType: "MX",
			rData:      map[string]interface{}{"exchange": "mail.example.com", "preference": float64(10)},
			want:       "mail.example.com",
		},
		{
			name:       "TXT record",
			recordType: "TXT",
			rData:      map[string]interface{}{"text": "v=spf1 -all"},
			want:       "v=spf1 -all",
		},
		{
			name:       "SRV record",
			recordType: "SRV",
			rData:      map[string]interface{}{"target": "sip.example.com", "priority": float64(10), "weight": float64(60), "port": float64(5060)},
			want:       "sip.example.com",
		},
		{
			name:       "CAA record",
			recordType: "CAA",
			rData:      map[string]interface{}{"value": "letsencrypt.org"},
			want:       "letsencrypt.org",
		},
		{
			name:       "SOA record",
			recordType: "SOA",
			rData:      map[string]interface{}{"primaryNameServer": "ns1.example.com"},
			want:       "ns1.example.com",
		},
		{
			name:       "FWD record",
			recordType: "FWD",
			rData:      map[string]interface{}{"forwarder": "1.1.1.1", "protocol": "Udp"},
			want:       "1.1.1.1",
		},
		{
			name:       "unknown type returns empty",
			recordType: "UNKNOWN",
			rData:      map[string]interface{}{"foo": "bar"},
			want:       "",
		},
		{
			name:       "missing rData returns empty",
			recordType: "A",
			rData:      nil,
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := map[string]interface{}{}
			if tt.rData != nil {
				rec["rData"] = tt.rData
			}
			got := recordValueFromRData(rec, tt.recordType)
			if got != tt.want {
				t.Errorf("recordValueFromRData() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCompositeID(t *testing.T) {
	got := compositeID("example.com", "www.example.com", "A", "1.2.3.4")
	want := "example.com:www.example.com:A:1.2.3.4"
	if got != want {
		t.Errorf("compositeID() = %q, want %q", got, want)
	}
}

func TestBuildAddParams_A(t *testing.T) {
	plan := &dnsRecordResourceModel{
		Zone:   types.StringValue("example.com"),
		Domain: types.StringValue("host.example.com"),
		Type:   types.StringValue("A"),
		Value:  types.StringValue("10.0.0.1"),
		TTL:    types.Int64Value(300),
	}
	params := buildAddParams(plan)

	if params.Get("domain") != "host.example.com" {
		t.Errorf("domain = %q", params.Get("domain"))
	}
	if params.Get("zone") != "example.com" {
		t.Errorf("zone = %q", params.Get("zone"))
	}
	if params.Get("type") != "A" {
		t.Errorf("type = %q", params.Get("type"))
	}
	if params.Get("ipAddress") != "10.0.0.1" {
		t.Errorf("ipAddress = %q", params.Get("ipAddress"))
	}
	if params.Get("ttl") != "300" {
		t.Errorf("ttl = %q", params.Get("ttl"))
	}
}

func TestBuildAddParams_MX(t *testing.T) {
	plan := &dnsRecordResourceModel{
		Zone:     types.StringValue("example.com"),
		Domain:   types.StringValue("example.com"),
		Type:     types.StringValue("MX"),
		Value:    types.StringValue("mail.example.com"),
		TTL:      types.Int64Value(3600),
		Priority: types.Int64Value(10),
	}
	params := buildAddParams(plan)

	if params.Get("exchange") != "mail.example.com" {
		t.Errorf("exchange = %q", params.Get("exchange"))
	}
	if params.Get("preference") != "10" {
		t.Errorf("preference = %q", params.Get("preference"))
	}
}

func TestBuildAddParams_SRV(t *testing.T) {
	plan := &dnsRecordResourceModel{
		Zone:     types.StringValue("example.com"),
		Domain:   types.StringValue("_sip._tcp.example.com"),
		Type:     types.StringValue("SRV"),
		Value:    types.StringValue("sip.example.com"),
		TTL:      types.Int64Value(3600),
		Priority: types.Int64Value(10),
		Weight:   types.Int64Value(60),
		Port:     types.Int64Value(5060),
	}
	params := buildAddParams(plan)

	if params.Get("target") != "sip.example.com" {
		t.Errorf("target = %q", params.Get("target"))
	}
	if params.Get("priority") != "10" {
		t.Errorf("priority = %q", params.Get("priority"))
	}
	if params.Get("weight") != "60" {
		t.Errorf("weight = %q", params.Get("weight"))
	}
	if params.Get("port") != "5060" {
		t.Errorf("port = %q", params.Get("port"))
	}
}

func TestBuildAddParams_FWD(t *testing.T) {
	plan := &dnsRecordResourceModel{
		Zone:     types.StringValue("example.com"),
		Domain:   types.StringValue("example.com"),
		Type:     types.StringValue("FWD"),
		Value:    types.StringValue("1.1.1.1"),
		TTL:      types.Int64Value(3600),
		Protocol: types.StringValue("Udp"),
	}
	params := buildAddParams(plan)

	if params.Get("forwarder") != "1.1.1.1" {
		t.Errorf("forwarder = %q", params.Get("forwarder"))
	}
	if params.Get("protocol") != "Udp" {
		t.Errorf("protocol = %q", params.Get("protocol"))
	}
}

func TestBuildDeleteParams_SRV(t *testing.T) {
	state := &dnsRecordResourceModel{
		Zone:     types.StringValue("example.com"),
		Domain:   types.StringValue("_sip._tcp.example.com"),
		Type:     types.StringValue("SRV"),
		Value:    types.StringValue("sip.example.com"),
		Priority: types.Int64Value(10),
		Weight:   types.Int64Value(60),
		Port:     types.Int64Value(5060),
	}
	params := buildDeleteParams(state)

	if params.Get("target") != "sip.example.com" {
		t.Errorf("target = %q", params.Get("target"))
	}
	if params.Get("priority") != "10" {
		t.Errorf("priority = %q", params.Get("priority"))
	}
	if params.Get("weight") != "60" {
		t.Errorf("weight = %q", params.Get("weight"))
	}
	if params.Get("port") != "5060" {
		t.Errorf("port = %q", params.Get("port"))
	}
}

func TestBuildAddParams_CAA(t *testing.T) {
	plan := &dnsRecordResourceModel{
		Zone:  types.StringValue("example.com"),
		Domain: types.StringValue("example.com"),
		Type:  types.StringValue("CAA"),
		Value: types.StringValue("letsencrypt.org"),
		TTL:   types.Int64Value(3600),
		Flags: types.Int64Value(0),
		Tag:   types.StringValue("issue"),
	}
	params := buildAddParams(plan)

	if params.Get("value") != "letsencrypt.org" {
		t.Errorf("value = %q", params.Get("value"))
	}
	if params.Get("flags") != "0" {
		t.Errorf("flags = %q", params.Get("flags"))
	}
	if params.Get("tag") != "issue" {
		t.Errorf("tag = %q", params.Get("tag"))
	}
}

func TestBuildAddParams_FWD_Advanced(t *testing.T) {
	plan := &dnsRecordResourceModel{
		Zone:             types.StringValue("example.com"),
		Domain:           types.StringValue("example.com"),
		Type:             types.StringValue("FWD"),
		Value:            types.StringValue("1.1.1.1"),
		TTL:              types.Int64Value(3600),
		Protocol:         types.StringValue("Tcp"),
		DnssecValidation: types.BoolValue(true),
		ProxyType:        types.StringValue("Http"),
		ProxyAddress:     types.StringValue("proxy.example.com"),
		ProxyPort:        types.Int64Value(8080),
	}
	params := buildAddParams(plan)

	if params.Get("forwarder") != "1.1.1.1" {
		t.Errorf("forwarder = %q", params.Get("forwarder"))
	}
	if params.Get("protocol") != "Tcp" {
		t.Errorf("protocol = %q", params.Get("protocol"))
	}
	if params.Get("dnssecValidation") != "true" {
		t.Errorf("dnssecValidation = %q", params.Get("dnssecValidation"))
	}
	if params.Get("proxyType") != "Http" {
		t.Errorf("proxyType = %q", params.Get("proxyType"))
	}
	if params.Get("proxyAddress") != "proxy.example.com" {
		t.Errorf("proxyAddress = %q", params.Get("proxyAddress"))
	}
	if params.Get("proxyPort") != "8080" {
		t.Errorf("proxyPort = %q", params.Get("proxyPort"))
	}
}

func TestBuildUpdateParams_CAA(t *testing.T) {
	state := &dnsRecordResourceModel{
		Zone:  types.StringValue("example.com"),
		Domain: types.StringValue("example.com"),
		Type:  types.StringValue("CAA"),
		Value: types.StringValue("letsencrypt.org"),
		TTL:   types.Int64Value(3600),
		Flags: types.Int64Value(0),
		Tag:   types.StringValue("issue"),
	}
	plan := &dnsRecordResourceModel{
		Zone:  types.StringValue("example.com"),
		Domain: types.StringValue("example.com"),
		Type:  types.StringValue("CAA"),
		Value: types.StringValue("buypass.com"),
		TTL:   types.Int64Value(3600),
		Flags: types.Int64Value(128),
		Tag:   types.StringValue("issuewild"),
	}
	params := buildUpdateParams(state, plan)

	if params.Get("value") != "letsencrypt.org" {
		t.Errorf("value = %q", params.Get("value"))
	}
	if params.Get("newValue") != "buypass.com" {
		t.Errorf("newValue = %q", params.Get("newValue"))
	}
	if params.Get("flags") != "0" {
		t.Errorf("flags = %q", params.Get("flags"))
	}
	if params.Get("newFlags") != "128" {
		t.Errorf("newFlags = %q", params.Get("newFlags"))
	}
	if params.Get("tag") != "issue" {
		t.Errorf("tag = %q", params.Get("tag"))
	}
	if params.Get("newTag") != "issuewild" {
		t.Errorf("newTag = %q", params.Get("newTag"))
	}
}

func TestBuildUpdateParams_FWD_Advanced(t *testing.T) {
	state := &dnsRecordResourceModel{
		Zone:     types.StringValue("example.com"),
		Domain:   types.StringValue("example.com"),
		Type:     types.StringValue("FWD"),
		Value:    types.StringValue("1.1.1.1"),
		TTL:      types.Int64Value(3600),
		Protocol: types.StringValue("Udp"),
	}
	plan := &dnsRecordResourceModel{
		Zone:             types.StringValue("example.com"),
		Domain:           types.StringValue("example.com"),
		Type:             types.StringValue("FWD"),
		Value:            types.StringValue("8.8.8.8"),
		TTL:              types.Int64Value(3600),
		Protocol:         types.StringValue("Tcp"),
		DnssecValidation: types.BoolValue(true),
		ProxyType:        types.StringValue("Socks5"),
		ProxyAddress:     types.StringValue("socks.example.com"),
		ProxyPort:        types.Int64Value(1080),
	}
	params := buildUpdateParams(state, plan)

	if params.Get("forwarder") != "1.1.1.1" {
		t.Errorf("forwarder = %q", params.Get("forwarder"))
	}
	if params.Get("newForwarder") != "8.8.8.8" {
		t.Errorf("newForwarder = %q", params.Get("newForwarder"))
	}
	if params.Get("dnssecValidation") != "true" {
		t.Errorf("dnssecValidation = %q", params.Get("dnssecValidation"))
	}
	if params.Get("proxyType") != "Socks5" {
		t.Errorf("proxyType = %q", params.Get("proxyType"))
	}
	if params.Get("proxyAddress") != "socks.example.com" {
		t.Errorf("proxyAddress = %q", params.Get("proxyAddress"))
	}
	if params.Get("proxyPort") != "1080" {
		t.Errorf("proxyPort = %q", params.Get("proxyPort"))
	}
}

func TestBuildDeleteParams_CAA(t *testing.T) {
	state := &dnsRecordResourceModel{
		Zone:  types.StringValue("example.com"),
		Domain: types.StringValue("example.com"),
		Type:  types.StringValue("CAA"),
		Value: types.StringValue("letsencrypt.org"),
		Flags: types.Int64Value(0),
		Tag:   types.StringValue("issue"),
	}
	params := buildDeleteParams(state)

	if params.Get("value") != "letsencrypt.org" {
		t.Errorf("value = %q", params.Get("value"))
	}
	if params.Get("flags") != "0" {
		t.Errorf("flags = %q", params.Get("flags"))
	}
	if params.Get("tag") != "issue" {
		t.Errorf("tag = %q", params.Get("tag"))
	}
}

func TestBuildDeleteParams_FWD(t *testing.T) {
	state := &dnsRecordResourceModel{
		Zone:     types.StringValue("example.com"),
		Domain:   types.StringValue("example.com"),
		Type:     types.StringValue("FWD"),
		Value:    types.StringValue("1.1.1.1"),
		Protocol: types.StringValue("Udp"),
	}
	params := buildDeleteParams(state)

	if params.Get("forwarder") != "1.1.1.1" {
		t.Errorf("forwarder = %q", params.Get("forwarder"))
	}
	if params.Get("protocol") != "Udp" {
		t.Errorf("protocol = %q", params.Get("protocol"))
	}
}

func TestBuildUpdateParams_SRV(t *testing.T) {
	state := &dnsRecordResourceModel{
		Zone:     types.StringValue("example.com"),
		Domain:   types.StringValue("_sip._tcp.example.com"),
		Type:     types.StringValue("SRV"),
		Value:    types.StringValue("old.example.com"),
		TTL:      types.Int64Value(3600),
		Priority: types.Int64Value(10),
		Weight:   types.Int64Value(60),
		Port:     types.Int64Value(5060),
	}
	plan := &dnsRecordResourceModel{
		Zone:     types.StringValue("example.com"),
		Domain:   types.StringValue("_sip._tcp.example.com"),
		Type:     types.StringValue("SRV"),
		Value:    types.StringValue("new.example.com"),
		TTL:      types.Int64Value(1800),
		Priority: types.Int64Value(20),
		Weight:   types.Int64Value(100),
		Port:     types.Int64Value(5061),
	}
	params := buildUpdateParams(state, plan)

	if params.Get("target") != "old.example.com" {
		t.Errorf("target = %q", params.Get("target"))
	}
	if params.Get("newTarget") != "new.example.com" {
		t.Errorf("newTarget = %q", params.Get("newTarget"))
	}
	if params.Get("priority") != "10" {
		t.Errorf("priority = %q", params.Get("priority"))
	}
	if params.Get("newPriority") != "20" {
		t.Errorf("newPriority = %q", params.Get("newPriority"))
	}
	if params.Get("weight") != "60" {
		t.Errorf("weight = %q", params.Get("weight"))
	}
	if params.Get("newWeight") != "100" {
		t.Errorf("newWeight = %q", params.Get("newWeight"))
	}
	if params.Get("port") != "5060" {
		t.Errorf("port = %q", params.Get("port"))
	}
	if params.Get("newPort") != "5061" {
		t.Errorf("newPort = %q", params.Get("newPort"))
	}
}

func TestBuildUpdateParams_FWD(t *testing.T) {
	state := &dnsRecordResourceModel{
		Zone:     types.StringValue("example.com"),
		Domain:   types.StringValue("example.com"),
		Type:     types.StringValue("FWD"),
		Value:    types.StringValue("1.1.1.1"),
		TTL:      types.Int64Value(3600),
		Protocol: types.StringValue("Udp"),
	}
	plan := &dnsRecordResourceModel{
		Zone:     types.StringValue("example.com"),
		Domain:   types.StringValue("example.com"),
		Type:     types.StringValue("FWD"),
		Value:    types.StringValue("8.8.8.8"),
		TTL:      types.Int64Value(3600),
		Protocol: types.StringValue("Tcp"),
	}
	params := buildUpdateParams(state, plan)

	if params.Get("forwarder") != "1.1.1.1" {
		t.Errorf("forwarder = %q", params.Get("forwarder"))
	}
	if params.Get("newForwarder") != "8.8.8.8" {
		t.Errorf("newForwarder = %q", params.Get("newForwarder"))
	}
	if params.Get("protocol") != "Udp" {
		t.Errorf("protocol = %q", params.Get("protocol"))
	}
	if params.Get("newProtocol") != "Tcp" {
		t.Errorf("newProtocol = %q", params.Get("newProtocol"))
	}
}
